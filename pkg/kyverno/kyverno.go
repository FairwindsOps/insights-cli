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

package kyverno

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/xlab/treeprint"

	"github.com/imroc/req/v3"
	"gopkg.in/yaml.v3"
)

// AddKyvernoPoliciesBranch builds a tree for Kyverno policies
func AddKyvernoPoliciesBranch(client *req.Client, org string, tree treeprint.Tree) error {
	policies, err := FetchKyvernoPolicies(client, org)
	if err != nil {
		logrus.Errorf("Unable to get Kyverno policies from insights: %v", err)
		return err
	}
	policiesBranch := tree.AddBranch("kyverno-policies")
	for _, policy := range policies {
		policyNode := policiesBranch.AddBranch(policy.Name)
		if policy.Kind != "" {
			value := fmt.Sprintf("Kind: %s", policy.Kind)
			policyNode.AddNode(value)
		}
		if policy.APIVersion != "" {
			value := fmt.Sprintf("API Version: %s", policy.APIVersion)
			policyNode.AddNode(value)
		}
	}
	return nil
}

// AddClusterKyvernoPoliciesBranch builds a tree for cluster-specific Kyverno policies
func AddClusterKyvernoPoliciesBranch(client *req.Client, org, cluster string, tree treeprint.Tree) error {
	policies, err := FetchClusterKyvernoPolicies(client, org, cluster)
	if err != nil {
		logrus.Errorf("Unable to get cluster Kyverno policies from insights: %v", err)
		return err
	}
	policiesBranch := tree.AddBranch(fmt.Sprintf("kyverno-policies (cluster: %s)", cluster))
	for _, policy := range policies {
		policyNode := policiesBranch.AddBranch(policy.Name)
		if policy.Kind != "" {
			value := fmt.Sprintf("Kind: %s", policy.Kind)
			policyNode.AddNode(value)
		}
		if policy.APIVersion != "" {
			value := fmt.Sprintf("API Version: %s", policy.APIVersion)
			policyNode.AddNode(value)
		}
	}
	return nil
}

// AddClusterKyvernoPoliciesWithAppGroupsBranch builds a tree for cluster-specific Kyverno policies with app groups applied
func AddClusterKyvernoPoliciesWithAppGroupsBranch(client *req.Client, org, cluster string, tree treeprint.Tree) error {
	policies, err := FetchClusterKyvernoPoliciesWithAppGroups(client, org, cluster)
	if err != nil {
		logrus.Errorf("Unable to get cluster Kyverno policies with app groups from insights: %v", err)
		return err
	}
	policiesBranch := tree.AddBranch(fmt.Sprintf("kyverno-policies (cluster: %s, with app groups)", cluster))
	for _, policy := range policies {
		policyNode := policiesBranch.AddBranch(policy.Name)
		if policy.Kind != "" {
			value := fmt.Sprintf("Kind: %s", policy.Kind)
			policyNode.AddNode(value)
		}
		if policy.APIVersion != "" {
			value := fmt.Sprintf("API Version: %s", policy.APIVersion)
			policyNode.AddNode(value)
		}
	}
	return nil
}

// PushKyvernoPolicies pushes Kyverno policies to insights using bulk API
func PushKyvernoPolicies(client *req.Client, policies []KyvernoPolicy, org string, deleteMissing, dryRun bool) error {
	logrus.Debugln("Pushing Kyverno policies")

	if dryRun {
		logrus.Infof("Dry run: Would push %d Kyverno policies", len(policies))
		for _, policy := range policies {
			logrus.Infof("  - %s (%s)", policy.Name, policy.Kind)
		}
		if deleteMissing {
			logrus.Info("Dry run: Would delete missing policies")
		}
		return nil
	}

	// Use bulk API for efficient operations
	result, err := BulkUpsertKyvernoPolicies(client, org, policies, deleteMissing)
	if err != nil {
		logrus.Errorf("Unable to bulk upsert Kyverno policies: %v", err)
		return err
	}

	logrus.Infof("Bulk upsert completed: %d created, %d updated, %d deleted",
		result.Created, result.Updated, result.Deleted)

	if len(result.Errors) > 0 {
		logrus.Warnf("Bulk upsert had %d errors:", len(result.Errors))
		for _, errMsg := range result.Errors {
			logrus.Warnf("  - %s", errMsg)
		}
	}

	return nil
}

// GetPolicyFilesForPush gets only policy files (excluding test cases) for push operations
func GetPolicyFilesForPush(policyDir string) ([]KyvernoPolicy, error) {
	var policies []KyvernoPolicy

	err := filepath.Walk(policyDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking directory %s: %w", path, err)
		}

		if info.IsDir() {
			return nil
		}

		filename := filepath.Base(path)

		// Only process policy files, exclude test case files
		if isPolicyFile(filename) && !isTestCaseFile(filename) {
			policy, err := readPolicyFromFile(path)
			if err != nil {
				return fmt.Errorf("error reading policy file %s: %w", filename, err)
			}
			policies = append(policies, policy)
		}

		return nil
	})

	return policies, err
}

// DiscoverPoliciesAndTestCases discovers all policies and their associated test cases
func DiscoverPoliciesAndTestCases(policyDir string) ([]PolicyWithTestCases, error) {
	var policies []PolicyWithTestCases
	policyMap := make(map[string]*PolicyWithTestCases)

	err := filepath.Walk(policyDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking directory %s: %w", path, err)
		}

		if info.IsDir() {
			return nil
		}

		filename := filepath.Base(path)

		if isTestCaseFile(filename) {
			// This is a test case file
			policyName := extractPolicyNameFromTestCase(filename)
			testCase := TestResource{
				Content:         readFileContent(path),
				FileName:        filename,
				PolicyName:      policyName,
				TestCaseName:    extractTestCaseName(filename),
				ExpectedOutcome: determineExpectedOutcome(filename),
			}

			if policyMap[policyName] == nil {
				policyMap[policyName] = &PolicyWithTestCases{
					TestCases: []TestResource{},
				}
			}
			policyMap[policyName].TestCases = append(policyMap[policyName].TestCases, testCase)

		} else if isPolicyFile(filename) {
			// This is a policy file
			policyName := extractPolicyNameFromFile(filename)
			policy, err := readPolicyFromFile(path)
			if err != nil {
				return fmt.Errorf("error reading policy file %s: %w", filename, err)
			}

			if policyMap[policyName] == nil {
				policyMap[policyName] = &PolicyWithTestCases{
					TestCases: []TestResource{},
				}
			}
			policyMap[policyName].Policy = policy
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Convert map to slice
	for _, policyWithTestCases := range policyMap {
		policies = append(policies, *policyWithTestCases)
	}

	return policies, nil
}

// Helper functions

// validatePath checks for path traversal attacks
func validatePath(path string) error {
	if strings.Contains(path, "..") {
		return fmt.Errorf("path traversal not allowed: %s", path)
	}
	return nil
}

func isPolicyFile(filename string) bool {
	// Policy file: ends with .yaml/.yml but doesn't contain .testcase
	return (strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml")) &&
		!strings.Contains(filename, ".testcase")
}

func isTestCaseFile(filename string) bool {
	return strings.Contains(filename, ".testcase")
}

func extractPolicyNameFromFile(filename string) string {
	name := strings.TrimSuffix(filename, ".yaml")
	name = strings.TrimSuffix(name, ".yml")
	return name
}

func extractPolicyNameFromTestCase(filename string) string {
	// Extract policy name from test case filename
	// e.g., "require-labels.testcase1.success.yaml" -> "require-labels"
	parts := strings.Split(filename, ".")
	if len(parts) >= 2 {
		return parts[0]
	}
	return ""
}

func extractTestCaseName(filename string) string {
	// Extract test case name from filename
	// e.g., "require-labels.testcase1.success.yaml" -> "testcase1"
	parts := strings.Split(filename, ".")
	for _, part := range parts {
		if strings.HasPrefix(part, "testcase") {
			return part
		}
	}
	return ""
}

func determineExpectedOutcome(filename string) string {
	if strings.Contains(filename, ".success.") {
		return "success"
	}
	if strings.Contains(filename, ".failure.") {
		return "failure"
	}
	return "unknown"
}

func readPolicyFromFile(filePath string) (KyvernoPolicy, error) {
	if err := validatePath(filePath); err != nil {
		return KyvernoPolicy{}, err
	}

	fileContents, err := os.ReadFile(filePath)
	if err != nil {
		return KyvernoPolicy{}, fmt.Errorf("failed to read policy file %s: %w", filePath, err)
	}

	var policy KyvernoPolicy
	err = yaml.Unmarshal(fileContents, &policy)
	if err != nil {
		return KyvernoPolicy{}, fmt.Errorf("failed to parse YAML in policy file %s: %w", filePath, err)
	}

	// Extract name from metadata if not set
	if policy.Name == "" {
		// Parse YAML to extract metadata.name
		var yamlData map[string]interface{}
		err = yaml.Unmarshal(fileContents, &yamlData)
		if err != nil {
			return KyvernoPolicy{}, fmt.Errorf("failed to parse YAML metadata in policy file %s: %w", filePath, err)
		}

		if metadata, ok := yamlData["metadata"].(map[string]interface{}); ok {
			if name, ok := metadata["name"].(string); ok {
				policy.Name = name
			}
		}
	}

	return policy, nil
}

func readFileContent(filePath string) string {
	if err := validatePath(filePath); err != nil {
		logrus.Warnf("Invalid file path %s: %v", filePath, err)
		return ""
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		logrus.Warnf("Failed to read file %s: %v", filePath, err)
		return ""
	}
	return string(content)
}

// ReadPolicyFromFile reads a Kyverno policy from a file
func ReadPolicyFromFile(filePath string) (KyvernoPolicy, error) {
	return readPolicyFromFile(filePath)
}

// ReadTestResourceFromFile reads a test resource from a file
func ReadTestResourceFromFile(filePath string) (TestResource, error) {
	if err := validatePath(filePath); err != nil {
		return TestResource{}, err
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return TestResource{}, fmt.Errorf("failed to read test resource file %s: %w", filePath, err)
	}

	filename := filepath.Base(filePath)
	return TestResource{
		Content:         string(content),
		FileName:        filename,
		PolicyName:      "", // Will be set by caller
		TestCaseName:    extractTestCaseName(filename),
		ExpectedOutcome: determineExpectedOutcome(filename),
	}, nil
}

// DisplayClusterValidationResults displays cluster validation results in a user-friendly format
func DisplayClusterValidationResults(result *ClusterValidationResponse) {
	fmt.Printf("🔍 Cluster Validation Results for: %s\n", result.Cluster)
	fmt.Printf("📊 Summary: %d total policies, %d valid, %d invalid\n",
		result.TotalPolicies, result.ValidPolicies, result.InvalidPolicies)

	// Display overall status
	switch result.ValidationSummary.OverallStatus {
	case "success":
		fmt.Printf("✅ Overall Status: SUCCESS\n")
	case "partial_success":
		fmt.Printf("⚠️  Overall Status: PARTIAL SUCCESS\n")
	case "failure":
		fmt.Printf("❌ Overall Status: FAILURE\n")
	default:
		fmt.Printf("❓ Overall Status: %s\n", result.ValidationSummary.OverallStatus)
	}

	fmt.Printf("📈 Errors: %d, Warnings: %d\n\n",
		result.ValidationSummary.TotalErrors, result.ValidationSummary.TotalWarnings)

	// Display individual policy results
	if len(result.PolicyResults) > 0 {
		fmt.Println("📋 Policy Results:")
		for _, policyResult := range result.PolicyResults {
			var statusIcon string
			switch policyResult.Status {
			case "invalid":
				statusIcon = "❌"
			case "error":
				statusIcon = "🚫"
			default:
				statusIcon = "✅"
			}

			fmt.Printf("  %s %s (%s)\n", statusIcon, policyResult.PolicyName, policyResult.Status)

			// Show app groups applied if any
			if len(policyResult.AppGroupsApplied) > 0 {
				fmt.Printf("    📦 App Groups: %s\n", strings.Join(policyResult.AppGroupsApplied, ", "))
			}

			// Show validation details
			if !policyResult.ValidationResult.Valid {
				if len(policyResult.ValidationResult.Errors) > 0 {
					fmt.Printf("    ❌ Errors:\n")
					for _, err := range policyResult.ValidationResult.Errors {
						fmt.Printf("      - %s\n", err)
					}
				}
			}

			if len(policyResult.ValidationResult.Warnings) > 0 {
				fmt.Printf("    ⚠️  Warnings:\n")
				for _, warning := range policyResult.ValidationResult.Warnings {
					fmt.Printf("      - %s\n", warning)
				}
			}

			fmt.Println()
		}
	}
}
