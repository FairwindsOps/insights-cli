package insights

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/fairwindsops/insights-plugins/opa/pkg/opa"
	"github.com/imroc/req"
	"github.com/sirupsen/logrus"
)

const opaURLFormat = "%s/v0/organizations/%s/opa/customChecks"

const opaCheckURLFormat = opaURLFormat + "/%s"

// GetChecks queries Fairwinds Insights to retrieve all of the Checks for an organization
func GetChecks(org, token, hostName string) ([]opa.OPACustomCheck, error) {
	url := fmt.Sprintf(opaURLFormat, hostName, org)
	logrus.Infof("Url: %s", url)
	resp, err := req.Get(url, getHeaders(token))
	if err != nil {
		return nil, err
	}
	var checks []opa.OPACustomCheck
	if resp.Response().StatusCode != http.StatusOK {
		logrus.Errorf("Invalid response code: %s %v", string(resp.Bytes()), resp.Response().StatusCode)
		return nil, errors.New("invalid response code")
	}
	err = resp.ToJSON(&checks)
	if err != nil {
		return nil, err
	}
	return checks, nil
}

// GetInstances queries Fairwinds Insights to retrieve all of the instances for a given check
func GetInstances(org, checkName, token, hostName string) ([]opa.CheckSetting, error) {
	url := fmt.Sprintf(opaCheckURLFormat, hostName, org, checkName)
	resp, err := req.Get(url, getHeaders(token))
	if err != nil {
		return nil, err
	}
	var instances []opa.CheckSetting
	err = resp.ToJSON(&instances)
	if err != nil {
		return nil, err
	}
	return instances, nil
}

func getHeaders(token string) req.Header {
	return req.Header{
		"Authorization": fmt.Sprintf("Bearer %s", token),
		"Accept":        "application/json",
	}
}
