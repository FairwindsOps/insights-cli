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
