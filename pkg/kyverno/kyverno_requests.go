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
	kyvernoPoliciesURLFormat       = "/v0/organizations/%s/kyverno-policies"
	kyvernoPolicyURLFormat         = "/v0/organizations/%s/kyverno-policies/%s"
	kyvernoPolicyValidateURLFormat = "/v0/organizations/%s/kyverno-policies/validate"
	kyvernoPolicyBulkURLFormat     = "/v0/organizations/%s/kyverno-policies/bulk"
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
	var policies []KyvernoPolicy
	if !utils.IsSuccessful(resp.StatusCode) {
		logrus.Errorf("FetchKyvernoPolicies: invalid response code: %s %v", string(resp.Bytes()), resp.StatusCode)
		return nil, errors.New("FetchKyvernoPolicies: invalid response code")
	}
	err = resp.Unmarshal(&policies)
	if err != nil {
		logrus.Errorf("Unable to convert response to json for Kyverno policies: %v", err)
		return nil, err
	}
	return policies, nil
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

// DeleteKyvernoPolicy deletes a Kyverno policy
func DeleteKyvernoPolicy(client *req.Client, org string, policyName string) error {
	url := fmt.Sprintf(kyvernoPolicyURLFormat, org, policyName)
	resp, err := client.R().SetHeaders(utils.GetHeaders("")).Delete(url)
	if err != nil {
		logrus.Errorf("Unable to delete Kyverno policy %s from insights: %v", policyName, err)
		return err
	}
	if !utils.IsSuccessful(resp.StatusCode) {
		logrus.Errorf("DeleteKyvernoPolicy: invalid response code: %s %v", string(resp.Bytes()), resp.StatusCode)
		return errors.New("DeleteKyvernoPolicy: invalid response code")
	}
	return nil
}

// ValidateKyvernoPolicyWithExpectedOutcomes validates a Kyverno policy with test resources
func ValidateKyvernoPolicyWithExpectedOutcomes(client *req.Client, org string, policy KyvernoPolicy, testResources []TestResource, expectOutcomes bool) (*ValidationResultWithExpectedOutcomes, error) {
	url := fmt.Sprintf(kyvernoPolicyValidateURLFormat, org)

	requestBody := map[string]interface{}{
		"policy_yaml":     policyToYAML(policy),
		"test_resources":  testResources,
		"expect_outcomes": expectOutcomes,
	}

	resp, err := client.R().SetHeaders(utils.GetHeaders("")).SetBody(&requestBody).Post(url)
	if err != nil {
		logrus.Errorf("Unable to validate Kyverno policy %s: %v", policy.Name, err)
		return nil, err
	}

	var result ValidationResultWithExpectedOutcomes
	if !utils.IsSuccessful(resp.StatusCode) {
		logrus.Errorf("ValidateKyvernoPolicyWithExpectedOutcomes: invalid response code: %s %v", string(resp.Bytes()), resp.StatusCode)
		return nil, errors.New("ValidateKyvernoPolicyWithExpectedOutcomes: invalid response code")
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

	requestBody := map[string]interface{}{
		"policies":       policies,
		"delete_missing": deleteMissing,
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
