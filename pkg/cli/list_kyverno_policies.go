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

	"github.com/fairwindsops/insights-cli/pkg/kyverno"
)

func init() {
	listCmd.AddCommand(listKyvernoPoliciesCmd)
}

var listKyvernoPoliciesCmd = &cobra.Command{
	Use:    "kyverno-policies",
	Short:  "List Kyverno policies.",
	Long:   "List Kyverno policies defined in Insights.",
	PreRun: validateAndLoadInsightsAPIConfigWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		org := configurationObject.Options.Organization
		tree := treeprint.New()
		err := kyverno.AddKyvernoPoliciesBranch(client, org, tree)
		if err != nil {
			logrus.Fatalf("Unable to get Kyverno policies from insights: %v", err)
		}
		fmt.Println(tree.String())
	},
}
