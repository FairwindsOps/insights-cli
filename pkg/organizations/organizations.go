package organizations

import (
	"fmt"

	"github.com/fairwindsops/insights-cli/pkg/utils"
	"github.com/fairwindsops/insights-cli/pkg/version"
	"github.com/imroc/req"
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
	resp, err := req.Patch(url, utils.GetHeaders(version.GetVersion(), token), req.BodyJSON(map[string]string{"PolicyStrategy": mode}))
	if err != nil {
		return fmt.Errorf("unable to fetch policy-mappings from insights: %w", err)
	}
	if !utils.IsSuccessful(resp.Response().StatusCode) {
		return fmt.Errorf("invalid response code - expected 200, got %d: %s", resp.Response().StatusCode, string(resp.Bytes()))
	}
	return nil
}
