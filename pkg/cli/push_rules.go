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
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/fairwindsops/insights-cli/pkg/rules"
)

var pushRulesSubDir string

const defaultPushRulesSubDir = "rules"

func init() {
	// This flag sets a variable defined in the parent `push` command.
	pushRulesCmd.PersistentFlags().StringVarP(&pushRulesSubDir, "push-rules-subdirectory", "", defaultPushRulesSubDir, "Sub-directory within push-directory, to contain automation rules.")
	pushCmd.AddCommand(pushRulesCmd)
}

var pushRulesCmd = &cobra.Command{
	Use:    "rules",
	Short:  "Push automation rules.",
	Long:   "Push automation rules to Insights.",
	PreRun: validateAndLoadInsightsAPIConfigWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		org := configurationObject.Options.Organization
		err := rules.PushRules(client, pushDir+"/"+pushRulesSubDir, org, pushDelete, pushDryRun)
		if err != nil {
			logrus.Fatalf("Unable to push rules: %v", err)
		}
		logrus.Infoln("Push succeeded.")
	},
}
