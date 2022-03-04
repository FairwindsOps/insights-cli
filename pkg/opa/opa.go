// Copyright 2020 FairwindsOps Inc
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
	"io/ioutil"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/fairwindsops/insights-plugins/plugins/opa/pkg/opa"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"github.com/xlab/treeprint"
	"gopkg.in/yaml.v3"

	"github.com/fairwindsops/insights-cli/pkg/directory"
	"github.com/fairwindsops/insights-cli/pkg/models"
)

// CompareResults shows the results of a comparison between what's present in the API and what's in a folder
type CompareResults struct {
	CheckInsert    []models.CustomCheckModel
	CheckUpdate    []models.CustomCheckModel
	CheckDelete    []models.CustomCheckModel
	InstanceInsert []models.CustomCheckInstanceModel
	InstanceUpdate []models.CustomCheckInstanceModel
	InstanceDelete []models.CustomCheckInstanceModel
}

// CompareChecks compares a folder vs the checks returned by the API.
func CompareChecks(folder, org, token, hostName string, fullsync bool) (CompareResults, error) {
	var results CompareResults
	files, err := directory.ScanFolder(folder)
	if err != nil {
		logrus.Error("Error scanning directory")
		return results, err
	}

	fileChecks, err := getChecksFromFiles(files)
	if err != nil {
		logrus.Error("Error Reading checks from files")
		return results, err
	}
	apiChecks, err := GetChecks(org, token, hostName)
	if err != nil {
		logrus.Error("Error getting checks from Insights")
		return results, err
	}
	if !fullsync {
		apiChecks = funk.Filter(apiChecks, func(c opa.OPACustomCheck) bool {
			return len(funk.Filter(fileChecks, func(fc models.CustomCheckModel) bool {
				return fc.CheckName == c.Name
			}).([]models.CustomCheckModel)) > 0
		}).([]opa.OPACustomCheck)
	}
	fileChecks = funk.Filter(fileChecks, func(fc models.CustomCheckModel) bool {
		return fc.Rego != ""
	}).([]models.CustomCheckModel)
	var apiInstances []opa.CheckSetting
	// TODO replace with org wide get.
	for _, check := range apiChecks {
		newInstances, err := GetInstances(org, check.Name, token, hostName)
		if err != nil {
			logrus.Error("Error getting instances from Insights")
			return results, err
		}
		apiInstances = append(apiInstances, newInstances...)
	}
	results = compareChecks(fileChecks, apiChecks, apiInstances)
	return results, nil
}

func instanceMatchesName(name string) func(opa.CheckSetting) bool {
	return func(instance opa.CheckSetting) bool {
		return instance.CheckName == name
	}
}

func getMissingChecks(apiChecks []opa.OPACustomCheck, fileChecks []models.CustomCheckModel) []models.CustomCheckModel {
	diff := funk.Subtract(
		funk.Map(apiChecks, func(c opa.OPACustomCheck) string {
			return c.Name
		}),
		funk.Map(fileChecks, func(c models.CustomCheckModel) string {
			return c.CheckName
		}))
	diffChecks := funk.Map(diff, func(s string) models.CustomCheckModel {
		return models.CustomCheckModel{
			CheckName: s,
		}
	}).([]models.CustomCheckModel)
	return diffChecks
}

func compareChecks(fileChecks []models.CustomCheckModel, apiChecks []opa.OPACustomCheck, apiInstances []opa.CheckSetting) CompareResults {
	var results CompareResults
	results.CheckDelete = append(results.CheckDelete, getMissingChecks(apiChecks, fileChecks)...)
	for _, deletedCheck := range results.CheckDelete {
		for _, instance := range funk.Filter(apiInstances, instanceMatchesName(deletedCheck.CheckName)).([]opa.CheckSetting) {
			results.InstanceDelete = append(results.InstanceDelete, models.CustomCheckInstanceModel{
				CheckName:    instance.CheckName,
				InstanceName: instance.AdditionalData.Name,
			})
		}
	}

	for _, fileCheck := range fileChecks {
		found := false
		for _, check := range apiChecks {
			if check.Name == fileCheck.CheckName {
				found = true
				if checksDoNotMatch(fileCheck, check) {
					results.CheckUpdate = append(results.CheckUpdate, fileCheck)
				}
				break
			}
		}
		if !found {
			results.CheckInsert = append(results.CheckInsert, fileCheck)
		}
		instances := funk.Filter(apiInstances, instanceMatchesName(fileCheck.CheckName)).([]opa.CheckSetting)
		for _, fileInstance := range fileCheck.Instances {
			found := false
			for _, instance := range instances {
				if fileInstance.InstanceName == instance.AdditionalData.Name {
					found = true
					if instancesDoNotMatch(fileInstance, instance) {
						results.InstanceUpdate = append(results.InstanceUpdate, fileInstance)
					}
					break
				}
			}
			if !found {
				results.InstanceInsert = append(results.InstanceInsert, fileInstance)
			}
		}
		results.InstanceDelete = append(results.InstanceDelete, getDifferenceInstances(instances, fileCheck)...)
	}
	return results
}

func getDifferenceInstances(instances []opa.CheckSetting, fileCheck models.CustomCheckModel) []models.CustomCheckInstanceModel {
	diff := funk.Subtract(
		funk.Map(instances, func(i opa.CheckSetting) string {
			return i.AdditionalData.Name
		}),
		funk.Map(fileCheck.Instances, func(i models.CustomCheckInstanceModel) string {
			return i.InstanceName
		}))
	diffInstances := funk.Map(diff, func(s string) models.CustomCheckInstanceModel {
		return models.CustomCheckInstanceModel{
			InstanceName: s,
			CheckName:    fileCheck.CheckName,
		}
	}).([]models.CustomCheckInstanceModel)
	return diffInstances
}

func instancesDoNotMatch(fileInstance models.CustomCheckInstanceModel, apiInstance opa.CheckSetting) bool {
	// TODO check for changed clusters/run environments
	return !reflect.DeepEqual(apiInstance.AdditionalData.Parameters, fileInstance.Parameters) ||
		targetsNotEqual(apiInstance.Targets, fileInstance.Targets) ||
		!reflect.DeepEqual(apiInstance.AdditionalData.Output.Category, fileInstance.Output.Category) ||
		!reflect.DeepEqual(apiInstance.AdditionalData.Output.Remediation, fileInstance.Output.Remediation) ||
		!reflect.DeepEqual(apiInstance.AdditionalData.Output.Severity, fileInstance.Output.Severity) ||
		!reflect.DeepEqual(apiInstance.AdditionalData.Output.Title, fileInstance.Output.Title)
}

func checksDoNotMatch(fileCheck models.CustomCheckModel, apiCheck opa.OPACustomCheck) bool {
	return apiCheck.Rego != fileCheck.Rego ||
		!reflect.DeepEqual(apiCheck.Category, fileCheck.Output.Category) ||
		!reflect.DeepEqual(apiCheck.Remediation, fileCheck.Output.Remediation) ||
		!reflect.DeepEqual(apiCheck.Severity, fileCheck.Output.Severity) ||
		!reflect.DeepEqual(apiCheck.Title, fileCheck.Output.Title)
}

func targetsNotEqual(apiTarget []string, fileTarget []models.KubernetesTarget) bool {
	var fileStringTargets []string
	for _, target := range fileTarget {
		for _, kind := range target.Kinds {
			for _, group := range target.APIGroups {
				fileStringTargets = append(fileStringTargets, fmt.Sprintf("%s/%s", group, kind))
			}
		}
	}
	sort.Strings(fileStringTargets)
	sort.Strings(apiTarget)
	return !reflect.DeepEqual(apiTarget, fileStringTargets)
}

func getChecksFromFiles(files map[string][]string) ([]models.CustomCheckModel, error) {
	var checks []models.CustomCheckModel
	for checkName, checkFiles := range files {
		var check models.CustomCheckModel
		var instances []models.CustomCheckInstanceModel
		var onlyRegoFormat bool
		check.Version = 1.0
		for _, filePath := range checkFiles {
			fileContents, err := ioutil.ReadFile(filePath)
			if err != nil {
				logrus.Error(err, "Error reading file", filePath)
				return nil, err
			}
			if strings.HasPrefix(filepath.Base(filePath), "policy.") {
				extension := filepath.Ext(filePath)
				if extension == ".rego" {
					onlyRegoFormat = true
					check.Rego = string(fileContents)
				} else if extension == ".yaml" {
					err = yaml.Unmarshal(fileContents, &check)
					if err != nil {
						logrus.Error(err, "Error Unmarshalling check YAML", filePath)
						return nil, err
					}
				} else {
					logrus.Errorf("policy file is not a rego or yaml file: %s", extension)
					return nil, errors.New("policy file must be rego or yaml")
				}
			} else {
				var instance models.CustomCheckInstanceModel
				err = yaml.Unmarshal(fileContents, &instance)

				if err != nil {
					logrus.Error(err, "Error Unmarshalling instance YAML", filePath)
					return nil, err
				}
				instance.InstanceName = strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
				instance.CheckName = checkName
				instances = append(instances, instance)

			}
		}
		check.CheckName = checkName
		if len(instances) == 0 && onlyRegoFormat {
			check.Version = 2.0
		} else {
			check.Instances = instances
		}
		checks = append(checks, check)
	}
	return checks, nil
}

// BuildChecksTree builds the tree for OPA checks
func BuildChecksTree(org, token, hostName string, tree treeprint.Tree) error {
	checks, err := GetChecks(org, token, hostName)
	if err != nil {
		logrus.Errorf("Unable to get checks from insights: %v", err)
		return err
	}
	opaBranch := tree.AddBranch("opa")
	for _, check := range checks {
		branch := opaBranch.AddBranch(fmt.Sprintf("%s (v%.0f)", check.Name, check.Version))
		instances, err := GetInstances(org, check.Name, token, hostName)
		if err != nil {
			logrus.Errorf("Unable to get instances from insights: %v", err)
			return err
		}
		for _, instance := range instances {
			branch.AddNode(instance.AdditionalData.Name)
		}
	}
	return nil
}
