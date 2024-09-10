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

	"github.com/fairwindsops/insights-cli/pkg/opa"
)

var pushExternalOPAFile string
var pushExternalOPAHeaders []string

func init() {
	pushExternalOPACmd.PersistentFlags().BoolVarP(&deleteMissingOPA, "delete", "D", false, "Delete any OPA policies from Insights that are not present in the local directory.")
	// This flag sets a variable defined in the parent `push` command.
	pushExternalOPACmd.PersistentFlags().StringVarP(&pushOPASubDir, "subdirectory", "s", defaultPushOPASubDir, "Sub-directory within push-directory, to contain the external OPA policy file.")
	pushExternalOPACmd.PersistentFlags().StringVarP(&pushExternalOPAFile, "file", "f", "external_sources.yaml", "file name to be used for external OPA policies")
	pushExternalOPACmd.PersistentFlags().StringSliceVarP(&pushExternalOPAHeaders, "header", "", []string{}, "these headers are passed to the external OPA policies server. i.e.: for authentication")
	pushCmd.AddCommand(pushExternalOPACmd)
}

var pushExternalOPACmd = &cobra.Command{
	Use:    "external-opa",
	Short:  "Push External OPA policies.",
	Long:   "Push External OPA policies to Insights.",
	PreRun: validateAndLoadInsightsAPIConfigWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		org := configurationObject.Options.Organization
		host := configurationObject.Options.Hostname
		filePath := fmt.Sprintf("%s/%s/%s", pushDir, pushOPASubDir, pushExternalOPAFile)
		err := opa.PushExternalOPAChecks(filePath, org, insightsToken, pushExternalOPAHeaders, host, deleteMissingOPA, pushDryRun)
		if err != nil {
			logrus.Fatalf("Unable to push OPA Checks: %v", err)
		}
		logrus.Infoln("Push succeeded.")
	},
}
