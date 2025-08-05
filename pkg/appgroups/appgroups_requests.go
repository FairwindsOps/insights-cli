package appgroups

import (
	"fmt"

	"github.com/fairwindsops/insights-cli/pkg/utils"
	"github.com/imroc/req/v3"
	"github.com/sirupsen/logrus"
)

const (
	appGroupURLFormat       = "/v0/organizations/%s/app-groups"
	appGroupURLSingleFormat = "/v0/organizations/%s/app-groups/%s"
)

// FetchAppGroups queries Fairwinds Insights to retrieve all of the app-groups for an organization
func FetchAppGroups(client *req.Client, org string) ([]AppGroup, error) {
	url := fmt.Sprintf(appGroupURLFormat, org)
	logrus.Debugf("fetchAppGroups: appGroups URL: %s", url)
	resp, err := client.R().SetHeaders(utils.GetHeaders("")).Get(url)
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
func upsertAppGroup(client *req.Client, org string, appGroup AppGroup) error {
	url := fmt.Sprintf(appGroupURLFormat, org)
	logrus.Debugf("upsertAppGroup: appGroups URL: %s", url)
	resp, err := client.R().SetHeaders(utils.GetHeaders("")).SetBody(&appGroup).Post(url)
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
func deleteAppGroup(client *req.Client, org string, appGroup AppGroup) error {
	url := fmt.Sprintf(appGroupURLSingleFormat, org, appGroup.Name)
	logrus.Debugf("deleteAppGroup: appGroups URL: %s", url)
	resp, err := client.R().SetHeaders(utils.GetHeaders("")).Delete(url)
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
