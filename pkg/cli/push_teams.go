// Copyright 2024 FairwindsOps Inc
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
	"github.com/fairwindsops/insights-cli/pkg/teams"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	pushCmd.AddCommand(pushTeamsCmd)
}

var pushTeamsCmd = &cobra.Command{
	Use:    "teams",
	Short:  "Push teams configuration.",
	Long:   "Push teams configuration to Insights for user permissions management.",
	PreRun: validateAndLoadInsightsAPIConfigWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		org := configurationObject.Options.Organization
		host := configurationObject.Options.Hostname
		err := teams.PushTeams(pushDir, org, insightsToken, host, pushDeleteNonProvidedTeams, pushDryRun)
		if err != nil {
			logrus.Fatalf("Unable to push teams configuration: %v", err)
		}
		logrus.Infoln("Push succeeded.")
	},
}
