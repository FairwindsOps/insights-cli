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
	"github.com/fairwindsops/insights-cli/pkg/policymappings"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var deleteMissingPolicyMappings bool
var pushPolicyMappingsSubDir string

const defaultPushPolicyMappingsSubDir = "policy-mappings"

func init() {
	pushPolicyMappingsCmd.PersistentFlags().BoolVarP(&deleteMissingPolicyMappings, "delete", "D", false, "Delete any policy-mappings from Fairwinds Insights that are not present in the local directory.")
	pushPolicyMappingsCmd.PersistentFlags().StringVarP(&pushPolicyMappingsSubDir, "push-policy-mappings-subdirectory", "", defaultPushPolicyMappingsSubDir, "Sub-directory within push-directory, to contain policy-mappings.")
	pushCmd.AddCommand(pushPolicyMappingsCmd)
}

var pushPolicyMappingsCmd = &cobra.Command{
	Use:    "policy-mappings",
	Short:  "Push policy-mappings.",
	Long:   "Push policy-mappings to Fairwinds Insights.",
	Hidden: hideAppGroupCommands,
	PreRun: validateAndLoadInsightsAPIConfigWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		org := configurationObject.Options.Organization
		host := configurationObject.Options.Hostname
		err := policymappings.PushPolicyMappings(pushDir+"/"+pushPolicyMappingsSubDir, org, insightsToken, host, deleteMissingPolicyMappings)
		if err != nil {
			logrus.Fatalf("Unable to push policy-mappings: %v", err)
		}
		logrus.Infoln("Push succeeded.")
	},
}
