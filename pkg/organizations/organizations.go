package organizations

import (
	"encoding/json"
	"fmt"

	"github.com/fairwindsops/insights-cli/pkg/utils"
	"github.com/fairwindsops/insights-cli/pkg/version"
	"github.com/imroc/req/v3"
	"github.com/sirupsen/logrus"
)

const organizationURLFormat = "%s/v0/organizations/%s"

func ManageOrganizationPolicyMappings(org, token, hostName string, enable bool) error {
	mode := "scan-all"
	if enable {
		mode = "app-groups"
	}
	url := fmt.Sprintf(organizationURLFormat, hostName, org)
	logrus.Debugf("enable/disable: policyMappings URL: %s, value: %v", url, enable)
	bodyBytes, err := json.Marshal(map[string]string{"PolicyStrategy": mode})
	if err != nil {
		return err
	}
	resp, err := req.C().R().SetHeaders(utils.GetHeaders(version.GetVersion(), token, "")).SetBodyBytes(bodyBytes).Patch(url)
	if err != nil {
		return fmt.Errorf("unable to fetch policy-mappings from insights: %w", err)
	}
	if !utils.IsSuccessful(resp.Response.StatusCode) {
		return fmt.Errorf("invalid response code - expected 200, got %d: %s", resp.Response.StatusCode, string(resp.Bytes()))
	}
	return nil
}
