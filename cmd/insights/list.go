// Copyright 2020 FairwindsOps Inc
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

package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xlab/treeprint"

	"github.com/fairwindsops/insights-cli/pkg/insights"
)

func init() {
	policyCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List policies",
	Long:  "List OPA policies in Fairwinds Insights",
	Run: func(cmd *cobra.Command, args []string) {
		org := configurationObject.Options.Organization
		host := configurationObject.Options.Hostname
		checks, err := insights.GetChecks(org, insightsToken, host)
		if err != nil {
			panic(err)
		}
		tree := treeprint.New()
		opa := tree.AddBranch("opa")
		for _, check := range checks {
			branch := opa.AddBranch(check.Name)
			instances, err := insights.GetInstances(org, check.Name, insightsToken, host)
			if err != nil {
				panic(err)
			}
			for _, instance := range instances {
				branch.AddNode(instance.AdditionalData.Name)
			}

		}
		fmt.Println(tree.String())
	},
}
