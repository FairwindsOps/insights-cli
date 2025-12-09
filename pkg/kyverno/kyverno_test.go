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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsPolicyFile(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"policy.yaml", true},
		{"policy.yml", true},
		{"policy.testcase1.success.yaml", false},
		{"policy.testcase1.failure.yaml", false},
		{"policy.txt", false},
		{"policy", false},
	}

	for _, test := range tests {
		t.Run(test.filename, func(t *testing.T) {
			result := isPolicyFile(test.filename)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestIsTestCaseFile(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"policy.yaml", false},
		{"policy.yml", false},
		{"policy.testcase1.success.yaml", true},
		{"policy.testcase1.failure.yaml", true},
		{"policy.txt", false},
		{"policy", false},
	}

	for _, test := range tests {
		t.Run(test.filename, func(t *testing.T) {
			result := isTestCaseFile(test.filename)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestExtractPolicyNameFromFile(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"require-labels.yaml", "require-labels"},
		{"disallow-privileged.yml", "disallow-privileged"},
		{"policy.yaml", "policy"},
	}

	for _, test := range tests {
		t.Run(test.filename, func(t *testing.T) {
			result := extractPolicyNameFromFile(test.filename)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestExtractPolicyNameFromTestCase(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"require-labels.testcase1.success.yaml", "require-labels"},
		{"disallow-privileged.testcase1.failure.yaml", "disallow-privileged"},
		{"policy.testcase2.success.yml", "policy"},
	}

	for _, test := range tests {
		t.Run(test.filename, func(t *testing.T) {
			result := extractPolicyNameFromTestCase(test.filename)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestExtractTestCaseName(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"require-labels.testcase1.success.yaml", "testcase1.success"},
		{"disallow-privileged.testcase2.failure.yaml", "testcase2.failure"},
		{"policy.testcase3.success.yml", "testcase3.success"},
		{"require-labels.success.yaml", "success"},
		{"disallow-privileged.failure.yaml", "failure"},
		{"policy.success.yml", "success"},
	}

	for _, test := range tests {
		t.Run(test.filename, func(t *testing.T) {
			result := extractTestCaseName(test.filename)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestDetermineExpectedOutcome(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"require-labels.testcase1.success.yaml", "success"},
		{"disallow-privileged.testcase1.failure.yaml", "failure"},
		{"policy.testcase2.success.yml", "success"},
		{"policy.testcase3.failure.yml", "failure"},
		{"policy.yaml", "unknown"},
	}

	for _, test := range tests {
		t.Run(test.filename, func(t *testing.T) {
			result := determineExpectedOutcome(test.filename)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestDisplayClusterValidationResults(t *testing.T) {
	// Test successful validation
	result := &ClusterValidationResponse{
		Cluster:         "test-cluster",
		TotalPolicies:   3,
		ValidPolicies:   2,
		InvalidPolicies: 1,
		PolicyResults: []ClusterPolicyValidationResult{
			{
				PolicyName:       "policy1",
				Status:           "valid",
				AppGroupsApplied: []string{"app-group-1"},
				ValidationResult: ValidationResponse{
					Valid: true,
				},
			},
			{
				PolicyName:       "policy2",
				Status:           "valid",
				AppGroupsApplied: []string{"app-group-2"},
				ValidationResult: ValidationResponse{
					Valid: true,
				},
			},
			{
				PolicyName:       "policy3",
				Status:           "invalid",
				AppGroupsApplied: []string{"app-group-3"},
				ValidationResult: ValidationResponse{
					Valid:  false,
					Errors: []string{"Policy validation failed"},
				},
			},
		},
		ValidationSummary: ClusterPolicyValidationSummary{
			OverallStatus: "partial_success",
			TotalErrors:   1,
			TotalWarnings: 0,
		},
	}

	// This test just ensures the function doesn't panic
	// In a real test environment, we'd capture stdout and verify the output
	DisplayClusterValidationResults(result)
}

func TestConvertPolicySpecToYAML(t *testing.T) {
	policy := KyvernoPolicy{
		Name:       "test-policy",
		Kind:       "ClusterPolicy",
		APIVersion: "kyverno.io/v1",
		Spec: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": "test-policy",
			},
		},
	}

	yaml, err := convertPolicySpecToYAML(policy)
	assert.NoError(t, err)
	assert.Equal(t, "apiVersion: kyverno.io/v1\nkind: ClusterPolicy\nmetadata:\n  name: test-policy\nspec:\n  metadata:\n    name: test-policy", yaml)
}

func TestReadPolicyFromFileWithLabelsAndAnnotations(t *testing.T) {
	policy, err := readPolicyFromFile("testdata/policy-with-metadata.yaml")
	assert.NoError(t, err)

	// Verify basic fields
	assert.Equal(t, "policy-with-metadata", policy.Name)
	assert.Equal(t, "ClusterPolicy", policy.Kind)
	assert.Equal(t, "kyverno.io/v1", policy.APIVersion)

	// Verify labels are extracted
	assert.NotNil(t, policy.Labels)
	assert.Equal(t, "my-app", policy.Labels["app"])
	assert.Equal(t, "production", policy.Labels["environment"])
	assert.Equal(t, "security", policy.Labels["team"])

	// Verify annotations are extracted
	assert.NotNil(t, policy.Annotations)
	assert.Equal(t, "This is a test policy with labels and annotations", policy.Annotations["description"])
	assert.Equal(t, "platform-team", policy.Annotations["owner"])

	// Verify spec is extracted
	assert.NotNil(t, policy.Spec)
}

func TestReadPolicyFromFileWithoutLabelsAndAnnotations(t *testing.T) {
	policy, err := readPolicyFromFile("testdata/disallow-privileged.yaml")
	assert.NoError(t, err)

	// Verify basic fields
	assert.Equal(t, "disallow-privileged", policy.Name)
	assert.Equal(t, "Policy", policy.Kind)
	assert.Equal(t, "kyverno.io/v1", policy.APIVersion)

	// Labels and annotations should be nil or empty
	assert.Empty(t, policy.Labels)
	assert.Empty(t, policy.Annotations)

	// Verify spec is extracted
	assert.NotNil(t, policy.Spec)
}

func TestReadPolicyFromFileWithNamespaceAndFullMetadata(t *testing.T) {
	policy, err := readPolicyFromFile("testdata/namespaced-policy.yaml")
	assert.NoError(t, err)

	// Verify basic fields
	assert.Equal(t, "namespaced-policy", policy.Name)
	assert.Equal(t, "Policy", policy.Kind)
	assert.Equal(t, "kyverno.io/v1", policy.APIVersion)

	// Verify namespace is extracted
	assert.Equal(t, "my-namespace", policy.Namespace)

	// Verify labels are extracted
	assert.NotNil(t, policy.Labels)
	assert.Equal(t, "test-app", policy.Labels["app"])

	// Verify annotations are extracted
	assert.NotNil(t, policy.Annotations)
	assert.Equal(t, "A namespaced policy", policy.Annotations["description"])

	// Verify full metadata is preserved (including generateName, finalizers, etc.)
	assert.NotNil(t, policy.Metadata)
	assert.Equal(t, "namespaced-policy", policy.Metadata["name"])
	assert.Equal(t, "my-namespace", policy.Metadata["namespace"])
	assert.Equal(t, "test-prefix-", policy.Metadata["generateName"])

	// Check finalizers are preserved
	finalizers, ok := policy.Metadata["finalizers"].([]interface{})
	assert.True(t, ok, "finalizers should be a slice")
	assert.Len(t, finalizers, 1)
	assert.Equal(t, "kyverno.io/finalizer", finalizers[0])

	// Verify spec is extracted
	assert.NotNil(t, policy.Spec)
}

func TestToKyvernoPolicyInputPreservesAllFields(t *testing.T) {
	policy := KyvernoPolicy{
		Name:       "test-policy",
		Kind:       "Policy",
		APIVersion: "kyverno.io/v1",
		Namespace:  "test-namespace",
		Labels: map[string]any{
			"app": "my-app",
		},
		Annotations: map[string]any{
			"description": "Test policy",
		},
		Metadata: map[string]any{
			"name":         "test-policy",
			"namespace":    "test-namespace",
			"generateName": "prefix-",
		},
		Spec: map[string]any{
			"validationFailureAction": "enforce",
		},
	}

	input := policy.ToKyvernoPolicyInput()

	// Verify all fields are preserved
	assert.Equal(t, "test-policy", input.Name)
	assert.Equal(t, "Policy", input.Kind)
	assert.Equal(t, "kyverno.io/v1", input.APIVersion)
	assert.Equal(t, "test-namespace", input.Namespace)
	assert.Equal(t, "my-app", input.Labels["app"])
	assert.Equal(t, "Test policy", input.Annotations["description"])
	assert.NotNil(t, input.Metadata)
	assert.Equal(t, "prefix-", input.Metadata["generateName"])
	assert.NotNil(t, input.Spec)
}
