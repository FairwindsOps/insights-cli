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
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/fairwindsops/insights-cli/pkg/appgroups"
	"github.com/fairwindsops/insights-cli/pkg/opa"
	"github.com/fairwindsops/insights-cli/pkg/policies"
	"github.com/fairwindsops/insights-cli/pkg/policymappings"
	"github.com/fairwindsops/insights-cli/pkg/rules"
)

var warningsAreFatal bool

func init() {
	pushAllCmd.PersistentFlags().StringVarP(&pushOPASubDir, "push-opa-subdirectory", "", defaultPushOPASubDir, "Sub-directory within push-directory, to contain OPA policies.")
	pushAllCmd.PersistentFlags().StringVarP(&pushRulesSubDir, "push-rules-subdirectory", "", defaultPushRulesSubDir, "Sub-directory within push-directory, to contain automation rules.")
	pushAllCmd.PersistentFlags().StringVarP(&pushAppGroupsSubDir, "push-app-groups-subdirectory", "", defaultPushAppGroupsSubDir, "Sub-directory within push-directory, to contain App Groups.")
	pushAllCmd.PersistentFlags().StringVarP(&pushPolicyMappingsSubDir, "push-policy-mappings-subdirectory", "", defaultPushPolicyMappingsSubDir, "Sub-directory within push-directory, to contain Policy Mappings.")
	pushAllCmd.PersistentFlags().BoolVarP(&warningsAreFatal, "warnings-are-fatal", "", false, "Treat warnings as a failure and exit with a non-zero status. For example, if pushing OPA policies and automation rules succeeds, but pushing policies configuration fails because the settings.yaml file is not present.")
	pushCmd.AddCommand(pushAllCmd)
}

var pushAllCmd = &cobra.Command{
	Use:    "all",
	Short:  "Push OPA policies, automation rules, app-groups, policy mappings and policies configuration.",
	Long:   "Push OPA policies, automation rules, app-groups, policy mappings and policies configuration to Insights.",
	PreRun: validateAndLoadInsightsAPIConfigWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		_, err := os.Stat(pushDir)
		if err != nil {
			logrus.Fatalf("Unable to push  to Insights: %v", err)
		}
		org := configurationObject.Options.Organization
		host := configurationObject.Options.Hostname
		const (
			doNotDeleteMissingResources bool = false
			numExpectedSuccesses             = 5
		)
		var numWarnings, numFailures int
		logrus.Infoln("Pushing OPA policies, automation rules, and policies configuration to Insights.")
		absPushOPADir := filepath.Join(pushDir, pushOPASubDir)
		_, err = os.Stat(absPushOPADir)
		if err != nil {
			logrus.Warnf("Unable to push OPA policies: %v", err)
			numWarnings++
		} else {
			err = opa.PushOPAChecks(absPushOPADir, org, insightsToken, host, doNotDeleteMissingResources, pushDryRun)
			if err != nil {
				logrus.Errorf("Unable to push OPA policies: %v", err)
				numFailures++
			}
		}
		absPushRulesDir := filepath.Join(pushDir, pushRulesSubDir)
		_, err = os.Stat(absPushRulesDir)
		if err != nil {
			logrus.Warnf("Unable to push automation rules: %v", err)
			numWarnings++
		} else {
			err = rules.PushRules(absPushRulesDir, org, insightsToken, host, doNotDeleteMissingResources, pushDryRun)
			if err != nil {
				logrus.Errorf("Unable to push automation rules: %v", err)
				numFailures++
			}
		}
		absPushPoliciesConfigFile := filepath.Join(pushDir, "settings.yaml")
		_, err = os.Stat(absPushPoliciesConfigFile)
		if err != nil {
			logrus.Warnf("Unable to push policies configuration: %v", err)
			numWarnings++
		} else {
			err = policies.PushPolicies(pushDir, org, insightsToken, host, pushDryRun)
			if err != nil {
				logrus.Errorf("Unable to push policies configuration: %v", err)
				numFailures++
			}
		}

		absPushAppGroupsDir := filepath.Join(pushDir, pushAppGroupsSubDir)
		_, err = os.Stat(absPushAppGroupsDir)
		if err != nil {
			logrus.Warnf("Unable to push app-groups: %v", err)
			numWarnings++
		} else {
			err = appgroups.PushAppGroups(absPushAppGroupsDir, org, insightsToken, host, doNotDeleteMissingResources, pushDryRun)
			if err != nil {
				logrus.Errorf("Unable to push app-groups: %v", err)
				numFailures++
			}
		}

		absPushPolicyMappingsDir := filepath.Join(pushDir, pushPolicyMappingsSubDir)
		_, err = os.Stat(absPushPolicyMappingsDir)
		if err != nil {
			logrus.Warnf("Unable to push policy-mappings: %v", err)
			numWarnings++
		} else {
			err = policymappings.PushPolicyMappings(absPushPolicyMappingsDir, org, insightsToken, host, doNotDeleteMissingResources, pushDryRun)
			if err != nil {
				logrus.Errorf("Unable to push policy-mappings: %v", err)
				numFailures++
			}
		}

		if numFailures == 0 && numWarnings == 0 {
			logrus.Infoln("Push succeeded")
			return
		}
		if !warningsAreFatal && numFailures == 0 && numWarnings > 0 {
			logrus.Warnf("Push failed with %d warning(s)", numWarnings)
			return
		}
		if warningsAreFatal {
			numFailures += numWarnings
		}
		if numFailures > 0 && numFailures < numExpectedSuccesses {
			logrus.Fatalln("Push partially failed.")
		}
		if numFailures == numExpectedSuccesses {
			logrus.Fatalln("Push failed.")
		}
	},
}
