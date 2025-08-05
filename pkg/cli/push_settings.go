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

	"github.com/fairwindsops/insights-cli/pkg/policies"
)

func init() {
	pushCmd.AddCommand(pushSettingsCmd)
}

var pushSettingsCmd = &cobra.Command{
	Use:    "settings",
	Short:  "Push policies configuration.",
	Long:   "Push policies configuration to Insights to streamline settings across multiple Insights plugins.",
	PreRun: validateAndLoadInsightsAPIConfigWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		org := configurationObject.Options.Organization
		host := configurationObject.Options.Hostname
		err := policies.PushPolicies(client, pushDir, org, insightsToken, host, pushDryRun)
		if err != nil {
			logrus.Fatalf("Unable to push policies configuration: %v", err)
		}
		logrus.Infoln("Push succeeded.")
	},
}
