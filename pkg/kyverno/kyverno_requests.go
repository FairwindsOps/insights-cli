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
	"errors"
	"fmt"

	"github.com/imroc/req/v3"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/fairwindsops/insights-cli/pkg/utils"
)

const (
	kyvernoPoliciesURLFormat                     = "/v0/organizations/%s/kyverno-policies"
	kyvernoPolicyURLFormat                       = "/v0/organizations/%s/kyverno-policies/%s"
	kyvernoPolicyValidateURLFormat               = "/v0/organizations/%s/kyverno-policies/validate"
	kyvernoPolicyBulkURLFormat                   = "/v0/organizations/%s/kyverno-policies/bulk"
	clusterKyvernoPoliciesWithAppGroupsURLFormat = "/v0/organizations/%s/clusters/%s/kyverno-policies/with-app-groups-applied"
	clusterKyvernoPoliciesYamlURLFormat          = "/v0/organizations/%s/clusters/%s/kyverno-policies/with-app-groups-applied/yaml"
	clusterKyvernoPoliciesValidateURLFormat      = "/v0/organizations/%s/clusters/%s/kyverno-policies/with-app-groups-applied/validate"
)

// FetchKyvernoPolicies queries Fairwinds Insights to retrieve all Kyverno policies for an organization
func FetchKyvernoPolicies(client *req.Client, org string) ([]KyvernoPolicy, error) {
	url := fmt.Sprintf(kyvernoPoliciesURLFormat, org)
	logrus.Debugf("Kyverno policies URL: %s", url)
	resp, err := client.R().SetHeaders(utils.GetHeaders("")).Get(url)
	if err != nil {
		logrus.Errorf("Unable to get Kyverno policies from insights: %v", err)
		return nil, err
	}
	var policyList KyvernoPolicyList
	if !utils.IsSuccessful(resp.StatusCode) {
		logrus.Errorf("FetchKyvernoPolicies: invalid response code: %s %v", string(resp.Bytes()), resp.StatusCode)
		return nil, errors.New("FetchKyvernoPolicies: invalid response code")
	}
	err = resp.Unmarshal(&policyList)
	if err != nil {
		logrus.Errorf("Unable to convert response to json for Kyverno policies: %v", err)
		return nil, err
	}
	return policyList.Policies, nil
}

// UpsertKyvernoPolicy creates or updates a Kyverno policy
func UpsertKyvernoPolicy(client *req.Client, org string, policy KyvernoPolicy) error {
	url := fmt.Sprintf(kyvernoPolicyURLFormat, org, policy.Name)
	resp, err := client.R().SetHeaders(utils.GetHeaders("")).SetBody(&policy).Post(url)
	if err != nil {
		logrus.Errorf("Unable to upsert Kyverno policy %s to insights: %v", policy.Name, err)
		return err
	}
	if !utils.IsSuccessful(resp.StatusCode) {
		logrus.Errorf("UpsertKyvernoPolicy: invalid response code: %s %v", string(resp.Bytes()), resp.StatusCode)
		return errors.New("UpsertKyvernoPolicy: invalid response code")
	}
	return nil
}

// ValidateKyvernoPolicy validates a Kyverno policy with test resources
func ValidateKyvernoPolicy(client *req.Client, org string, policy KyvernoPolicy, testResources []TestResource, expectOutcomes bool) (*ValidationResult, error) {
	url := fmt.Sprintf(kyvernoPolicyValidateURLFormat, org)

	// Convert test resources to string array for the API
	var resources []string
	for _, testResource := range testResources {
		resources = append(resources, testResource.Content)
	}

	requestBody := ValidationRequest{
		Policy:    policyToYAML(policy),
		Resources: resources,
	}

	resp, err := client.R().SetHeaders(utils.GetHeaders("")).SetBody(&requestBody).Post(url)
	if err != nil {
		logrus.Errorf("Unable to validate Kyverno policy %s: %v", policy.Name, err)
		return nil, err
	}

	var result ValidationResult
	if !utils.IsSuccessful(resp.StatusCode) {
		logrus.Errorf("ValidateKyvernoPolicy: invalid response code: %s %v", string(resp.Bytes()), resp.StatusCode)
		return nil, errors.New("ValidateKyvernoPolicy: invalid response code")
	}

	err = resp.Unmarshal(&result)
	if err != nil {
		logrus.Errorf("Unable to convert validation response to json: %v", err)
		return nil, err
	}

	return &result, nil
}

// BulkUpsertKyvernoPolicies performs bulk upsert operations
func BulkUpsertKyvernoPolicies(client *req.Client, org string, policies []KyvernoPolicy, deleteMissing bool) (*BulkUpsertResponse, error) {
	url := fmt.Sprintf(kyvernoPolicyBulkURLFormat, org)

	// Convert policies to the format expected by the API
	policyInputs := make([]KyvernoPolicyInput, len(policies))
	for i, policy := range policies {
		policyInputs[i] = policy.ToKyvernoPolicyInput()
	}

	requestBody := map[string]interface{}{
		"policies":      policyInputs,
		"deleteMissing": deleteMissing,
	}

	resp, err := client.R().SetHeaders(utils.GetHeaders("")).SetBody(&requestBody).Put(url)
	if err != nil {
		logrus.Errorf("Unable to bulk upsert Kyverno policies: %v", err)
		return nil, err
	}

	var result BulkUpsertResponse
	if !utils.IsSuccessful(resp.StatusCode) {
		logrus.Errorf("BulkUpsertKyvernoPolicies: invalid response code: %s %v", string(resp.Bytes()), resp.StatusCode)
		return nil, errors.New("BulkUpsertKyvernoPolicies: invalid response code")
	}

	err = resp.Unmarshal(&result)
	if err != nil {
		logrus.Errorf("Unable to convert bulk upsert response to json: %v", err)
		return nil, err
	}

	return &result, nil
}

// Helper function to convert policy to YAML string
func policyToYAML(policy KyvernoPolicy) string {
	// Create a map structure that matches the Kyverno policy format
	policyMap := map[string]interface{}{
		"apiVersion": policy.APIVersion,
		"kind":       policy.Kind,
		"metadata": map[string]interface{}{
			"name": policy.Name,
		},
	}

	// Add labels if present
	if len(policy.Labels) > 0 {
		metadata := policyMap["metadata"].(map[string]interface{})
		metadata["labels"] = policy.Labels
	}

	// Add annotations if present
	if len(policy.Annotations) > 0 {
		metadata := policyMap["metadata"].(map[string]interface{})
		metadata["annotations"] = policy.Annotations
	}

	// Add spec if present
	if len(policy.Spec) > 0 {
		policyMap["spec"] = policy.Spec
	}

	// Add status if present
	if len(policy.Status) > 0 {
		policyMap["status"] = policy.Status
	}

	// Marshal to YAML
	yamlBytes, err := yaml.Marshal(policyMap)
	if err != nil {
		logrus.Errorf("Unable to marshal policy to YAML: %v", err)
		// Fallback to basic format if marshaling fails
		return fmt.Sprintf("apiVersion: %s\nkind: %s\nmetadata:\n  name: %s\n",
			policy.APIVersion, policy.Kind, policy.Name)
	}

	return string(yamlBytes)
}

// FetchClusterKyvernoPoliciesWithAppGroups queries Fairwinds Insights to retrieve Kyverno policies for a specific cluster with app groups applied
func FetchClusterKyvernoPoliciesWithAppGroups(client *req.Client, org, cluster string) ([]KyvernoPolicy, error) {
	url := fmt.Sprintf(clusterKyvernoPoliciesWithAppGroupsURLFormat, org, cluster)
	logrus.Debugf("Cluster Kyverno policies with app groups URL: %s", url)
	resp, err := client.R().SetHeaders(utils.GetHeaders("")).Get(url)
	if err != nil {
		logrus.Errorf("Unable to get cluster Kyverno policies with app groups from insights: %v", err)
		return nil, err
	}
	var policyList KyvernoPolicyList
	if !utils.IsSuccessful(resp.StatusCode) {
		logrus.Errorf("FetchClusterKyvernoPoliciesWithAppGroups: invalid response code: %s %v", string(resp.Bytes()), resp.StatusCode)
		return nil, errors.New("FetchClusterKyvernoPoliciesWithAppGroups: invalid response code")
	}
	err = resp.Unmarshal(&policyList)
	if err != nil {
		logrus.Errorf("Unable to convert response to json for cluster Kyverno policies with app groups: %v", err)
		return nil, err
	}
	return policyList.Policies, nil
}

// ExportClusterKyvernoPoliciesYaml exports Kyverno policies for a specific cluster as YAML
func ExportClusterKyvernoPoliciesYaml(client *req.Client, org, cluster string) (string, error) {
	url := fmt.Sprintf(clusterKyvernoPoliciesYamlURLFormat, org, cluster)
	logrus.Debugf("Cluster Kyverno policies YAML export URL: %s", url)
	resp, err := client.R().SetHeaders(utils.GetHeaders("")).Get(url)
	if err != nil {
		logrus.Errorf("Unable to export cluster Kyverno policies YAML from insights: %v", err)
		return "", err
	}
	if !utils.IsSuccessful(resp.StatusCode) {
		logrus.Errorf("ExportClusterKyvernoPoliciesYaml: invalid response code: %s %v", string(resp.Bytes()), resp.StatusCode)
		return "", errors.New("ExportClusterKyvernoPoliciesYaml: invalid response code")
	}
	return string(resp.Bytes()), nil
}

// ValidateClusterKyvernoPolicies validates all Kyverno policies for a specific cluster
func ValidateClusterKyvernoPolicies(client *req.Client, org, cluster string) (*ClusterValidationResponse, error) {
	url := fmt.Sprintf(clusterKyvernoPoliciesValidateURLFormat, org, cluster)
	logrus.Debugf("Cluster Kyverno policies validation URL: %s", url)
	resp, err := client.R().SetHeaders(utils.GetHeaders("")).Get(url)
	if err != nil {
		logrus.Errorf("Unable to validate cluster Kyverno policies: %v", err)
		return nil, err
	}
	var result ClusterValidationResponse
	if !utils.IsSuccessful(resp.StatusCode) {
		logrus.Errorf("ValidateClusterKyvernoPolicies: invalid response code: %s %v", string(resp.Bytes()), resp.StatusCode)
		return nil, errors.New("ValidateClusterKyvernoPolicies: invalid response code")
	}
	err = resp.Unmarshal(&result)
	if err != nil {
		logrus.Errorf("Unable to convert validation response to json: %v", err)
		return nil, err
	}
	return &result, nil
}
