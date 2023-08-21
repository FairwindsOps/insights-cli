package appgroups

import (
	"fmt"
	"net/http"

	cliversion "github.com/fairwindsops/insights-cli/pkg/version"
	"github.com/imroc/req"
	"github.com/sirupsen/logrus"
	"github.com/xlab/treeprint"
)

const (
	appGroupURLFormat       = "%s/v0/organizations/%s/app-groups"
	appGroupURLFormatSingle = "%s/v0/organizations/%s/app-groups/%d"
)

// BuildAppGroupsTree builds a tree for app-groups
func BuildAppGroupsTree(org, token, hostName string) (treeprint.Tree, error) {
	appGroups, err := fetchAppGroups(org, token, hostName)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch app-groups from insights: %v", err)
	}
	tree := treeprint.New()
	appGroupsBranch := tree.AddBranch("app-groups")
	for _, appGroup := range appGroups {
		rulesNode := appGroupsBranch.AddBranch(appGroup.Name)
		if appGroup.Spec.Enabled != nil {
			rulesNode.AddNode(fmt.Sprintf("Enabled: %t", *appGroup.Spec.Enabled))
		}
	}
	return tree, nil
}

// fetchAppGroups queries Fairwinds Insights to retrieve all of the app-groups for an organization
func fetchAppGroups(org, token, hostName string) ([]AppGroup, error) {
	url := fmt.Sprintf(appGroupURLFormat, hostName, org)
	logrus.Debugf("AppGroups URL: %s", url)
	resp, err := req.Get(url, getHeaders(token))
	if err != nil {
		return nil, fmt.Errorf("unable to fetch app-groups from insights: %w", err)
	}
	var appGroups []AppGroup
	if resp.Response().StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid response code - expected 200, got %d: %s", resp.Response().StatusCode, string(resp.Bytes()))
	}
	err = resp.ToJSON(&appGroups)
	if err != nil {
		logrus.Errorf("Unable to convert response to json for rules: %v", err)
		return nil, err
	}
	return appGroups, nil
}

func getHeaders(token string) req.Header {
	return req.Header{
		"X-Fairwinds-CLI-Version": cliversion.GetVersion(),
		"Authorization":           fmt.Sprintf("Bearer %s", token),
		"Accept":                  "application/json",
	}
}
