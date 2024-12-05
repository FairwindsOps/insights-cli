package appgroups

import (
	"encoding/json"
	"fmt"

	"github.com/fairwindsops/insights-cli/pkg/utils"
	"github.com/fairwindsops/insights-cli/pkg/version"
	"github.com/imroc/req/v3"
	"github.com/sirupsen/logrus"
)

const (
	appGroupURLFormat       = "%s/v0/organizations/%s/app-groups"
	appGroupURLSingleFormat = "%s/v0/organizations/%s/app-groups/%s"
)

// FetchAppGroups queries Fairwinds Insights to retrieve all of the app-groups for an organization
func FetchAppGroups(org, token, hostName string) ([]AppGroup, error) {
	url := fmt.Sprintf(appGroupURLFormat, hostName, org)
	logrus.Debugf("fetchAppGroups: appGroups URL: %s", url)
	resp, err := req.C().R().SetHeaders(utils.GetHeaders(version.GetVersion(), token, "")).Get(url)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch app-groups from insights: %w", err)
	}
	var appGroups []AppGroup
	if !utils.IsSuccessful(resp.Response.StatusCode) {
		return nil, fmt.Errorf("invalid response code - expected 200, got %d: %s", resp.Response.StatusCode, string(resp.Bytes()))
	}
	err = resp.Unmarshal(&appGroups)
	if err != nil {
		return nil, fmt.Errorf("unable to convert response to json for app-groups: %w", err)
	}
	return appGroups, nil
}

// upsertAppGroup requests Fairwinds Insights to upsert an app-group for an organization
func upsertAppGroup(org, token, hostName string, appGroup AppGroup) error {
	url := fmt.Sprintf(appGroupURLFormat, hostName, org)
	logrus.Debugf("upsertAppGroup: appGroups URL: %s", url)
	bodyBytes, err := json.Marshal(appGroup)
	if err != nil {
		return err
	}
	resp, err := req.C().R().SetHeaders(utils.GetHeaders(version.GetVersion(), token, "")).SetBodyBytes(bodyBytes).Post(url)
	if err != nil {
		return fmt.Errorf("unable to fetch app-groups from insights: %w", err)
	}
	if !utils.IsSuccessful(resp.Response.StatusCode) {
		return fmt.Errorf("invalid response code - expected 200, got %d: %s", resp.Response.StatusCode, string(resp.Bytes()))
	}
	var response AppGroup
	err = resp.Unmarshal(&response)
	if err != nil {
		return fmt.Errorf("Unable to convert response to json for app-groups: %w", err)

	}
	return nil
}

// deleteAppGroup requests Fairwinds Insights to remove an app-group for an organization
func deleteAppGroup(org, token, hostName string, appGroup AppGroup) error {
	url := fmt.Sprintf(appGroupURLSingleFormat, hostName, org, appGroup.Name)
	logrus.Debugf("deleteAppGroup: appGroups URL: %s", url)
	resp, err := req.C().R().SetHeaders(utils.GetHeaders(version.GetVersion(), token, "")).Delete(url)
	if err != nil {
		return fmt.Errorf("unable to fetch app-groups from insights: %w", err)
	}
	if !utils.IsSuccessful(resp.Response.StatusCode) {
		return fmt.Errorf("invalid response code - expected 200, got %d: %s", resp.Response.StatusCode, string(resp.Bytes()))
	}
	var response AppGroup
	err = resp.Unmarshal(&response)
	if err != nil {
		return fmt.Errorf("Unable to convert response to json for app-groups: %w", err)
	}
	return nil
}
