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
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/fairwindsops/insights-cli/pkg/opa"
)

var pushExternalOPAFile string
var pushExternalOPASubDir string
var pushExternalOPAHeaders []string
var pushExternalRegoVersion string

const defaultPushExternalOPASubDir = "external-opa"

func init() {
	pushExternalOPACmd.PersistentFlags().BoolVarP(&deleteMissingOPA, "delete", "D", false, "Delete any OPA policies from Insights that are not present in the external OPA file definition.")
	// This flag sets a variable defined in the parent `push` command.
	pushExternalOPACmd.PersistentFlags().StringVarP(&pushExternalOPASubDir, "subdirectory", "s", defaultPushExternalOPASubDir, "Sub-directory within push-directory, to contain the external OPA file definition.")
	pushExternalOPACmd.PersistentFlags().StringVarP(&pushExternalOPAFile, "file", "f", "external-sources.yaml", "file name of the external OPA file definition.")
	pushExternalOPACmd.PersistentFlags().StringSliceVarP(&pushExternalOPAHeaders, "header", "", []string{}, "these headers are passed to the external service provider. i.e.: for authentication")
	pushExternalOPACmd.PersistentFlags().StringVarP(&pushExternalRegoVersion, "rego-version", "", "", "The version of Rego used to compile the policies.")
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
		filePath := fmt.Sprintf("%s/%s/%s", pushDir, pushExternalOPASubDir, pushExternalOPAFile)
		err := opa.PushExternalOPAChecks(filePath, org, insightsToken, pushExternalOPAHeaders, host, deleteMissingOPA, pushDryRun, &pushExternalRegoVersion)
		if err != nil {
			logrus.Fatalf("Unable to push external OPA checks: %v", err)
		}
		logrus.Infoln("Push succeeded.")
	},
}
