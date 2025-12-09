// Copyright 2025 FairwindsOps Inc
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

	"github.com/fairwindsops/insights-cli/pkg/kyverno"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var downloadKyvernoPoliciesSubDir string

func init() {
	downloadKyvernoPoliciesCmd.PersistentFlags().StringVar(&downloadKyvernoPoliciesSubDir, "download-subdirectory", "kyverno-policies", "Sub-directory within download-directory, to download Kyverno policies.")
	downloadCmd.AddCommand(downloadKyvernoPoliciesCmd)
}

var downloadKyvernoPoliciesCmd = &cobra.Command{
	Use:   "kyverno-policies",
	Short: "Download Kyverno policies from Insights to local files.",
	Long:  "Download Kyverno policies from Insights to local files. This creates/updates the local kyverno-policies directory structure for synchronization.",
	Example: `
	# Download all policies from Insights
	insights-cli download kyverno-policies -d .

	# Download to custom subdirectory
	insights-cli download kyverno-policies -d . --download-subdirectory my-policies

	# Download to specific directory
	insights-cli download kyverno-policies -d /path/to/my/project

	# Download to specific directory with custom subdirectory
	insights-cli download kyverno-policies -d /path/to/my/project --download-subdirectory policies

	# Download with override
	insights-cli download kyverno-policies -d . --override`,
	PreRun: validateAndLoadInsightsAPIConfigWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		org := configurationObject.Options.Organization
		kyvernoPolicies, err := kyverno.FetchKyvernoPolicies(client, org)
		if err != nil {
			logrus.Fatalf("unable to fetch kyverno-policies from insights: %v", err)
		}

		// Build the full save directory path
		saveDir := downloadDir + "/" + downloadKyvernoPoliciesSubDir

		// Ensure the save directory exists
		err = os.MkdirAll(saveDir, 0755)
		if err != nil {
			logrus.Fatalf("unable to create directory %s: %v", saveDir, err)
		}

		// We do not want to delete files that contains .success.yaml and .failure.yaml.
		c, err := saveEntitiesLocally(saveDir, kyvernoPolicies, overrideLocalFiles, []string{".success.yaml", ".failure.yaml"})
		if err != nil {
			logrus.Fatalf("error saving kyverno-policies locally: %v", err)
		}

		logrus.Infof("Downloaded %d kyverno-policies from Insights to %s\n", c, saveDir)
		logrus.Infof("You can now add test cases and push changes back to Insights\n")
	},
}
