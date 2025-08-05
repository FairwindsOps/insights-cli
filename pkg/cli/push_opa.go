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

	"github.com/fairwindsops/insights-cli/pkg/opa"
)

var pushOPASubDir string
var deleteMissingOPA bool
var pushRegoVersion string

const defaultPushOPASubDir = "opa"

func init() {
	pushOPACmd.PersistentFlags().BoolVarP(&deleteMissingOPA, "delete", "D", false, "Delete any OPA policies from Insights that are not present in the local directory.")
	// This flag sets a variable defined in the parent `push` command.
	pushOPACmd.PersistentFlags().StringVarP(&pushOPASubDir, "push-opa-subdirectory", "", defaultPushOPASubDir, "Sub-directory within push-directory, to contain OPA policies.")
	pushOPACmd.PersistentFlags().StringVarP(&pushRegoVersion, "rego-version", "v", "v0", "The version of Rego used to compile the policies.")
	pushCmd.AddCommand(pushOPACmd)
}

var pushOPACmd = &cobra.Command{
	Use:    "opa",
	Short:  "Push OPA policies.",
	Long:   "Push OPA policies to Insights.",
	PreRun: validateAndLoadInsightsAPIConfigWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		org := configurationObject.Options.Organization
		err := opa.PushOPAChecks(client, pushDir+"/"+pushOPASubDir, org, deleteMissingOPA, pushDryRun, pushRegoVersion)
		if err != nil {
			logrus.Fatalf("Unable to push OPA Checks: %v", err)
		}
		logrus.Infoln("Push succeeded.")
	},
}
