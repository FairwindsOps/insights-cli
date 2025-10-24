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
		{"require-labels.testcase1.success.yaml", "testcase1"},
		{"disallow-privileged.testcase2.failure.yaml", "testcase2"},
		{"policy.testcase3.success.yml", "testcase3"},
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
