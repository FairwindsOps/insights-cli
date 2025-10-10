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

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/xlab/treeprint"

	"github.com/fairwindsops/insights-cli/pkg/kyverno"
)

func init() {
	kyvernoPoliciesClusterCmd.AddCommand(listClusterKyvernoPoliciesCmd)
}

var listClusterKyvernoPoliciesCmd = &cobra.Command{
	Use:   "list <cluster>",
	Short: "List Kyverno policies for a specific cluster",
	Long:  "List Kyverno policies that apply to the specified cluster based on app groups and policy mappings.",
	Example: `
	# List policies for a specific cluster
	insights-cli kyverno-policies cluster list my-cluster
	
	# List with pagination
	insights-cli kyverno-policies cluster list my-cluster --page 1 --page-size 20`,
	Args:   cobra.ExactArgs(1),
	PreRun: validateAndLoadInsightsAPIConfigWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		cluster := args[0]
		org := configurationObject.Options.Organization

		tree := treeprint.New()
		err := kyverno.AddClusterKyvernoPoliciesBranch(client, org, cluster, tree)
		if err != nil {
			logrus.Fatalf("Unable to get cluster Kyverno policies: %v", err)
		}
		fmt.Println(tree.String())
	},
}
