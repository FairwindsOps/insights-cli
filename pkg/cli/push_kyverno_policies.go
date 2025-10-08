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
	"os"
	"path/filepath"

	"github.com/fairwindsops/insights-cli/pkg/kyverno"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var pushKyvernoPoliciesSubDir string
var pushSpecificPolicies []string
var pushSkipValidation bool
var pushForce bool

const defaultPushKyvernoPoliciesSubDir = "kyverno-policies"

func init() {
	pushKyvernoPoliciesCmd.PersistentFlags().StringVarP(&pushKyvernoPoliciesSubDir, "push-kyverno-policies-subdirectory", "", defaultPushKyvernoPoliciesSubDir, "Sub-directory within push-directory, to contain Kyverno policies.")
	pushKyvernoPoliciesCmd.PersistentFlags().StringSliceVarP(&pushSpecificPolicies, "policies", "p", []string{}, "Specific policy names to push (e.g., require-labels,disallow-privileged). If not specified, all policies will be pushed.")
	pushKyvernoPoliciesCmd.PersistentFlags().BoolVar(&pushSkipValidation, "skip-validation", false, "Skip validation before pushing (not recommended).")
	pushKyvernoPoliciesCmd.PersistentFlags().BoolVar(&pushForce, "force", false, "Force push even if validation fails (use with extreme caution).")
	pushCmd.AddCommand(pushKyvernoPoliciesCmd)
}

var pushKyvernoPoliciesCmd = &cobra.Command{
	Use:   "kyverno-policies [-p policy1,policy2]",
	Short: "Push Kyverno policies from local files to Insights.",
	Long:  "Push Kyverno policies from local files to Insights. This command automatically validates all policies before pushing. If ANY validation fails, the push operation is aborted unless --force is used.",
	Example: `
	# Push all local policies to Insights (with validation)
	insights-cli push kyverno-policies -d .

	# Push specific policies (with validation)
	insights-cli push kyverno-policies -d . -p require-labels,disallow-privileged

	# Push with dry run to see what would be changed
	insights-cli push kyverno-policies -d . --dry-run

	# Skip validation (not recommended)
	insights-cli push kyverno-policies -d . --skip-validation

	# Force push even if validation fails (use with extreme caution)
	insights-cli push kyverno-policies -d . --force`,
	PreRun: validateAndLoadInsightsAPIConfigWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		org := configurationObject.Options.Organization
		policyDir := pushDir + "/" + pushKyvernoPoliciesSubDir

		// Check if the policy directory exists
		_, err := os.Stat(policyDir)
		if err != nil {
			logrus.Fatalf("Policy directory %s does not exist. Run 'insights-cli download kyverno-policies -d %s' first to create it.", policyDir, filepath.Dir(policyDir))
		}

		// Get all policy files (excluding test cases)
		allPolicies, err := kyverno.GetPolicyFilesForPush(policyDir)
		if err != nil {
			logrus.Fatalf("Unable to read policy files: %v", err)
		}

		// Filter policies if specific ones are requested
		var policiesToPush []kyverno.KyvernoPolicy
		if len(pushSpecificPolicies) > 0 {
			for _, requestedPolicy := range pushSpecificPolicies {
				found := false
				for _, policy := range allPolicies {
					if policy.Name == requestedPolicy {
						policiesToPush = append(policiesToPush, policy)
						found = true
						break
					}
				}
				if !found {
					logrus.Warnf("Policy '%s' not found in directory", requestedPolicy)
				}
			}
		} else {
			policiesToPush = allPolicies
		}

		if len(policiesToPush) == 0 {
			logrus.Info("No policies to push")
			return
		}

		// Validate policies before pushing (unless skipped or forced)
		if !pushSkipValidation {
			logrus.Info("Validating policies before push...")

			// Discover all policies and test cases for validation
			policiesWithTestCases, err := kyverno.DiscoverPoliciesAndTestCases(policyDir)
			if err != nil {
				logrus.Fatalf("Unable to discover policies for validation: %v", err)
			}

			// Validate each policy that will be pushed
			validationFailed := false
			for _, policyToPush := range policiesToPush {
				logrus.Infof("Validating policy: %s", policyToPush.Name)

				// Find test cases for this policy
				var testCases []kyverno.TestResource
				for _, policyWithTestCases := range policiesWithTestCases {
					if policyWithTestCases.Policy.Name == policyToPush.Name {
						testCases = policyWithTestCases.TestCases
						break
					}
				}

				// Validate the policy
				result, err := kyverno.ValidateKyvernoPolicyWithExpectedOutcomes(
					client, org, policyToPush, testCases, true)
				if err != nil {
					logrus.Errorf("❌ Unable to validate policy %s: %v", policyToPush.Name, err)
					validationFailed = true
					continue
				}

				// Display validation results
				displayValidationResultsWithExpectedOutcomes(result)
				if !result.Valid {
					logrus.Errorf("❌ Policy %s failed validation", policyToPush.Name)
					validationFailed = true
				} else {
					logrus.Infof("✅ Policy %s passed validation", policyToPush.Name)
				}
			}

			// If ANY validation failed, check if force push is enabled
			if validationFailed {
				if pushForce {
					logrus.Warnf("⚠️  Validation failed but --force flag is set. Proceeding with push anyway...")
					logrus.Warnf("⚠️  WARNING: You are pushing policies that failed validation!")
				} else {
					logrus.Fatalf("🚫 Push aborted: One or more policies failed validation. Please fix the issues before pushing to Insights, or use --force to override.")
				}
			} else {
				logrus.Info("✅ All policies validated successfully!")
			}
		}

		if pushDryRun {
			logrus.Infof("Dry run: Would synchronize %d Kyverno policies with Insights:", len(policiesToPush))
			for _, policy := range policiesToPush {
				logrus.Infof("  - %s (%s)", policy.Name, policy.Kind)
			}
			if pushDelete {
				logrus.Info("Dry run: Would delete policies that exist in Insights but not locally")
			}
			if pushForce {
				logrus.Warnf("⚠️  Dry run: Force push is enabled - validation failures would be ignored")
			}
			return
		}

		// Push to Insights
		err = kyverno.PushKyvernoPolicies(client, policiesToPush, org, pushDelete, pushDryRun)
		if err != nil {
			logrus.Fatalf("Unable to synchronize kyverno-policies with Insights: %v", err)
		}

		if pushForce {
			logrus.Warnf("🎉 Force push completed. Policies have been pushed to Insights despite validation failures.")
		} else {
			logrus.Infoln("🎉 Successfully synchronized kyverno-policies with Insights.")
		}
	},
}
