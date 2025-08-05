// Copyright 2023 FairwindsOps Inc //
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
	"os"

	"github.com/imroc/req/v3"
	"github.com/sirupsen/logrus"
)

const policiesPutURLFormat = "/v0/organizations/%s/policies"

// PutPolicies submits the io.Reader as an HTTP PUT request, content-type
// text/yaml, to the Insights API policies endpoint.
func PutPolicies(client *req.Client, policies io.Reader, org string) error {
	url := fmt.Sprintf(policiesPutURLFormat, org)
	bodyBytes, err := io.ReadAll(policies)
	if err != nil {
		return err
	}
	resp, err := client.R().SetHeaders(getHeaders()).SetBodyBytes(bodyBytes).Post(url)
	if err != nil {
		return err
	}
	if resp.IsErrorState() {
		return fmt.Errorf("invalid HTTP response %d %s", resp.StatusCode, string(resp.Bytes()))
	}
	return nil
}

// PushPolicies verifies the policies settings file is readable, then pushes
// it to the Insights API.
func PushPolicies(client *req.Client, pushDir, org string, dryRun bool) error {
	if pushDir == "" {
		return errors.New("pushDir cannot be empty")
	}
	policiesFileName := pushDir + "/settings.yaml"
	logrus.Infof("Pushing policies configuration from %s", policiesFileName)
	_, err := os.Stat(policiesFileName)
	if err != nil {
		return err
	}
	if dryRun {
		logrus.Infoln("NOTE: The policies configuration is currently validated at the time it is submitted to Insights. Pushing this configuration in dry-run mode only validates the file is readable.")
		return nil
	}
	policies, err := os.Open(policiesFileName)
	if err != nil {
		return err
	}
	err = PutPolicies(client, policies, org)
	if err != nil {
		return err
	}
	logrus.Debugln("Done pushing policies configuration")
	return nil
}

// getHeaders returns headers to be used when communicating with e Insights API for
// policies configuration.
func getHeaders() map[string]string {
	return map[string]string{
		"Content-Type": "text/yaml",
		"Accept":       "application/yaml",
	}
}
