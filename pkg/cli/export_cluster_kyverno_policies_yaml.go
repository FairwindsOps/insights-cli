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
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/fairwindsops/insights-cli/pkg/kyverno"
)

func init() {
	kyvernoPoliciesClusterCmd.AddCommand(exportClusterKyvernoPoliciesYamlCmd)
}

var exportClusterKyvernoPoliciesYamlCmd = &cobra.Command{
	Use:   "export-yaml <cluster>",
	Short: "Export cluster Kyverno policies as YAML",
	Long:  "Export Kyverno policies for a specific cluster with app groups applied as YAML ready for kubectl apply.",
	Example: `
	# Export policies as YAML
	insights-cli kyverno-policies cluster export-yaml my-cluster
	
	# Export to file
	insights-cli kyverno-policies cluster export-yaml my-cluster > cluster-policies.yaml
	
	# Apply directly to cluster
	insights-cli kyverno-policies cluster export-yaml my-cluster | kubectl apply -f -`,
	Args:   cobra.ExactArgs(1),
	PreRun: validateAndLoadInsightsAPIConfigWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		cluster := args[0]
		org := configurationObject.Options.Organization

		yamlContent, err := kyverno.ExportClusterKyvernoPoliciesYaml(client, org, cluster)
		if err != nil {
			logrus.Fatalf("Unable to export cluster Kyverno policies: %v", err)
		}

		fmt.Print(yamlContent)
	},
}
