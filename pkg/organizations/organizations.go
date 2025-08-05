package organizations

import (
	"fmt"

	"github.com/fairwindsops/insights-cli/pkg/utils"
	"github.com/imroc/req/v3"
	"github.com/sirupsen/logrus"
)

const organizationURLFormat = "/v0/organizations/%s"

func ManageOrganizationPolicyMappings(client *req.Client, org string, enable bool) error {
	mode := "scan-all"
	if enable {
		mode = "app-groups"
	}
	url := fmt.Sprintf(organizationURLFormat, org)
	logrus.Debugf("enable/disable: policyMappings URL: %s, value: %v", url, enable)
	body := map[string]string{"PolicyStrategy": mode}
	resp, err := client.R().SetHeaders(utils.GetHeaders("")).SetBody(&body).Patch(url)
	if err != nil {
		return fmt.Errorf("unable to fetch policy-mappings from insights: %w", err)
	}
	if !utils.IsSuccessful(resp.Response.StatusCode) {
		return fmt.Errorf("invalid response code - expected 200, got %d: %s", resp.Response.StatusCode, string(resp.Bytes()))
	}
	return nil
}
