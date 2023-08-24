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
	"github.com/fairwindsops/insights-cli/pkg/appgroups"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var deleteMissingAppGroups bool
var pushAppGroupsSubDir string

func init() {
	pushAppGroupsCmd.PersistentFlags().BoolVarP(&deleteMissingAppGroups, "delete", "D", false, "Delete any app groups from Fairwinds Insights that are not present in the local directory.")
	// This flag sets a variable defined in the parent `push` command.
	pushAppGroupsCmd.PersistentFlags().StringVarP(&pushAppGroupsSubDir, "push-app-groups-subdirectory", "", "app-groups", "Sub-directory within push-directory, to contain app-groups.")
	pushCmd.AddCommand(pushAppGroupsCmd)
}

var pushAppGroupsCmd = &cobra.Command{
	Use:    "app-groups",
	Short:  "Push app-groups.",
	Long:   "Push app-groups to Fairwinds Insights.",
	PreRun: validateAndLoadInsightsAPIConfigWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		org := configurationObject.Options.Organization
		host := configurationObject.Options.Hostname
		err := appgroups.PushAppGroups(pushDir+"/"+pushAppGroupsSubDir, org, insightsToken, host, deleteMissingAppGroups)
		if err != nil {
			logrus.Fatalf("Unable to push app-groups: %v", err)
		}
		logrus.Infoln("Push succeeded.")
	},
}
