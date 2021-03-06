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

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/xlab/treeprint"

	"github.com/fairwindsops/insights-cli/pkg/opa"
	"github.com/fairwindsops/insights-cli/pkg/rules"
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
		tree := treeprint.New()
		err := opa.BuildChecksTree(org, insightsToken, host, tree)
		if err != nil {
			logrus.Fatalf("Unable to get OPA checks from insights: %v", err)
		}
		err = rules.BuildRulesTree(org, insightsToken, host, tree)
		if err != nil {
			logrus.Fatalf("Unable to get rules from insights: %v", err)
		}
		fmt.Println(tree.String())
	},
}
