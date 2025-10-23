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
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/fairwindsops/insights-cli/pkg/kyverno"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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

		c, err := saveEntitiesLocally(saveDir, kyvernoPolicies, overrideLocalFiles)
		if err != nil {
			logrus.Fatalf("error saving kyverno-policies locally: %v", err)
		}

		logrus.Infof("Downloaded %d kyverno-policies from Insights to %s\n", c, saveDir)
		logrus.Infof("You can now add test cases and push changes back to Insights\n")
	},
}

// Helper function to convert a KyvernoPolicy spec to YAML string
func convertPolicySpecToYAML(policy kyverno.KyvernoPolicy) (string, error) {
	// Create the full policy structure
	policyMap := map[string]any{
		"apiVersion": policy.APIVersion,
		"kind":       policy.Kind,
		"metadata": map[string]any{
			"name": policy.Name,
		},
		"spec": policy.Spec,
	}

	// Add labels and annotations if they exist and are not empty
	if len(policy.Labels) > 0 {
		policyMap["metadata"].(map[string]any)["labels"] = policy.Labels
	}

	// Initialize annotations map
	annotations := make(map[string]any)

	// Add existing annotations if they exist
	if len(policy.Annotations) > 0 {
		policyMap["metadata"].(map[string]any)["annotations"] = policy.Annotations
	}

	// Only add annotations if there are any
	if len(annotations) > 0 {
		policyMap["metadata"].(map[string]any)["annotations"] = annotations
	}

	// Add status only if it exists and is not null
	if len(policy.Status) > 0 {
		policyMap["status"] = policy.Status
	}
	// Clean up any null values before marshaling
	cleanPolicyMap := cleanNullValues(policyMap)

	// Convert to YAML
	// yaml converter should use only 2 spaces for indentation
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	err := encoder.Encode(cleanPolicyMap)
	if err != nil {
		return "", fmt.Errorf("failed to marshal policy to YAML: %w", err)
	}
	yamlContent := buf.String()
	yamlContent = strings.TrimSpace(yamlContent)
	return yamlContent, nil
}

// cleanNullValues recursively removes null values from a map
func cleanNullValues(data any) any {
	switch v := data.(type) {
	case map[string]any:
		cleaned := make(map[string]any)
		for key, value := range v {
			if value != nil {
				cleanedValue := cleanNullValues(value)
				if cleanedValue != nil {
					cleaned[key] = cleanedValue
				}
			}
		}
		return cleaned
	case []any:
		var cleaned []any
		for _, item := range v {
			if item != nil {
				cleanedItem := cleanNullValues(item)
				if cleanedItem != nil {
					cleaned = append(cleaned, cleanedItem)
				}
			}
		}
		return cleaned
	case nil:
		return nil
	case string:
		// Don't process strings, return as-is
		return v
	default:
		return v
	}
}
