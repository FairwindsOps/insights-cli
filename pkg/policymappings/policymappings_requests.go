package policymappings

import (
	"fmt"

	"github.com/fairwindsops/insights-cli/pkg/utils"
	"github.com/imroc/req"
	"github.com/sirupsen/logrus"
)

const (
	policyMappingURLFormat       = "%s/v0/organizations/%s/policy-mappings"
	policyMappingURLSingleFormat = "%s/v0/organizations/%s/policy-mappings/%s"
)

// FetchPolicyMappings queries Fairwinds Insights to retrieve all of the policy-mappings for an organization
func FetchPolicyMappings(org, token, hostName string) ([]PolicyMapping, error) {
	url := fmt.Sprintf(policyMappingURLFormat, hostName, org)
	logrus.Debugf("fetchPolicyMappings: policyMappings URL: %s", url)
	resp, err := req.Get(url, getHeaders(token))
	if err != nil {
		return nil, fmt.Errorf("unable to fetch policy-mappings from insights: %w", err)
	}
	if !utils.IsSuccessful(resp.Response().StatusCode) {
		return nil, fmt.Errorf("invalid response code - expected 200, got %d: %s", resp.Response().StatusCode, string(resp.Bytes()))
	}
	var policyMappings []PolicyMapping
	err = resp.ToJSON(&policyMappings)
	if err != nil {
		return nil, fmt.Errorf("Unable to convert response to json for policy-mappings: %w", err)
	}
	return policyMappings, nil
}

// upsertPolicyMapping requests Fairwinds Insights to upsert an policy-mapping for an organization
func upsertPolicyMapping(org, token, hostName string, policyMapping PolicyMapping) error {
	url := fmt.Sprintf(policyMappingURLFormat, hostName, org)
	logrus.Debugf("upsertPolicyMapping: policyMappings URL: %s", url)
	resp, err := req.Post(url, getHeaders(token), req.BodyJSON(&policyMapping))
	if err != nil {
		return fmt.Errorf("unable to fetch policy-mapping from insights: %w", err)
	}
	if !utils.IsSuccessful(resp.Response().StatusCode) {
		return fmt.Errorf("invalid response code - expected 200, got %d: %s", resp.Response().StatusCode, string(resp.Bytes()))
	}
	var response PolicyMapping
	err = resp.ToJSON(&response)
	if err != nil {
		return fmt.Errorf("Unable to convert response to json for policy-mapping: %w", err)

	}
	return nil
}

// deletePolicyMapping requests Fairwinds Insights to remove an policy-mapping for an organization
func deletePolicyMapping(org, token, hostName string, policyMapping PolicyMapping) error {
	url := fmt.Sprintf(policyMappingURLSingleFormat, hostName, org, policyMapping.Name)
	logrus.Debugf("deletePolicyMapping: policyMappings URL: %s", url)
	resp, err := req.Delete(url, getHeaders(token))
	if err != nil {
		return fmt.Errorf("unable to fetch policy-mapping from insights: %w", err)
	}
	if !utils.IsSuccessful(resp.Response().StatusCode) {
		return fmt.Errorf("invalid response code - expected 200, got %d: %s", resp.Response().StatusCode, string(resp.Bytes()))
	}
	var response PolicyMapping
	err = resp.ToJSON(&response)
	if err != nil {
		return fmt.Errorf("Unable to convert response to json for policy-mapping: %w", err)
	}
	return nil
}
