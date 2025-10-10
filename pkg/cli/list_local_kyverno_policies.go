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
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/xlab/treeprint"

	"github.com/fairwindsops/insights-cli/pkg/kyverno"
)

var listLocalKyvernoPoliciesDir string

func init() {
	listLocalKyvernoPoliciesCmd.Flags().StringVarP(&listLocalKyvernoPoliciesDir, "directory", "d", "kyverno-policies", "Directory containing local Kyverno policy files")
	kyvernoPoliciesCmd.AddCommand(listLocalKyvernoPoliciesCmd)
}

var listLocalKyvernoPoliciesCmd = &cobra.Command{
	Use:   "list-local",
	Short: "List local Kyverno policy files",
	Long:  "List Kyverno policy files in the local directory structure.",
	Example: `
	# List local policy files in default directory
	insights-cli kyverno-policies list-local
	
	# List local policy files in specific directory
	insights-cli kyverno-policies list-local -d ./my-policies`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if directory exists
		if _, err := os.Stat(listLocalKyvernoPoliciesDir); os.IsNotExist(err) {
			logrus.Fatalf("Directory %s does not exist", listLocalKyvernoPoliciesDir)
		}

		tree := treeprint.New()
		err := addLocalKyvernoPoliciesBranch(listLocalKyvernoPoliciesDir, tree)
		if err != nil {
			logrus.Fatalf("Unable to list local Kyverno policies: %v", err)
		}
		fmt.Println(tree.String())
	},
}

// addLocalKyvernoPoliciesBranch builds a tree for local Kyverno policy files
func addLocalKyvernoPoliciesBranch(dir string, tree treeprint.Tree) error {
	policiesBranch := tree.AddBranch("kyverno-policies (local)")

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		filename := filepath.Base(path)

		// Only process policy files, exclude test case files
		if isPolicyFile(filename) && !isTestCaseFile(filename) {
			policyName := extractPolicyNameFromFile(filename)
			policyNode := policiesBranch.AddBranch(policyName)
			policyNode.AddNode(fmt.Sprintf("File: %s", filename))

			// Try to read and display basic policy info
			policy, err := kyverno.ReadPolicyFromFile(path)
			if err == nil {
				if policy.Kind != "" {
					policyNode.AddNode(fmt.Sprintf("Kind: %s", policy.Kind))
				}
				if policy.APIVersion != "" {
					policyNode.AddNode(fmt.Sprintf("API Version: %s", policy.APIVersion))
				}
			}
		} else if isTestCaseFile(filename) {
			// Add test case files as nodes under their policy
			policyName := extractPolicyNameFromTestCase(filename)
			testCaseName := extractTestCaseName(filename)
			expectedOutcome := determineExpectedOutcome(filename)

			// Find or create the policy node
			var policyNode treeprint.Tree
			// For simplicity, always create a new policy node for test cases
			// In a more sophisticated implementation, we'd track existing nodes
			policyNode = policiesBranch.AddBranch(policyName)

			testNode := policyNode.AddBranch("test-cases")
			testCaseNode := testNode.AddBranch(testCaseName)
			testCaseNode.AddNode(fmt.Sprintf("File: %s", filename))
			testCaseNode.AddNode(fmt.Sprintf("Expected: %s", expectedOutcome))
		}

		return nil
	})

	return err
}

// Helper functions (copied from kyverno.go for local use)
func isPolicyFile(filename string) bool {
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
	parts := strings.Split(filename, ".")
	if len(parts) >= 2 {
		return parts[0]
	}
	return ""
}

func extractTestCaseName(filename string) string {
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
