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
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/fairwindsops/insights-plugins/plugins/opa/pkg/opa"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
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
func CompareChecks(folder, org, token, hostName string, deleteMissing bool) (CompareResults, error) {
	var results CompareResults
	files, err := directory.ScanOPAFolder(folder)
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
	if !deleteMissing {
		apiChecks = lo.Filter(apiChecks, func(c opa.OPACustomCheck, _ int) bool {
			return len(lo.Filter(fileChecks, func(fc models.CustomCheckModel, _ int) bool {
				return fc.CheckName == c.Name
			})) > 0
		})
	}
	fileChecks = lo.Filter(fileChecks, func(fc models.CustomCheckModel, _ int) bool {
		return fc.Rego != ""
	})
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

func instanceMatchesName(name string) func(opa.CheckSetting, int) bool {
	return func(instance opa.CheckSetting, _ int) bool {
		return instance.CheckName == name
	}
}

func getMissingChecks(apiChecks []opa.OPACustomCheck, fileChecks []models.CustomCheckModel) []models.CustomCheckModel {
	left, _ := lo.Difference(
		lo.Map(apiChecks, func(c opa.OPACustomCheck, _ int) string {
			return c.Name
		}),
		lo.Map(fileChecks, func(c models.CustomCheckModel, _ int) string {
			return c.CheckName
		}))
	diffChecks := lo.Map(left, func(s string, _ int) models.CustomCheckModel {
		return models.CustomCheckModel{
			CheckName: s,
		}
	})
	return diffChecks
}

func compareChecks(fileChecks []models.CustomCheckModel, apiChecks []opa.OPACustomCheck, apiInstances []opa.CheckSetting) CompareResults {
	var results CompareResults
	results.CheckDelete = append(results.CheckDelete, getMissingChecks(apiChecks, fileChecks)...)
	for _, deletedCheck := range results.CheckDelete {
		for _, instance := range lo.Filter(apiInstances, instanceMatchesName(deletedCheck.CheckName)) {
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
		instances := lo.Filter(apiInstances, instanceMatchesName(fileCheck.CheckName))
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
	left, _ := lo.Difference(
		lo.Map(instances, func(i opa.CheckSetting, _ int) string {
			return i.AdditionalData.Name
		}),
		lo.Map(fileCheck.Instances, func(i models.CustomCheckInstanceModel, _ int) string {
			return i.InstanceName
		}))
	diffInstances := lo.Map(left, func(s string, _ int) models.CustomCheckInstanceModel {
		return models.CustomCheckInstanceModel{
			InstanceName: s,
			CheckName:    fileCheck.CheckName,
		}
	})
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
		var onlyRegoFormat bool = true
		check.Version = 1.0
		for _, filePath := range checkFiles {
			fileContents, err := os.ReadFile(filePath)
			if err != nil {
				logrus.Error(err, "Error reading file", filePath)
				return nil, err
			}
			extension := strings.ToLower(filepath.Ext(filePath))
			if extension == ".rego" {
				logrus.Debugf("using content of file %s as rego for OPA policy %s\n", filePath, checkName)
				check.Rego = string(fileContents)
				continue
			}
			if strings.ToLower(filepath.Base(filePath)) == "policy.yaml" {
				onlyRegoFormat = false
				logrus.Debugf("using content of file %s as instance and rego for V1 OPA policy %s\n", filePath, checkName)
				err = yaml.Unmarshal(fileContents, &check)
				if err != nil {
					logrus.Error(err, "Error Unmarshalling check YAML", filePath)
					return nil, err
				}
				continue
			}
			if extension == ".yaml" {
				logrus.Debugf("using content of file %s as instance for V1 OPA policy %s\n", filePath, checkName)
				onlyRegoFormat = false
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
		logrus.Debugf("processed files %s as v%.1f OPA policy %s\n", checkFiles, check.Version, check.CheckName)
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
