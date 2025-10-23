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
var validateClusterName string

func init() {
	validateKyvernoPoliciesCmd.Flags().StringVarP(&kyvernoPolicyDir, "batch-directory", "b", "kyverno-policies", "A directory containing Kyverno policy .yaml files and corresponding test case .yaml files to validate. This option validates multiple Kyverno policies at once, and is mutually exclusive with the policy-file option.")
	validateKyvernoPoliciesCmd.Flags().StringVarP(&kyvernoPolicyFileName, "policy-file", "r", "", "A Kyverno policy file to validate. The --test-resource-file option is also required. This option validates a single policy, and is mutually exclusive with the batch-directory option.")
	validateKyvernoPoliciesCmd.Flags().StringVarP(&kyvernoTestResourceFileName, "test-resource-file", "k", "", "A Kubernetes manifest to provide as input when validating a single Kyverno policy. This option is mutually exclusive with the batch-directory option. A manifest file ending in a .success.yaml extension is expected to pass validation. A manifest file ending in a .failure.yaml extension is expected to fail validation.")
	validateKyvernoPoliciesCmd.Flags().StringSliceVarP(&validateSpecificPolicies, "policies", "p", []string{}, "Specific policy names to validate (e.g., require-labels,disallow-privileged). If not specified, all policies will be validated.")
	validateKyvernoPoliciesCmd.Flags().StringVar(&validateClusterName, "cluster", "", "Validate policies for specific cluster from Insights")
	validateCmd.AddCommand(validateKyvernoPoliciesCmd)
}

var validateKyvernoPoliciesCmd = &cobra.Command{
	Use:   "kyverno-policies {-r <policy file> -k <test resource file> | -b <directory of policies and test resources>} [flags]",
	Short: "Validate the syntax and behavior of Kyverno policies",
	Long:  `kyverno-policies runs Kyverno policies with Kubernetes manifests as input, validating policy syntax and proper validation behavior. Test case files with .success.yaml and .failure.yaml suffixes will be validated with expected outcomes.`,
	Example: `
	To validate a single policy: insights-cli validate kyverno-policies -r policy.yaml -k test-resource.yaml

	To validate a directory of policies and test resources: insights-cli validate kyverno-policies -b ./kyverno-policies

	To validate specific policies: insights-cli validate kyverno-policies -b ./kyverno-policies -p require-labels,disallow-privileged

	To validate policies for a specific cluster: insights-cli validate kyverno-policies --cluster production`,
	PreRun: validateAndLoadInsightsAPIConfigWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		if !checkValidateKyvernoPoliciesFlags() {
			err := cmd.Help()
			if err != nil {
				fmt.Println(err)
			}
			os.Exit(1)
		}

		org := configurationObject.Options.Organization

		// Handle cluster validation
		if validateClusterName != "" {
			result, err := kyverno.ValidateClusterKyvernoPolicies(client, org, validateClusterName)
			if err != nil {
				fmt.Printf("Unable to validate cluster Kyverno policies: %v", err)
			}

			// Display validation results
			kyverno.DisplayClusterValidationResults(result)

			if result.ValidationSummary.OverallStatus != "success" {
				os.Exit(1)
			}
			return
		}

		if kyvernoPolicyFileName != "" {
			// Single policy validation
			policy, err := kyverno.ReadPolicyFromFile(kyvernoPolicyFileName)
			if err != nil {
				fmt.Printf("Unable to read policy file: %v", err)
			}

			testResource, err := kyverno.ReadTestResourceFromFile(kyvernoTestResourceFileName)
			if err != nil {
				fmt.Printf("Unable to read test resource file: %v", err)
			}

			result, err := kyverno.ValidateKyvernoPolicy(
				client, org, policy, []kyverno.TestResource{testResource}, true)
			if err != nil {
				fmt.Printf("Unable to validate policy: %v", err)
			}

			displayValidationResults(result, []kyverno.TestResource{testResource})
			if !determineActualValidationResult(result, []kyverno.TestResource{testResource}) {
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
				fmt.Println("No policies to validate")
				return
			}

			// Validate each policy
			allValid := true
			for _, policyWithTestCases := range policiesToValidate {
				fmt.Printf("Validating policy: %s", policyWithTestCases.Policy.Name)

				result, err := kyverno.ValidateKyvernoPolicy(
					client, org, policyWithTestCases.Policy, policyWithTestCases.TestCases, true)
				if err != nil {
					fmt.Printf("Unable to validate policy %s: %v", policyWithTestCases.Policy.Name, err)
					allValid = false
					continue
				}

				displayValidationResults(result, policyWithTestCases.TestCases)
				if !determineActualValidationResult(result, policyWithTestCases.TestCases) {
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
		fmt.Println("Please specify one of the --policy-file or --batch-directory options to validate one or more Kyverno policies.")
		return false
	}
	if kyvernoPolicyDir != "" {
		if kyvernoPolicyFileName != "" {
			fmt.Println("Please specify only one of the --batch-directory or --policy-file option.")
			return false
		}
		if kyvernoTestResourceFileName != "" {
			fmt.Println("The --test-resource-file option is only used with the --policy-file option, to validate a single Kyverno policy.")
			return false
		}
	}
	if kyvernoPolicyFileName != "" && kyvernoTestResourceFileName == "" {
		fmt.Println("Please also specify a test resource file when validating a single Kyverno policy, using the --test-resource-file option.")
		return false
	}
	return true
}

// displayValidationResults displays validation results with visual indicators
func displayValidationResults(result *kyverno.ValidationResult, testCases []kyverno.TestResource) {
	// Determine if validation actually passed based on test case results
	actualValid := determineActualValidationResult(result, testCases)

	if actualValid {
		fmt.Printf("✅ Policy validation: PASSED\n")
	} else {
		fmt.Printf("❌ Policy validation: FAILED\n")
	}

	// Display errors if any
	if len(result.Errors) > 0 {
		fmt.Printf("❌ Errors:\n")
		for _, err := range result.Errors {
			fmt.Printf("  - %s\n", err)
		}
	}

	// Display warnings if any
	if len(result.Warnings) > 0 {
		fmt.Printf("⚠️  Warnings:\n")
		for _, warning := range result.Warnings {
			fmt.Printf("  - %s\n", warning)
		}
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

// determineActualValidationResult determines if validation actually passed based on test case results
func determineActualValidationResult(result *kyverno.ValidationResult, testCases []kyverno.TestResource) bool {
	// If there are test results from backend, use them
	if len(result.TestResults) > 0 {
		// Create a map of test case names to expected outcomes for quick lookup
		expectedOutcomes := make(map[string]string)
		for _, testCase := range testCases {
			expectedOutcomes[testCase.TestCaseName] = testCase.ExpectedOutcome
		}

		// Check each test result to see if it behaved as expected
		for _, testResult := range result.TestResults {
			expectedOutcome, exists := expectedOutcomes[testResult.TestCaseName]
			if !exists {
				// If we can't find the expected outcome, fall back to the test result's expected outcome
				expectedOutcome = testResult.ExpectedOutcome
			}

			// A test case passes if:
			// 1. Expected "success" and actual "success" (policy allows good resource)
			// 2. Expected "failure" and actual "failure" (policy rejects bad resource)
			expectedSuccess := expectedOutcome == "success"
			actualSuccess := testResult.ActualOutcome == "success"

			// Test case passes if expected outcome matches actual outcome
			if expectedSuccess != actualSuccess {
				return false
			}
		}

		return true
	}

	// Count test case types
	successTestCases := 0
	failureTestCases := 0
	for _, testCase := range testCases {
		switch testCase.ExpectedOutcome {
		case "success":
			successTestCases++
		case "failure":
			failureTestCases++
		default:
			// Do nothing
		}
	}

	// Simple logic:
	// - If we have SUCCESS test cases and NO errors → PASS (policy allows good resources)
	// - If we have FAILURE test cases and HAVE errors → PASS (policy rejects bad resources)
	// - If we have SUCCESS test cases and HAVE errors → FAIL (policy incorrectly rejects good resources)
	// - If we have FAILURE test cases and NO errors → FAIL (policy incorrectly allows bad resources)

	if successTestCases > 0 && len(result.Errors) == 0 {
		return true
	}

	if failureTestCases > 0 && len(result.Errors) > 0 {
		return true
	}

	if successTestCases > 0 && len(result.Errors) > 0 {
		return false
	}

	if failureTestCases > 0 && len(result.Errors) == 0 {
		return false
	}

	// Fall back to backend's determination
	return result.Valid
}
