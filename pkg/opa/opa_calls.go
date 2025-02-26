// Copyright 2023 FairwindsOps Inc
//
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

package opa

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/fairwindsops/insights-cli/pkg/directory"
	"github.com/fairwindsops/insights-cli/pkg/models"
	"github.com/fairwindsops/insights-cli/pkg/utils"
	"github.com/fairwindsops/insights-cli/pkg/version"
	opaPlugin "github.com/fairwindsops/insights-plugins/plugins/opa/pkg/opa"
	"github.com/imroc/req/v3"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

const opaURLFormat = "%s/v0/organizations/%s/opa/customChecks"

const opaCheckURLFormat = opaURLFormat + "/%s"
const opaPutCheckURLFormat = opaCheckURLFormat + "?version=%.1f&?includeRegoV1=true"
const opaCheckInstancesURLFormat = opaCheckURLFormat + "/instances"

const opaInstanceURLFormat = opaCheckInstancesURLFormat + "/%s"

// GetChecks queries Fairwinds Insights to retrieve all of the Checks for an organization
func GetChecks(org, token, hostName string) ([]opaPlugin.OPACustomCheck, error) {
	url := fmt.Sprintf(opaURLFormat, hostName, org)
	logrus.Debugf("OPA URL: %s", url)
	resp, err := req.C().R().SetHeaders(utils.GetHeaders(version.GetVersion(), token, "")).Get(url)
	if err != nil {
		return nil, err
	}
	var checks []opaPlugin.OPACustomCheck
	if resp.Response.StatusCode != http.StatusOK {
		logrus.Errorf("GetChecks: Invalid response code: %s %v", string(resp.Bytes()), resp.Response.StatusCode)
		return nil, errors.New("GetChecks: invalid response code")
	}
	err = resp.Unmarshal(&checks)
	if err != nil {
		return nil, err
	}
	return checks, nil
}

// GetInstances queries Fairwinds Insights to retrieve all of the instances for a given check
func GetInstances(org, checkName, token, hostName string) ([]opaPlugin.CheckSetting, error) {
	url := fmt.Sprintf(opaCheckInstancesURLFormat, hostName, org, checkName)
	resp, err := req.C().R().SetHeaders(utils.GetHeaders(version.GetVersion(), token, "")).Get(url)
	if err != nil {
		return nil, err
	}
	if resp.Response.StatusCode != http.StatusOK {
		logrus.Errorf("GetInstances: Invalid response code: %s %v", string(resp.Bytes()), resp.Response.StatusCode)
		return nil, errors.New("GetInstances: invalid response code")
	}
	var instances []opaPlugin.CheckSetting
	err = resp.Unmarshal(&instances)
	if err != nil {
		return nil, err
	}
	return instances, nil
}

// DeleteCheck deletes an OPA Check from Fairwinds Insights
func DeleteCheck(check models.CustomCheckModel, org, token, hostName string) error {
	url := fmt.Sprintf(opaCheckURLFormat, hostName, org, check.CheckName)
	resp, err := req.C().R().SetHeaders(utils.GetHeaders(version.GetVersion(), token, "")).Delete(url)
	if err != nil {
		return err
	}
	if resp.Response.StatusCode != http.StatusOK {
		logrus.Errorf("DeleteCheck: Invalid response code: %s %v", string(resp.Bytes()), resp.Response.StatusCode)
		return errors.New("DeleteCheck: invalid response code")
	}
	return nil
}

type PutCheckRequest struct {
	Rego, Description string
	Disabled          *bool
	RegoVersion       string
}

// PutCheck upserts an OPA Check to Fairwinds Insights
func PutCheck(check models.CustomCheckModel, org, token, hostName string, pushRegoVersion string) error {
	url := fmt.Sprintf(opaPutCheckURLFormat, hostName, org, check.CheckName, check.Version)
	body := PutCheckRequest{Rego: check.Rego, Description: check.Description, Disabled: check.Disabled}
	if pushRegoVersion != "" {
		body.RegoVersion = pushRegoVersion
	} else {
		body.RegoVersion = "v0"
	}
	resp, err := req.C().R().SetHeaders(utils.GetHeaders(version.GetVersion(), token, "application/yaml")).SetBody(&body).Put(url)
	if err != nil {
		return err
	}
	if resp.Response.StatusCode != http.StatusOK {
		logrus.Errorf("PutCheck: Invalid response code: %s %v", string(resp.Bytes()), resp.Response.StatusCode)
		return errors.New("PutCheck: invalid response code")
	}
	return nil
}

// DeleteInstance deletes an Instance from Fairwinds Insights
func DeleteInstance(instance models.CustomCheckInstanceModel, org, token, hostName string) error {
	url := fmt.Sprintf(opaInstanceURLFormat, hostName, org, instance.CheckName, instance.InstanceName)
	resp, err := req.C().R().SetHeaders(utils.GetHeaders(version.GetVersion(), token, "")).Delete(url)
	if err != nil {
		return err
	}
	if resp.Response.StatusCode != http.StatusOK {
		logrus.Errorf("DeleteInstance: Invalid response code: %s %v", string(resp.Bytes()), resp.Response.StatusCode)
		return errors.New("DeleteInstance: invalid response code")
	}
	return nil
}

// PutInstance upserts an Instance to Fairwinds Insights
func PutInstance(instance models.CustomCheckInstanceModel, org, token, hostName string) error {
	url := fmt.Sprintf(opaInstanceURLFormat, hostName, org, instance.CheckName, instance.InstanceName)
	bodyBytes, err := yaml.Marshal(instance)
	if err != nil {
		return err
	}
	resp, err := req.C().R().SetHeaders(utils.GetHeaders(version.GetVersion(), token, "application/yaml")).SetBodyBytes(bodyBytes).Put(url)
	if err != nil {
		return err
	}
	if resp.Response.StatusCode != http.StatusOK {
		logrus.Errorf("PutInstance: Invalid response code: %s %v", string(resp.Bytes()), resp.Response.StatusCode)
		return errors.New("PutInstance: invalid response code")
	}
	return nil
}

// PushOPAChecks pushes OPA checks to Insights.
func PushOPAChecks(pushDir, org, insightsToken, host string, deleteMissing, dryRun bool, pushRegoVersion string) error {
	logrus.Debugln("Pushing OPA policies")
	_, err := os.Stat(pushDir)
	if err != nil {
		return err
	}
	files, err := directory.ScanOPAFolder(pushDir)
	if err != nil {
		return fmt.Errorf("error scanning directory: %w", err)
	}
	fileChecks, err := getChecksFromFiles(files)
	if err != nil {
		return fmt.Errorf("error Reading checks from files: %w", err)
	}
	results, err := CompareChecks(pushDir, org, insightsToken, host, fileChecks, deleteMissing)
	if err != nil {
		return err
	}
	for _, instance := range results.InstanceDelete {
		logrus.Infof("Deleting instance: %s for OPA policy %s", instance.InstanceName, instance.CheckName)
		if !dryRun {
			err := DeleteInstance(instance, org, insightsToken, host)
			if err != nil {
				return err
			}
		}
	}
	for _, check := range results.CheckDelete {
		logrus.Infof("Deleting OPA policy: %s", check.CheckName)
		if !dryRun {
			err := DeleteCheck(check, org, insightsToken, host)
			if err != nil {
				return err
			}
		}
	}
	for _, check := range results.CheckInsert {
		logrus.Infof("Adding v%.0f OPA policy: %s", check.Version, check.CheckName)
		if !dryRun {
			err := PutCheck(check, org, insightsToken, host, pushRegoVersion)
			if err != nil {
				return err
			}
		}
	}
	for _, check := range results.CheckUpdate {
		logrus.Infof("Updating v%.0f OPA policy: %s", check.Version, check.CheckName)
		if !dryRun {
			err := PutCheck(check, org, insightsToken, host, pushRegoVersion)
			if err != nil {
				return err
			}
		}
	}
	for _, instance := range results.InstanceInsert {
		logrus.Infof("Adding instance: %s for OPA policy %s", instance.InstanceName, instance.CheckName)
		if !dryRun {
			err := PutInstance(instance, org, insightsToken, host)
			if err != nil {
				return err
			}
		}
	}
	for _, instance := range results.InstanceUpdate {
		logrus.Infof("Updating instance: %s for OPA policy %s", instance.InstanceName, instance.CheckName)
		if !dryRun {
			err := PutInstance(instance, org, insightsToken, host)
			if err != nil {
				return err
			}
		}
	}
	logrus.Debugln("Done pushing OPA policies")
	return nil
}

// PushExternalOPAChecks pushes external OPA checks to Insights.
func PushExternalOPAChecks(filePath, org, insightsToken string, headers []string, host string, deleteMissing, dryRun bool, pushRegoVersion string) error {
	logrus.Debugln("Pushing external OPA policies")
	_, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}

	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	checks, err := getExternalChecksFromFile(b, headers)
	if err != nil {
		return fmt.Errorf("error getting remote checks: %w", err)
	}

	results, err := CompareChecks(filePath, org, insightsToken, host, checks, deleteMissing)
	if err != nil {
		return fmt.Errorf("error comparing checks: %w", err)
	}
	for _, instance := range results.InstanceDelete {
		logrus.Infof("Deleting instance: %s for OPA policy %s", instance.InstanceName, instance.CheckName)
		if !dryRun {
			err := DeleteInstance(instance, org, insightsToken, host)
			if err != nil {
				return fmt.Errorf("error deleting instance: %w", err)
			}
		}
	}
	for _, check := range results.CheckDelete {
		logrus.Infof("Deleting OPA policy: %s", check.CheckName)
		if !dryRun {
			err := DeleteCheck(check, org, insightsToken, host)
			if err != nil {
				return fmt.Errorf("error deleting check: %w", err)
			}
		}
	}
	for _, check := range results.CheckInsert {
		logrus.Infof("Adding v%.0f OPA policy: %s", check.Version, check.CheckName)
		if !dryRun {
			err := PutCheck(check, org, insightsToken, host, pushRegoVersion)
			if err != nil {
				return fmt.Errorf("error adding check: %w", err)
			}
		}
	}
	for _, check := range results.CheckUpdate {
		logrus.Infof("Updating v%.0f OPA policy: %s", check.Version, check.CheckName)
		if !dryRun {
			err := PutCheck(check, org, insightsToken, host, pushRegoVersion)
			if err != nil {
				return fmt.Errorf("error updating check: %w", err)
			}
		}
	}
	for _, instance := range results.InstanceInsert {
		logrus.Infof("Adding instance: %s for OPA policy %s", instance.InstanceName, instance.CheckName)
		if !dryRun {
			err := PutInstance(instance, org, insightsToken, host)
			if err != nil {
				return fmt.Errorf("error adding instance: %w", err)
			}
		}
	}
	for _, instance := range results.InstanceUpdate {
		logrus.Infof("Updating instance: %s for OPA policy %s", instance.InstanceName, instance.CheckName)
		if !dryRun {
			err := PutInstance(instance, org, insightsToken, host)
			if err != nil {
				return fmt.Errorf("error updating instance: %w", err)
			}
		}
	}
	logrus.Debugln("Done pushing external OPA policies")
	return nil
}

type externalSource struct {
	ExternalSources []externalSourceItem `yaml:"externalSources"`
}

type externalSourceItem struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	URL         string `yaml:"url"`
	Enabled     *bool  `yaml:"enabled"`
}

// getExternalChecksFromFile reads the external sources file and fetches the OPA checks from them
func getExternalChecksFromFile(fileContent []byte, headers []string) ([]models.CustomCheckModel, error) {
	var externalSources externalSource
	err := yaml.Unmarshal(fileContent, &externalSources)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling yaml: %w", err)
	}

	if len(externalSources.ExternalSources) == 0 {
		return []models.CustomCheckModel{}, nil
	}

	var checks []models.CustomCheckModel
	for _, source := range externalSources.ExternalSources {
		logrus.Debugf("getting checks from %s", source.URL)
		resp, err := req.C().R().SetHeaders(formatHeaders(headers)).Get(source.URL)
		if err != nil {
			return nil, fmt.Errorf("error getting remote checks: %w", err)
		}
		if resp.Response.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("error getting remote checks: invalid response code (%v, expected 200)", resp.Response.StatusCode)
		}
		rego, err := resp.ToString()
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling remote checks: %w", err)
		}
		checks = append(checks, models.CustomCheckModel{
			CheckName:   source.Name,
			Description: source.Description,
			Rego:        rego,
			Version:     2.0,
			Disabled:    utils.InvertBoolPointer(source.Enabled),
		})
	}
	return checks, nil
}

func formatHeaders(headers []string) map[string]string {
	r := map[string]string{}
	for _, s := range headers {
		parts := strings.Split(s, ":")
		if len(parts) != 2 {
			logrus.Warnf("invalid header: '%s' - should be formatted as 'key:value'", s)
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		r[key] = value
	}

	logrus.Debugf("headers: %v", r)
	return r
}
