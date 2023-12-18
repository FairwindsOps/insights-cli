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
	"net/http"
	"os"

	"github.com/fairwindsops/insights-cli/pkg/models"
	"github.com/fairwindsops/insights-cli/pkg/utils"
	"github.com/fairwindsops/insights-cli/pkg/version"
	opaPlugin "github.com/fairwindsops/insights-plugins/plugins/opa/pkg/opa"
	"github.com/imroc/req"
	"github.com/sirupsen/logrus"
)

const opaURLFormat = "%s/v0/organizations/%s/opa/customChecks"

const opaCheckURLFormat = opaURLFormat + "/%s"
const opaPutCheckURLFormat = opaCheckURLFormat + "?version=%.1f"
const opaCheckInstancesURLFormat = opaCheckURLFormat + "/instances"

const opaInstanceURLFormat = opaCheckInstancesURLFormat + "/%s"

// GetChecks queries Fairwinds Insights to retrieve all of the Checks for an organization
func GetChecks(org, token, hostName string) ([]opaPlugin.OPACustomCheck, error) {
	url := fmt.Sprintf(opaURLFormat, hostName, org)
	logrus.Debugf("OPA URL: %s", url)
	resp, err := req.Get(url, utils.GetHeaders(version.GetVersion(), token))
	if err != nil {
		return nil, err
	}
	var checks []opaPlugin.OPACustomCheck
	if resp.Response().StatusCode != http.StatusOK {
		logrus.Errorf("GetChecks: Invalid response code: %s %v", string(resp.Bytes()), resp.Response().StatusCode)
		return nil, errors.New("GetChecks: invalid response code")
	}
	err = resp.ToJSON(&checks)
	if err != nil {
		return nil, err
	}
	return checks, nil
}

// GetInstances queries Fairwinds Insights to retrieve all of the instances for a given check
func GetInstances(org, checkName, token, hostName string) ([]opaPlugin.CheckSetting, error) {
	url := fmt.Sprintf(opaCheckInstancesURLFormat, hostName, org, checkName)
	resp, err := req.Get(url, utils.GetHeaders(version.GetVersion(), token))
	if err != nil {
		return nil, err
	}
	if resp.Response().StatusCode != http.StatusOK {
		logrus.Errorf("GetInstances: Invalid response code: %s %v", string(resp.Bytes()), resp.Response().StatusCode)
		return nil, errors.New("GetInstances: invalid response code")
	}
	var instances []opaPlugin.CheckSetting
	err = resp.ToJSON(&instances)
	if err != nil {
		return nil, err
	}
	return instances, nil
}

// DeleteCheck deletes an OPA Check from Fairwinds Insights
func DeleteCheck(check models.CustomCheckModel, org, token, hostName string) error {
	url := fmt.Sprintf(opaCheckURLFormat, hostName, org, check.CheckName)
	resp, err := req.Delete(url, utils.GetHeaders(version.GetVersion(), token))
	if err != nil {
		return err
	}
	if resp.Response().StatusCode != http.StatusOK {
		logrus.Errorf("DeleteCheck: Invalid response code: %s %v", string(resp.Bytes()), resp.Response().StatusCode)
		return errors.New("DeleteCheck: invalid response code")
	}
	return nil
}

// PutCheck upserts an OPA Check to Fairwinds Insights
func PutCheck(check models.CustomCheckModel, org, token, hostName string) error {
	url := fmt.Sprintf(opaPutCheckURLFormat, hostName, org, check.CheckName, check.Version)
	resp, err := req.Put(url, utils.GetHeaders(version.GetVersion(), token), req.BodyJSON(&check))
	if err != nil {
		return err
	}
	if resp.Response().StatusCode != http.StatusOK {
		logrus.Errorf("PutCheck: Invalid response code: %s %v", string(resp.Bytes()), resp.Response().StatusCode)
		return errors.New("PutCheck: invalid response code")
	}
	return nil
}

// DeleteInstance deletes an Instance from Fairwinds Insights
func DeleteInstance(instance models.CustomCheckInstanceModel, org, token, hostName string) error {
	url := fmt.Sprintf(opaInstanceURLFormat, hostName, org, instance.CheckName, instance.InstanceName)
	resp, err := req.Delete(url, utils.GetHeaders(version.GetVersion(), token))
	if err != nil {
		return err
	}
	if resp.Response().StatusCode != http.StatusOK {
		logrus.Errorf("DeleteInstance: Invalid response code: %s %v", string(resp.Bytes()), resp.Response().StatusCode)
		return errors.New("DeleteInstance: invalid response code")
	}
	return nil
}

// PutInstance upserts an Instance to Fairwinds Insights
func PutInstance(instance models.CustomCheckInstanceModel, org, token, hostName string) error {
	url := fmt.Sprintf(opaInstanceURLFormat, hostName, org, instance.CheckName, instance.InstanceName)
	resp, err := req.Put(url, utils.GetHeaders(version.GetVersion(), token), req.BodyJSON(&instance))
	if err != nil {
		return err
	}
	if resp.Response().StatusCode != http.StatusOK {
		logrus.Errorf("PutInstance: Invalid response code: %s %v", string(resp.Bytes()), resp.Response().StatusCode)
		return errors.New("PutInstance: invalid response code")
	}
	return nil
}

// PushOPAChecks pushes OPA checks to Insights.
func PushOPAChecks(pushDir, org, insightsToken, host string, deleteMissing, dryrun bool) error {
	logrus.Debugln("Pushing OPA policies")
	_, err := os.Stat(pushDir)
	if err != nil {
		return err
	}
	results, err := CompareChecks(pushDir, org, insightsToken, host, deleteMissing)
	if err != nil {
		return err
	}
	for _, instance := range results.InstanceDelete {
		logrus.Infof("Deleting instance: %s for OPA policy %s", instance.InstanceName, instance.CheckName)
		if !dryrun {
			err := DeleteInstance(instance, org, insightsToken, host)
			if err != nil {
				return err
			}
		}
	}
	for _, check := range results.CheckDelete {
		logrus.Infof("Deleting OPA policy: %s", check.CheckName)
		if !dryrun {
			err := DeleteCheck(check, org, insightsToken, host)
			if err != nil {
				return err
			}
		}
	}
	for _, check := range results.CheckInsert {
		logrus.Infof("Adding v%.0f OPA policy: %s", check.Version, check.CheckName)
		if !dryrun {
			err := PutCheck(check, org, insightsToken, host)
			if err != nil {
				return err
			}
		}
	}
	for _, check := range results.CheckUpdate {
		logrus.Infof("Updating v%.0f OPA policy: %s", check.Version, check.CheckName)
		if !dryrun {
			err := PutCheck(check, org, insightsToken, host)
			if err != nil {
				return err
			}
		}
	}
	for _, instance := range results.InstanceInsert {
		logrus.Infof("Adding instance: %s for OPA policy %s", instance.InstanceName, instance.CheckName)
		if !dryrun {
			err := PutInstance(instance, org, insightsToken, host)
			if err != nil {
				return err
			}
		}
	}
	for _, instance := range results.InstanceUpdate {
		logrus.Infof("Updating instance: %s for OPA policy %s", instance.InstanceName, instance.CheckName)
		if !dryrun {
			err := PutInstance(instance, org, insightsToken, host)
			if err != nil {
				return err
			}
		}
	}
	logrus.Debugln("Done pushing OPA policies")
	return nil
}
