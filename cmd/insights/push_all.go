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
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/fairwindsops/insights-cli/pkg/opa"
	"github.com/fairwindsops/insights-cli/pkg/policies"
	"github.com/fairwindsops/insights-cli/pkg/rules"
)

func init() {
	// This flag sets a variable defined in the parent `push` command.
	pushAllCmd.PersistentFlags().StringVarP(&pushOPASubDir, "push-opa-subdirectory", "", defaultPushOPASubDir, "Sub-directory within push-directory, to contain OPA policies.")
	// This flag sets a variable defined in the parent `push` command.
	pushAllCmd.PersistentFlags().StringVarP(&pushRulesSubDir, "push-rules-subdirectory", "", defaultPushRulesSubDir, "Sub-directory within push-directory, to contain automation rules.")
	pushCmd.AddCommand(pushAllCmd)
}

var pushAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Push OPA policies, automation rules, and policies configuration.",
	Long:  "Push OPA policies, automation rules, and policies configuration to Insights.",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := os.Stat(pushDir)
		if err != nil {
			logrus.Fatalf("Unable to push  to Insights: %v", err)
		}
		org := configurationObject.Options.Organization
		host := configurationObject.Options.Hostname
		const (
			doNotDeleteMissingResources bool = false
			numExpectedSuccesses             = 3
		)
		var numFailures int
		logrus.Infoln("Pushing OPA policies, automation rules, and policies configuration to Insights.")
		err = opa.PushOPAChecks(pushDir+"/"+pushOPASubDir, org, insightsToken, host, doNotDeleteMissingResources, pushDryRun)
		if err != nil {
			logrus.Errorf("Unable to push OPA Checks: %v", err)
			numFailures++
		}
		err = rules.PushRules(pushDir+"/"+pushRulesSubDir, org, insightsToken, host, doNotDeleteMissingResources, pushDryRun)
		if err != nil {
			logrus.Errorf("Unable to push rules: %v", err)
			numFailures++
		}
		err = policies.PushPolicies(pushDir, org, insightsToken, host, doNotDeleteMissingResources)
		if err != nil {
			logrus.Errorf("Unable to push policies configuration: %v", err)
			numFailures++
		}
		if numFailures > 0 && numFailures < numExpectedSuccesses {
			logrus.Fatalln("Push partially failed.")
			return
		}
		if numFailures == numExpectedSuccesses {
			logrus.Fatalln("Push failed.")
			return
		}
		logrus.Infoln("Push succeeded")
	},
}
