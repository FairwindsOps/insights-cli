package opa

import (
	"errors"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/fairwindsops/insights-plugins/opa/pkg/opa"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"gopkg.in/yaml.v3"

	"github.com/fairwindsops/insights-cli/pkg/directory"
	"github.com/fairwindsops/insights-cli/pkg/insights"
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
func CompareChecks(folder, org, token, hostName string, gitops bool) (CompareResults, error) {
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
	apiChecks, err := insights.GetChecks(org, token, hostName)
	if err != nil {
		logrus.Error("Error getting checks from Insights")
		return results, err
	}
	if !gitops {
		apiChecks = funk.Filter(apiChecks, func(c opa.OPACustomCheck) bool {
			return len(funk.Filter(fileChecks, func(fc models.CustomCheckModel) bool {
				return fc.CheckName == c.Name
			}).([]models.CustomCheckModel)) > 0
		}).([]opa.OPACustomCheck)
	}
	var apiInstances []opa.CheckSetting
	// TODO replace with org wide get.
	for _, check := range apiChecks {
		newInstances, err := insights.GetInstances(org, check.Name, token, hostName)
		if err != nil {
			logrus.Error("Error getting instances from Insights")
			return results, err
		}
		apiInstances = append(apiInstances, newInstances...)
	}
	results = compareChecks(fileChecks, apiChecks, apiInstances)
	return results, nil
}

func compareChecks(fileChecks []models.CustomCheckModel, apiChecks []opa.OPACustomCheck, apiInstances []opa.CheckSetting) CompareResults {
	var results CompareResults
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
	results.CheckDelete = append(results.CheckDelete, diffChecks...)

	for _, instance := range apiInstances {
		for _, deletedCheck := range results.CheckDelete {
			if instance.CheckName == deletedCheck.CheckName {
				results.InstanceDelete = append(results.InstanceDelete, models.CustomCheckInstanceModel{
					CheckName:    instance.CheckName,
					InstanceName: instance.AdditionalData.Name,
				})
				break
			}
		}
	}

	for _, fileCheck := range fileChecks {
		found := false
		for _, check := range apiChecks {
			if check.Name == fileCheck.CheckName {
				found = true
				if check.Rego != fileCheck.Rego ||
					notEqual(check.Category, fileCheck.Output.Category) ||
					notEqual(check.Remediation, fileCheck.Output.Remediation) ||
					notEqual(check.Severity, fileCheck.Output.Severity) ||
					notEqual(check.Title, fileCheck.Output.Title) {

					results.CheckUpdate = append(results.CheckUpdate, fileCheck)
				}
				break
			}
		}
		if !found {
			results.CheckInsert = append(results.CheckInsert, fileCheck)
		}
		instances := funk.Filter(apiInstances, func(instance opa.CheckSetting) bool {
			return instance.CheckName == fileCheck.CheckName
		}).([]opa.CheckSetting)
		for _, fileInstance := range fileCheck.Instances {
			found := false
			for _, instance := range instances {
				if fileInstance.InstanceName == instance.AdditionalData.Name {
					found = true
					// TODO check for changed properties
					break
				}
			}
			if !found {
				results.InstanceInsert = append(results.InstanceInsert, fileInstance)
			}
		}
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
		results.InstanceDelete = append(results.InstanceDelete, diffInstances...)
	}
	return results
}

func notEqual(i1 interface{}, i2 interface{}) bool {
	if (i1 == nil) != (i2 == nil) {
		return true
	}
	if i1 == nil {
		return false
	}
	return &i1 != &i2
}

func getChecksFromFiles(files map[string][]string) ([]models.CustomCheckModel, error) {
	var checks []models.CustomCheckModel
	for checkName, checkFiles := range files {
		var check models.CustomCheckModel
		var instances []models.CustomCheckInstanceModel
		for _, filePath := range checkFiles {
			fileContents, err := ioutil.ReadFile(filePath)
			if err != nil {
				logrus.Error(err, "Error reading file", filePath)
				return nil, err
			}
			if strings.HasPrefix(filepath.Base(filePath), "policy") {
				extension := filepath.Ext(filePath)
				if extension == ".rego" {
					check.Rego = string(fileContents)
				} else if extension == ".yaml" {

					err = yaml.Unmarshal(fileContents, &check)
					if err != nil {
						logrus.Error(err, "Error Unmarshalling check YAML", filePath)
						return nil, err
					}
				} else {
					logrus.Errorf("policy file is not a rego or yaml file: %s", extension)
					return nil, errors.New("Policy file must be rego or yaml")
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
		check.Instances = instances
		checks = append(checks, check)
	}
	return checks, nil
}
