// Copyright 2023 FairwindsOps Inc
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

	"github.com/fairwindsops/insights-cli/pkg/policymappings"
)

func init() {
	listCmd.AddCommand(listPolicyMappingsCmd)
}

var listPolicyMappingsCmd = &cobra.Command{
	Use:    "policy-mappings",
	Short:  "List Policy Mappings.",
	Long:   "List Policy Mappings defined in Insights.",
	PreRun: validateAndLoadInsightsAPIConfigWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		org := configurationObject.Options.Organization
		host := configurationObject.Options.Hostname
		policyMappings, err := policymappings.FetchPolicyMappings(org, insightsToken, host)
		if err != nil {
			logrus.Fatalf("unable to fetch policy-mappings from insights: %v", err)
		}
		tree, err := policymappings.BuildPolicyMappingsTree(policyMappings)
		if err != nil {
			logrus.Fatalf("error building policy-mappings tree: %v", err)
		}
		fmt.Println(tree.String())
	},
}
