package policymappings

import (
	"fmt"

	"github.com/imroc/req/v3"
	"github.com/sirupsen/logrus"
)

const (
	policyMappingURLFormat       = "/v0/organizations/%s/policy-mappings"
	policyMappingURLSingleFormat = "/v0/organizations/%s/policy-mappings/%s"
)

// FetchPolicyMappings queries Fairwinds Insights to retrieve all of the policy-mappings for an organization
func FetchPolicyMappings(client *req.Client, org string) ([]PolicyMapping, error) {
	url := fmt.Sprintf(policyMappingURLFormat, org)
	logrus.Debugf("fetchPolicyMappings: policyMappings URL: %s", url)
	resp, err := client.R().SetHeaders(getHeaders()).Get(url)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch policy-mappings from insights: %w", err)
	}
	if resp.IsErrorState() {
		return nil, fmt.Errorf("invalid response code - expected 200, got %d: %s", resp.StatusCode, string(resp.Bytes()))
	}
	var policyMappings []PolicyMapping
	err = resp.Unmarshal(&policyMappings)
	if err != nil {
		return nil, fmt.Errorf("unable to convert response to json for policy-mappings: %w", err)
	}
	return policyMappings, nil
}

// upsertPolicyMapping requests Fairwinds Insights to upsert an policy-mapping for an organization
func upsertPolicyMapping(client *req.Client, org string, policyMapping PolicyMapping) error {
	url := fmt.Sprintf(policyMappingURLFormat, org)
	logrus.Debugf("upsertPolicyMapping: policyMappings URL: %s", url)
	resp, err := client.R().SetHeaders(getHeaders()).SetBody(&policyMapping).Post(url)
	if err != nil {
		return fmt.Errorf("unable to fetch policy-mapping from insights: %w", err)
	}
	if resp.IsErrorState() {
		return fmt.Errorf("invalid response code - expected 200, got %d: %s", resp.StatusCode, string(resp.Bytes()))
	}
	var response PolicyMapping
	err = resp.Unmarshal(&response)
	if err != nil {
		return fmt.Errorf("unable to convert response to json for policy-mapping: %w", err)

	}
	return nil
}

// deletePolicyMapping requests Fairwinds Insights to remove an policy-mapping for an organization
func deletePolicyMapping(client *req.Client, org string, policyMapping PolicyMapping) error {
	url := fmt.Sprintf(policyMappingURLSingleFormat, org, policyMapping.Name)
	logrus.Debugf("deletePolicyMapping: policyMappings URL: %s", url)
	resp, err := client.R().SetHeaders(getHeaders()).Delete(url)
	if err != nil {
		return fmt.Errorf("unable to fetch policy-mapping from insights: %w", err)
	}
	if resp.IsErrorState() {
		return fmt.Errorf("invalid response code - expected 200, got %d: %s", resp.StatusCode, string(resp.Bytes()))
	}
	var response PolicyMapping
	err = resp.Unmarshal(&response)
	if err != nil {
		return fmt.Errorf("unable to convert response to json for policy-mapping: %w", err)
	}
	return nil
}
