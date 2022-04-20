// Copyright 2020 FairwindsOps Inc //
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

// Package policies provides an interface to Insights policy settings and
// enforcement.
package policies

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/imroc/req"
	"github.com/sirupsen/logrus"

	cliversion "github.com/fairwindsops/insights-cli/pkg/version"
)

const policiesPutURLFormat = "%s/v0/organizations/%s/policies"

// PutPolicies submits the io.Reader as an HTTP PUT request, content-type
// text/yaml, to the Insights API policies endpoint.
func PutPolicies(policies io.Reader, org, token, hostName string) error {
	url := fmt.Sprintf(policiesPutURLFormat, hostName, org)
	resp, err := req.Post(url, getHeaders(token), policies)
	if err != nil {
		return err
	}
	if resp.Response().StatusCode != http.StatusOK {
		return fmt.Errorf("invalid HTTP response %d %s", resp.Response().StatusCode, string(resp.Bytes()))
	}
	return nil
}

// PushPolicies verifies the policies settings file is readable, then pushes
// it to the Insights API.
func PushPolicies(pushDir, org, insightsToken, host string, dryrun bool) error {
	if pushDir == "" {
		return errors.New("pushDir cannot be empty")
	}
	policiesFileName := pushDir + "/settings.yaml"
	logrus.Infof("Pushing policies configuration from %s", policiesFileName)
	_, err := os.Stat(policiesFileName)
	if err != nil {
		return err
	}
	if dryrun {
		logrus.Infoln("NOTE: The policies configuration is currently validated at the time it is submitted to Insights. Pushing this configuration in dry-run mode only validates the file is readable.")
		return nil
	}
	policies, err := os.Open(policiesFileName)
	if err != nil {
		return err
	}
	err = PutPolicies(policies, org, insightsToken, host)
	if err != nil {
		return err
	}
	logrus.Debugln("Done pushing policies configuration")
	return nil
}

// getHeaders returns headers to be used when communicating with e Insights API for
// policies configuration.
func getHeaders(token string) req.Header {
	return req.Header{
		"Content-Type":            "text/yaml",
		"X-Fairwinds-CLI-Version": cliversion.GetVersion(),
		"Authorization":           fmt.Sprintf("Bearer %s", token),
		"Accept":                  "application/json",
	}
}
