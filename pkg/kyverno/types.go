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
	"time"
)

// KyvernoPolicy represents a Kyverno policy
type KyvernoPolicy struct {
	ID             int                    `json:"id,omitempty"`
	OrganizationID int64                  `json:"organization_id,omitempty"`
	Name           string                 `json:"name"`
	Kind           string                 `json:"kind"`
	APIVersion     string                 `json:"api_version"`
	Labels         map[string]interface{} `json:"labels,omitempty"`
	Annotations    map[string]interface{} `json:"annotations,omitempty"`
	Spec           map[string]interface{} `json:"spec"`
	Status         map[string]interface{} `json:"status,omitempty"`
	CreatedAt      *time.Time             `json:"created_at,omitempty"`
	UpdatedAt      *time.Time             `json:"updated_at,omitempty"`
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
	Name        string                  `json:"name"`
	Kind        string                  `json:"kind"`
	APIVersion  string                  `json:"apiVersion"`
	Labels      map[string]string       `json:"labels,omitempty"`
	Annotations map[string]string       `json:"annotations,omitempty"`
	Spec        map[string]interface{}  `json:"spec"`
	Status      *map[string]interface{} `json:"status,omitempty"`
}

// ValidationRequest represents the request format for policy validation
type ValidationRequest struct {
	Policy    string   `json:"policy"`
	Resources []string `json:"resources,omitempty"`
}

// ToKyvernoPolicyInput converts a KyvernoPolicy to KyvernoPolicyInput format
func (k KyvernoPolicy) ToKyvernoPolicyInput() KyvernoPolicyInput {
	// Convert labels from map[string]interface{} to map[string]string
	labels := make(map[string]string)
	for key, value := range k.Labels {
		if str, ok := value.(string); ok {
			labels[key] = str
		}
	}

	// Convert annotations from map[string]interface{} to map[string]string
	annotations := make(map[string]string)
	for key, value := range k.Annotations {
		if str, ok := value.(string); ok {
			annotations[key] = str
		}
	}

	// Convert status to pointer if it exists
	var status *map[string]interface{}
	if k.Status != nil {
		status = &k.Status
	}

	return KyvernoPolicyInput{
		Name:        k.Name,
		Kind:        k.Kind,
		APIVersion:  k.APIVersion,
		Labels:      labels,
		Annotations: annotations,
		Spec:        k.Spec,
		Status:      status,
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
