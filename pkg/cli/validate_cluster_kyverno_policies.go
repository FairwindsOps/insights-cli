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

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/fairwindsops/insights-cli/pkg/kyverno"
)

func init() {
	kyvernoPoliciesClusterCmd.AddCommand(validateClusterKyvernoPoliciesCmd)
}

var validateClusterKyvernoPoliciesCmd = &cobra.Command{
	Use:   "validate <cluster>",
	Short: "Validate all Kyverno policies for a cluster",
	Long:  "Validate all Kyverno policies that apply to the specified cluster with app groups applied.",
	Example: `
	# Validate all policies for a cluster
	insights-cli kyverno-policies cluster validate my-cluster`,
	Args:   cobra.ExactArgs(1),
	PreRun: validateAndLoadInsightsAPIConfigWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		cluster := args[0]
		org := configurationObject.Options.Organization

		result, err := kyverno.ValidateClusterKyvernoPolicies(client, org, cluster)
		if err != nil {
			logrus.Fatalf("Unable to validate cluster Kyverno policies: %v", err)
		}

		// Display validation results
		kyverno.DisplayClusterValidationResults(result)

		if result.ValidationSummary.OverallStatus != "success" {
			os.Exit(1)
		}
	},
}
