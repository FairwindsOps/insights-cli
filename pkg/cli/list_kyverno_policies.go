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

var listLocal bool
var listClusterName string
var listFormat string

func init() {
	listKyvernoPoliciesCmd.Flags().BoolVar(&listLocal, "local", false, "List local policy files")
	listKyvernoPoliciesCmd.Flags().StringVar(&listClusterName, "cluster", "", "List policies for specific cluster from Insights")
	listKyvernoPoliciesCmd.Flags().StringVar(&listFormat, "format", "tree", "Output format: tree, yaml")
	listCmd.AddCommand(listKyvernoPoliciesCmd)
}

var listKyvernoPoliciesCmd = &cobra.Command{
	Use:   "kyverno-policies",
	Short: "List Kyverno policies.",
	Long:  "List Kyverno policies from local files or Insights. Use --local for local files, --cluster for cluster-specific policies.",
	Example: `
	# List all policies from Insights
	insights-cli list kyverno-policies

	# List local policy files
	insights-cli list kyverno-policies --local

	# List policies for specific cluster
	insights-cli list kyverno-policies --cluster production

	# Export cluster policies as YAML
	insights-cli list kyverno-policies --cluster production --format yaml`,
	PreRun: func(cmd *cobra.Command, args []string) {
		// Only require API config if not listing local files
		if !listLocal {
			validateAndLoadInsightsAPIConfigWrapper(cmd, args)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		if listLocal {
			// Local file system listing
			if _, err := os.Stat("kyverno-policies"); os.IsNotExist(err) {
				logrus.Fatalf("Directory kyverno-policies does not exist")
			}
			tree := treeprint.New()
			err := addLocalKyvernoPoliciesBranch("kyverno-policies", tree)
			if err != nil {
				logrus.Fatalf("Unable to list local policies: %v", err)
			}
			fmt.Println(tree.String())
		} else {
			// API listing
			org := configurationObject.Options.Organization

			if listClusterName != "" {
				// Cluster-specific listing
				if listFormat == "yaml" {
					// Export as YAML
					yamlContent, err := kyverno.ExportClusterKyvernoPoliciesYaml(client, org, listClusterName)
					if err != nil {
						logrus.Fatalf("Unable to export cluster policies: %v", err)
					}
					fmt.Print(yamlContent)
				} else {
					// List as tree (with app groups applied by default)
					tree := treeprint.New()
					err := kyverno.AddClusterKyvernoPoliciesWithAppGroupsBranch(client, org, listClusterName, tree)
					if err != nil {
						logrus.Fatalf("Unable to get cluster policies: %v", err)
					}
					fmt.Println(tree.String())
				}
			} else {
				// General API listing (existing functionality)
				tree := treeprint.New()
				err := kyverno.AddKyvernoPoliciesBranch(client, org, tree)
				if err != nil {
					logrus.Fatalf("Unable to get Kyverno policies from insights: %v", err)
				}
				fmt.Println(tree.String())
			}
		}
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
			// For simplicity, always create a new policy node for test cases
			// In a more sophisticated implementation, we'd track existing nodes
			policyNode := policiesBranch.AddBranch(policyName)

			testNode := policyNode.AddBranch("test-cases")
			testCaseNode := testNode.AddBranch(testCaseName)
			testCaseNode.AddNode(fmt.Sprintf("File: %s", filename))
			testCaseNode.AddNode(fmt.Sprintf("Expected: %s", expectedOutcome))
		}

		return nil
	})

	return err
}

// Helper functions for local policy file processing
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
