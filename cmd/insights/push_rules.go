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
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/fairwindsops/insights-cli/pkg/rules"
)

var deleteMissingRules bool

func init() {
	pushRulesCmd.PersistentFlags().BoolVarP(&deleteMissingRules, "delete", "D", false, "Delete any automation rules from Insights that are not present in the local directory.")
	// This flag sets a variable defined in the parent `push` command.
	pushRulesCmd.PersistentFlags().StringVarP(&pushRulesSubDir, "push-rules-subdirectory", "", "rules", "Sub-directory within push-directory, to contain automation rules.")
	pushCmd.AddCommand(pushRulesCmd)
}

var pushRulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "Push automation rules.",
	Long:  "Push automation rules to Insights.",
	Run: func(cmd *cobra.Command, args []string) {
		org := configurationObject.Options.Organization
		host := configurationObject.Options.Hostname
		err := rules.PushRules(pushDir+"/"+pushRulesSubDir, org, insightsToken, host, deleteMissingRules, pushDryRun)
		if err != nil {
			logrus.Fatalf("Unable to push rules: %v", err)
		}
		logrus.Infoln("Push succeeded.")
	},
}
