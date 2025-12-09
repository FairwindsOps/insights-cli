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

package kyverno

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"go.yaml.in/yaml/v3"
)

// KyvernoPolicy represents a Kyverno policy
type KyvernoPolicy struct {
	Name              string         `json:"name" yaml:"metadata.name"`
	Kind              string         `json:"kind" yaml:"kind"`
	APIVersion        string         `json:"apiVersion" yaml:"apiVersion"`
	Namespace         string         `json:"namespace,omitempty" yaml:"metadata.namespace"`
	Labels            map[string]any `json:"labels,omitempty" yaml:"metadata.labels"`
	Annotations       map[string]any `json:"annotations,omitempty" yaml:"metadata.annotations"`
	Metadata          map[string]any `json:"metadata,omitempty" yaml:"metadata"` // Full metadata for any additional fields
	Spec              map[string]any `json:"spec" yaml:"spec"`
	Status            map[string]any `json:"status,omitempty"`
	ManagedByInsights *bool          `json:"managedByInsights,omitempty" yaml:"managedByInsights,omitempty"`
	CreatedAt         *time.Time     `json:"createdAt,omitempty" yaml:"createdAt,omitempty"`
	UpdatedAt         *time.Time     `json:"updatedAt,omitempty" yaml:"updatedAt,omitempty"`
}

func (k KyvernoPolicy) GetYamlBytes() ([]byte, error) {
	yamlBytes, err := convertPolicySpecToYAML(k)
	if err != nil {
		return nil, fmt.Errorf("failed to convert policy spec to YAML: %w", err)
	}
	return []byte(yamlBytes), nil
}

// Helper function to convert a KyvernoPolicy spec to YAML string
func convertPolicySpecToYAML(policy KyvernoPolicy) (string, error) {
	// Start with the full metadata if available, otherwise create a new map
	var metadata map[string]any
	if len(policy.Metadata) > 0 {
		// Deep copy the metadata to avoid modifying the original
		metadata = make(map[string]any)
		for k, v := range policy.Metadata {
			metadata[k] = v
		}
	} else {
		metadata = make(map[string]any)
	}

	// Ensure name is set in metadata
	metadata["name"] = policy.Name

	// Add namespace if set
	if policy.Namespace != "" {
		metadata["namespace"] = policy.Namespace
	}

	// Add labels if they exist and are not empty
	if len(policy.Labels) > 0 {
		metadata["labels"] = policy.Labels
	}

	// Add annotations if they exist
	if len(policy.Annotations) > 0 {
		metadata["annotations"] = policy.Annotations
	}

	// Create the full policy structure
	policyMap := map[string]any{
		"apiVersion": policy.APIVersion,
		"kind":       policy.Kind,
		"metadata":   metadata,
		"spec":       policy.Spec,
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

// GetName implements the nameable interface for download functionality
func (k KyvernoPolicy) GetName() string {
	return k.Name
}

// ValidationResult represents the result of policy validation
type ValidationResult struct {
	Valid          bool         `json:"valid"`
	Message        string       `json:"message"`
	ValidationType string       `json:"validation_type"`
	Errors         []string     `json:"errors,omitempty"`
	Warnings       []string     `json:"warnings,omitempty"`
	TestResults    []TestResult `json:"test_results,omitempty"`
}

// TestResource represents a test resource for validation
type TestResource struct {
	Content         string `json:"content"`
	FileName        string `json:"fileName"`
	PolicyName      string `json:"policyName"`
	TestCaseName    string `json:"testCaseName"`
	ExpectedOutcome string `json:"expectedOutcome"`
}

// TestResult represents the result of a single test case
type TestResult struct {
	TestCaseName    string `json:"test_case_name"`
	FileName        string `json:"file_name"`
	ExpectedOutcome string `json:"expected_outcome"`
	ActualOutcome   string `json:"actual_outcome"`
	Passed          bool   `json:"passed"`
}

// PolicyWithTestCases represents a policy with its associated test cases
type PolicyWithTestCases struct {
	Policy    KyvernoPolicy
	TestCases []TestResource
}

// BulkUpsertResponse represents the response from bulk upsert operations
type BulkUpsertResponse struct {
	Created int      `json:"created"`
	Updated int      `json:"updated"`
	Deleted int      `json:"deleted"`
	Errors  []string `json:"errors"`
}

// KyvernoPolicyList represents the response from the list endpoint
type KyvernoPolicyList struct {
	Policies []KyvernoPolicy `json:"policies"`
	Total    int             `json:"total"`
}

// KyvernoPolicyInput represents the input format expected by the API
type KyvernoPolicyInput struct {
	Name              string            `json:"name"`
	Kind              string            `json:"kind"`
	APIVersion        string            `json:"apiVersion"`
	Namespace         string            `json:"namespace,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	Annotations       map[string]string `json:"annotations,omitempty"`
	Metadata          map[string]any    `json:"metadata,omitempty"` // Full metadata for any additional fields
	Spec              map[string]any    `json:"spec"`
	Status            *map[string]any   `json:"status,omitempty"`
	ManagedByInsights *bool             `json:"managedByInsights,omitempty"`
}

// ValidationRequest represents the request format for policy validation
type ValidationRequest struct {
	Policy    string   `json:"policy"`
	Resources []string `json:"resources,omitempty"`
}

// ToKyvernoPolicyInput converts a KyvernoPolicy to KyvernoPolicyInput format
func (k KyvernoPolicy) ToKyvernoPolicyInput() KyvernoPolicyInput {
	// Convert labels from map[string]any to map[string]string
	labels := make(map[string]string)
	for key, value := range k.Labels {
		if str, ok := value.(string); ok {
			labels[key] = str
		}
	}

	// Convert annotations from map[string]any to map[string]string
	annotations := make(map[string]string)
	for key, value := range k.Annotations {
		if str, ok := value.(string); ok {
			annotations[key] = str
		}
	}

	// Convert status to pointer if it exists
	var status *map[string]any
	if k.Status != nil {
		status = &k.Status
	}

	return KyvernoPolicyInput{
		Name:              k.Name,
		Kind:              k.Kind,
		APIVersion:        k.APIVersion,
		Namespace:         k.Namespace,
		Labels:            labels,
		Annotations:       annotations,
		Metadata:          k.Metadata,
		Spec:              k.Spec,
		Status:            status,
		ManagedByInsights: k.ManagedByInsights,
	}
}

// ClusterValidationResponse represents the response from cluster policy validation
type ClusterValidationResponse struct {
	Cluster           string                          `json:"cluster"`
	TotalPolicies     int                             `json:"totalPolicies"`
	ValidPolicies     int                             `json:"validPolicies"`
	InvalidPolicies   int                             `json:"invalidPolicies"`
	PolicyResults     []ClusterPolicyValidationResult `json:"policyResults"`
	ValidationSummary ClusterPolicyValidationSummary  `json:"validationSummary"`
}

// ClusterPolicyValidationResult represents the validation result for a single policy in a cluster
type ClusterPolicyValidationResult struct {
	PolicyName       string             `json:"policyName"`
	Status           string             `json:"status"`
	AppGroupsApplied []string           `json:"appGroupsApplied"`
	ValidationResult ValidationResponse `json:"validationResult"`
}

// ClusterPolicyValidationSummary represents the overall validation summary for a cluster
type ClusterPolicyValidationSummary struct {
	OverallStatus string `json:"overallStatus"`
	TotalErrors   int    `json:"totalErrors"`
	TotalWarnings int    `json:"totalWarnings"`
}

// ValidationResponse represents the basic validation response structure
type ValidationResponse struct {
	Valid           bool     `json:"valid"`
	Errors          []string `json:"errors,omitempty"`
	Warnings        []string `json:"warnings,omitempty"`
	ValidationType  string   `json:"validationType,omitempty"`
	ResourcesTested int      `json:"resourcesTested,omitempty"`
	RulesApplied    int      `json:"rulesApplied,omitempty"`
}
