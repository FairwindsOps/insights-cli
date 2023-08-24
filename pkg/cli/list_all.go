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
	"github.com/xlab/treeprint"

	"github.com/fairwindsops/insights-cli/pkg/appgroups"
	"github.com/fairwindsops/insights-cli/pkg/opa"
	"github.com/fairwindsops/insights-cli/pkg/policymappings"
	"github.com/fairwindsops/insights-cli/pkg/rules"
)

func init() {
	listCmd.AddCommand(listAllCmd)
}

var listAllCmd = &cobra.Command{
	Use:    "all",
	Short:  "List OPA policies and automation rules.",
	Long:   "List OPA policies and automation rules defined in Insights.",
	PreRun: validateAndLoadInsightsAPIConfigWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		org := configurationObject.Options.Organization
		host := configurationObject.Options.Hostname
		tree := treeprint.New()
		err := opa.AddOPAChecksBranch(org, insightsToken, host, tree)
		if err != nil {
			logrus.Fatalf("Unable to get OPA checks from insights: %v", err)
		}
		err = rules.AddRulesBranch(org, insightsToken, host, tree)
		if err != nil {
			logrus.Fatalf("Unable to get rules from insights: %v", err)
		}

		appGroups, err := appgroups.FetchAppGroups(org, insightsToken, host)
		if err != nil {
			logrus.Fatalf("unable to fetch app-groups from insights: %v", err)
		}
		err = appgroups.AddAppGroupsBranch(tree, appGroups)
		if err != nil {
			logrus.Fatalf("error building app-groups tree: %v", err)
		}

		policyMappings, err := policymappings.FetchPolicyMappings(org, insightsToken, host)
		if err != nil {
			logrus.Fatalf("unable to fetch policy-mappings from insights: %v", err)
		}
		err = policymappings.AddPolicyMappingsBranch(tree, policyMappings)
		if err != nil {
			logrus.Fatalf("error building policy-mappings tree: %v", err)
		}

		fmt.Println(tree.String())
	},
}
