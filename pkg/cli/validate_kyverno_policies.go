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
	validateKyvernoPoliciesCmd.Flags().StringVarP(&kyvernoPolicyDir, "batch-directory", "b", "", "A directory containing Kyverno policy .yaml files and corresponding test case .yaml files to validate. This option validates multiple Kyverno policies at once, and is mutually exclusive with the policy-file option.")
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
				fmt.Println("❌ Kyverno policy validation failed.")
				os.Exit(1)
			}
			fmt.Println("✅ Kyverno policy validated successfully.")
			return
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
				fmt.Println("❌ No policies to validate")
				os.Exit(1)
			}

			// Validate each policy
			allValid := true
			for _, policyWithTestCases := range policiesToValidate {
				fmt.Println("\n--------------------------------")
				fmt.Printf("🔍 Validating policy: %s\n", policyWithTestCases.Policy.Name)
				result, err := kyverno.ValidateKyvernoPolicy(
					client, org, policyWithTestCases.Policy, policyWithTestCases.TestCases, true)
				if err != nil {
					allValid = false
					fmt.Printf("❌ Unable to validate policy %s: %v\n", policyWithTestCases.Policy.Name, err)
					continue
				}

				displayValidationResults(result, policyWithTestCases.TestCases)
				if !determineActualValidationResult(result, policyWithTestCases.TestCases) {
					allValid = false
				}
			}

			if !allValid {
				fmt.Println("\n--------------------------------")
				fmt.Println("❌ Some Kyverno policies validation failed. Please check the output for details.")
				os.Exit(1)
			}
			fmt.Println("✅ All Kyverno policies validated successfully!")
			return
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

	// Display test case results with clear messaging
	if len(result.TestResults) > 0 {
		fmt.Printf("\n📋 Test case results:\n")
		for _, testResult := range result.TestResults {
			passed := testResult.Passed

			// Display the result with test case name prominently
			if passed {
				fmt.Printf("  ✅ %s (%s)\n", testResult.TestCaseName, testResult.FileName)
			} else {
				fmt.Printf("  ❌ %s (%s)\n", testResult.TestCaseName, testResult.FileName)
			}
		}
	}

	// If TestResults are empty but we have test cases, show them with fallback info
	if len(result.TestResults) == 0 && len(testCases) > 0 {
		fmt.Printf("\n📋 Test cases:\n")

		// Check if we have mixed test cases (both success and failure)
		successCount := 0
		failureCount := 0
		for _, tc := range testCases {
			switch tc.ExpectedOutcome {
			case "success":
				successCount++
			case "failure":
				failureCount++
			default:
				continue
			}
		}
		hasMixedCases := successCount > 0 && failureCount > 0

		for _, testCase := range testCases {
			var passed bool
			if hasMixedCases {
				passed = actualValid
			} else {
				switch testCase.ExpectedOutcome {
				case "success":
					passed = len(result.Errors) == 0
				case "failure":
					passed = len(result.Errors) > 0
				default:
					passed = actualValid
				}
			}

			if passed {
				fmt.Printf("  ✅ %s (%s)\n", testCase.TestCaseName, testCase.FileName)
			} else {
				fmt.Printf("  ❌ %s (%s)\n", testCase.TestCaseName, testCase.FileName)
			}
		}
	}

	if len(result.Errors) > 0 {
		fmt.Printf("\n⚠️ Output:\n")
		for _, err := range result.Errors {
			fmt.Printf("  - %s\n", err)
		}
	}

	if len(result.Warnings) > 0 {
		fmt.Printf("\n⚠️ Output:\n")
		for _, warning := range result.Warnings {
			fmt.Printf("  - %s\n", warning)
		}
	}
}

// matchTestResultsToTestCases matches TestResults to TestCases using FileName and TestCaseName
// Returns a map of test case index to its TestResult (if found)
// Uses FileName:TestCaseName as the unique identifier to ensure correct matching
// Handles both patterns:
//   - "policy.testcase1.success.yaml" -> TestCaseName: "testcase1.success", FileName: "policy.testcase1.success.yaml"
//   - "policy.success.yaml" -> TestCaseName: "success", FileName: "policy.success.yaml"
func matchTestResultsToTestCases(result *kyverno.ValidationResult, testCases []kyverno.TestResource) map[int]*kyverno.TestResult {
	resultMap := make(map[int]*kyverno.TestResult)

	// Create a map of test cases by FileName and TestCaseName for quick lookup
	testCaseMap := make(map[string]int)
	for i, testCase := range testCases {
		// Use FileName and TestCaseName as the unique identifier
		// This handles both patterns:
		// - "policy.testcase1.success.yaml" with TestCaseName "testcase1.success"
		// - "policy.success.yaml" with TestCaseName "success"
		key := fmt.Sprintf("%s:%s", testCase.FileName, testCase.TestCaseName)
		testCaseMap[key] = i
	}

	// Match TestResults to test cases using FileName and TestCaseName
	for _, testResult := range result.TestResults {
		if testResult.FileName != "" && testResult.TestCaseName != "" {
			key := fmt.Sprintf("%s:%s", testResult.FileName, testResult.TestCaseName)
			if idx, exists := testCaseMap[key]; exists {
				resultMap[idx] = &testResult
			}
		}
	}

	return resultMap
}

// determineActualValidationResult determines if validation actually passed based on test case results
// A test case passes if:
//   - .success.yaml file: Expected "success" and actual "success" (policy allows good resource) → PASS
//   - .failure.yaml file: Expected "failure" and actual "failure" (policy rejects bad resource) → PASS
//
// A test case fails if:
//   - .success.yaml file: Expected "success" but actual "failure" (policy incorrectly rejects good resource) → FAIL
//   - .failure.yaml file: Expected "failure" but actual "success" (policy incorrectly allows bad resource) → FAIL
func determineActualValidationResult(result *kyverno.ValidationResult, testCases []kyverno.TestResource) bool {
	// If we have TestResults, use them to match test cases by FileName and TestCaseName
	if len(result.TestResults) > 0 {
		resultMap := matchTestResultsToTestCases(result, testCases)

		// Check each test case individually
		allPassed := true
		for i, testCase := range testCases {
			testResult, hasResult := resultMap[i]

			if hasResult {
				// We have a TestResult for this test case - use it directly
				// A test case passes if expected outcome matches actual outcome
				expectedSuccess := testCase.ExpectedOutcome == "success"
				actualSuccess := testResult.ActualOutcome == "success"

				// Test case fails if expected outcome doesn't match actual outcome
				if expectedSuccess != actualSuccess {
					allPassed = false
				}

				// Also check the Passed field if available
				if !testResult.Passed {
					allPassed = false
				}
			} else {
				// No TestResult for this test case - we can't determine if it passed
				// This should not happen if the backend is working correctly
				// Fail validation if we can't find a result for a test case
				allPassed = false
			}
		}

		return allPassed
	}

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

	// Simple fallback logic:
	// - If we have ONLY SUCCESS test cases and NO errors → PASS
	// - If we have ONLY FAILURE test cases and HAVE errors → PASS
	// - If we have ONLY SUCCESS test cases and HAVE errors → FAIL
	// - If we have ONLY FAILURE test cases and NO errors → FAIL
	// - If we have MIXED test cases → Cannot reliably determine without TestResults

	if successTestCases > 0 && failureTestCases > 0 {
		// Mixed test cases: we can't reliably determine which errors belong to which test case
		// Without TestResults, we cannot properly validate mixed cases
		// If we have errors, assume failure test cases failed (expected) and success test cases passed (expected)
		if len(result.Errors) > 0 {
			return true
		} else {
			// No errors means failure test cases passed when they should have failed
			return false
		}
	}

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
