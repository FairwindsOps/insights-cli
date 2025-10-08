// Copyright 2025 FairwindsOps Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cli

import (
	"fmt"
	"os"

	"github.com/fairwindsops/insights-cli/pkg/kyverno"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var kyvernoPolicyDir string
var validateSpecificPolicies []string
var kyvernoPolicyFileName string
var kyvernoTestResourceFileName string

func init() {
	validateKyvernoPoliciesCmd.Flags().StringVarP(&kyvernoPolicyDir, "batch-directory", "b", "", "A directory containing Kyverno policy .yaml files and corresponding test case .yaml files to validate. This option validates multiple Kyverno policies at once, and is mutually exclusive with the policy-file option.")
	validateKyvernoPoliciesCmd.Flags().StringVarP(&kyvernoPolicyFileName, "policy-file", "r", "", "A Kyverno policy file to validate. The --test-resource-file option is also required. This option validates a single policy, and is mutually exclusive with the batch-directory option.")
	validateKyvernoPoliciesCmd.Flags().StringVarP(&kyvernoTestResourceFileName, "test-resource-file", "k", "", "A Kubernetes manifest to provide as input when validating a single Kyverno policy. This option is mutually exclusive with the batch-directory option. A manifest file ending in a .success.yaml extension is expected to pass validation. A manifest file ending in a .failure.yaml extension is expected to fail validation.")
	validateKyvernoPoliciesCmd.Flags().StringSliceVarP(&validateSpecificPolicies, "policies", "p", []string{}, "Specific policy names to validate (e.g., require-labels,disallow-privileged). If not specified, all policies will be validated.")
	validateCmd.AddCommand(validateKyvernoPoliciesCmd)
}

var validateKyvernoPoliciesCmd = &cobra.Command{
	Use:   "kyverno-policies {-r <policy file> -k <test resource file> | -b <directory of policies and test resources>} [flags]",
	Short: "Validate the syntax and behavior of Kyverno policies",
	Long:  `kyverno-policies runs Kyverno policies with Kubernetes manifests as input, validating policy syntax and proper validation behavior. Test case files with .success.yaml and .failure.yaml suffixes will be validated with expected outcomes.`,
	Example: `
	To validate a single policy: insights-cli validate kyverno-policies -r policy.yaml -k test-resource.yaml

	To validate a directory of policies and test resources: insights-cli validate kyverno-policies -b ./kyverno-policies

	To validate specific policies: insights-cli validate kyverno-policies -b ./kyverno-policies -p require-labels,disallow-privileged`,
	PreRun: validateAndLoadInsightsAPIConfigWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		if !checkValidateKyvernoPoliciesFlags() {
			err := cmd.Help()
			if err != nil {
				logrus.Error(err)
			}
			os.Exit(1)
		}

		org := configurationObject.Options.Organization

		if kyvernoPolicyFileName != "" {
			// Single policy validation
			policy, err := kyverno.ReadPolicyFromFile(kyvernoPolicyFileName)
			if err != nil {
				logrus.Fatalf("Unable to read policy file: %v", err)
			}

			testResource, err := kyverno.ReadTestResourceFromFile(kyvernoTestResourceFileName)
			if err != nil {
				logrus.Fatalf("Unable to read test resource file: %v", err)
			}

			result, err := kyverno.ValidateKyvernoPolicyWithExpectedOutcomes(
				client, org, policy, []kyverno.TestResource{testResource}, true)
			if err != nil {
				logrus.Fatalf("Unable to validate policy: %v", err)
			}

			displayValidationResultsWithExpectedOutcomes(result)
			if !result.Valid {
				os.Exit(1)
			}
			fmt.Println("Kyverno policy validated successfully.")
		}

		if kyvernoPolicyDir != "" {
			// Batch validation
			policiesWithTestCases, err := kyverno.DiscoverPoliciesAndTestCases(kyvernoPolicyDir)
			if err != nil {
				logrus.Fatalf("Unable to discover policies: %v", err)
			}

			// Filter policies if specific ones are requested
			var policiesToValidate []kyverno.PolicyWithTestCases
			if len(validateSpecificPolicies) > 0 {
				// Validate specific policies
				for _, requestedPolicy := range validateSpecificPolicies {
					for _, policyWithTestCases := range policiesWithTestCases {
						if policyWithTestCases.Policy.Name == requestedPolicy {
							policiesToValidate = append(policiesToValidate, policyWithTestCases)
							break
						}
					}
				}
			} else {
				// Validate all policies
				policiesToValidate = policiesWithTestCases
			}

			if len(policiesToValidate) == 0 {
				logrus.Info("No policies to validate")
				return
			}

			// Validate each policy
			allValid := true
			for _, policyWithTestCases := range policiesToValidate {
				logrus.Infof("Validating policy: %s", policyWithTestCases.Policy.Name)

				result, err := kyverno.ValidateKyvernoPolicyWithExpectedOutcomes(
					client, org, policyWithTestCases.Policy, policyWithTestCases.TestCases, true)
				if err != nil {
					logrus.Errorf("Unable to validate policy %s: %v", policyWithTestCases.Policy.Name, err)
					allValid = false
					continue
				}

				displayValidationResultsWithExpectedOutcomes(result)
				if !result.Valid {
					allValid = false
				}
			}

			if !allValid {
				os.Exit(1)
			}

			fmt.Println("All Kyverno policies validated successfully!")
		}
	},
}

// checkValidateKyvernoPoliciesFlags verifies supplied flags for `validate kyverno-policies` are valid.
func checkValidateKyvernoPoliciesFlags() bool {
	if kyvernoPolicyDir == "" && kyvernoPolicyFileName == "" {
		logrus.Errorln("Please specify one of the --policy-file or --batch-directory options to validate one or more Kyverno policies.")
		return false
	}
	if kyvernoPolicyDir != "" {
		if kyvernoPolicyFileName != "" {
			logrus.Errorln("Please specify only one of the --batch-directory or --policy-file option.")
			return false
		}
		if kyvernoTestResourceFileName != "" {
			logrus.Errorln("The --test-resource-file option is only used with the --policy-file option, to validate a single Kyverno policy.")
			return false
		}
	}
	if kyvernoPolicyFileName != "" && kyvernoTestResourceFileName == "" {
		logrus.Errorln("Please also specify a test resource file when validating a single Kyverno policy, using the --test-resource-file option.")
		return false
	}
	return true
}

// displayValidationResultsWithExpectedOutcomes displays validation results with visual indicators
func displayValidationResultsWithExpectedOutcomes(result *kyverno.ValidationResultWithExpectedOutcomes) {
	if result.Valid {
		fmt.Printf("✅ Policy validation: PASSED\n")
	} else {
		fmt.Printf("❌ Policy validation: FAILED\n")
	}

	// Display test case results
	for _, testResult := range result.TestResults {
		if testResult.Passed {
			fmt.Printf("✓ Test case %s (%s): PASSED - Expected %s, got %s\n",
				testResult.TestCaseName, testResult.FileName,
				testResult.ExpectedOutcome, testResult.ActualOutcome)
		} else {
			fmt.Printf("❌ Test case %s (%s): FAILED - Expected %s, got %s\n",
				testResult.TestCaseName, testResult.FileName,
				testResult.ExpectedOutcome, testResult.ActualOutcome)
		}
	}
}
