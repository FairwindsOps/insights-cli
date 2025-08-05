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

package rules

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/xlab/treeprint"

	"github.com/fairwindsops/insights-cli/pkg/directory"
	"github.com/imroc/req/v3"
)

const rulesURLFormat = "/v0/organizations/%s/rules"
const rulesURLFormatCreate = "/v0/organizations/%s/rules/create"
const rulesURLFormatUpdateDelete = "/v0/organizations/%s/rules/%d"

// Rule is the struct to hold the information for a rule
type Rule struct {
	ID          int
	Cluster     string
	Name        string
	Description string
	Context     string
	ReportType  string `json:"reportType" yaml:"reportType"`
	Repository  string
	Action      string
}

// CompareResults holds the rules for inserting, updating, and deleting
type CompareResults struct {
	RuleInsert []Rule
	RuleUpdate []Rule
	RuleDelete []Rule
}

// getRules queries Fairwinds Insights to retrieve all of the Rules for an organization
func getRules(client *req.Client, org string) ([]Rule, error) {
	url := fmt.Sprintf(rulesURLFormat, org)
	logrus.Debugf("Rules URL: %s", url)
	resp, err := client.R().SetHeaders(getHeaders()).Get(url)
	if err != nil {
		logrus.Errorf("Unable to get rules from insights: %v", err)
		return nil, err
	}
	var rules []Rule
	if resp.Response.StatusCode != http.StatusOK {
		logrus.Errorf("getRules: invalid response code: %s %v", string(resp.Bytes()), resp.Response.StatusCode)
		return nil, errors.New("getRules: invalid response code")
	}
	err = resp.Unmarshal(&rules)
	if err != nil {
		logrus.Errorf("Unable to convert response to json for rules: %v", err)
		return nil, err
	}
	return rules, nil
}

// insertRule adds a new rule
func insertRule(client *req.Client, org string, rule Rule) error {
	url := fmt.Sprintf(rulesURLFormatCreate, org)
	resp, err := client.R().SetHeaders(getHeaders()).SetBody(&rule).Post(url)
	if err != nil {
		logrus.Errorf("Unable to add rule %s to insights: %v", rule.Name, err)
		return err
	}
	if resp.Response.StatusCode != http.StatusOK {
		logrus.Errorf("insertRule: invalid response code: %s %v", string(resp.Bytes()), resp.Response.StatusCode)
		return errors.New("insertRule: invalid response code")
	}
	return nil
}

// updateRule updates an existing rule
func updateRule(client *req.Client, org string, rule Rule) error {
	url := fmt.Sprintf(rulesURLFormatUpdateDelete, org, rule.ID)
	resp, err := client.R().SetHeaders(getHeaders()).SetBody(&rule).Post(url)
	if err != nil {
		logrus.Errorf("Unable to update rule %s to insights: %v", rule.Name, err)
		return err
	}
	if resp.Response.StatusCode != http.StatusOK {
		logrus.Errorf("updateRule: invalid response code: %s %v", string(resp.Bytes()), resp.Response.StatusCode)
		return errors.New("updateRule: invalid response code")
	}
	return nil
}

// deleteRule deletes an existing rule
func deleteRule(client *req.Client, org string, rule Rule) error {
	url := fmt.Sprintf(rulesURLFormatUpdateDelete, org, rule.ID)
	resp, err := client.R().SetHeaders(getHeaders()).Delete(url)
	if err != nil {
		logrus.Errorf("Unable to delete rule %s from insights: %v", rule.Name, err)
		return err
	}
	if resp.Response.StatusCode != http.StatusOK {
		logrus.Errorf("deleteRule: Invalid response code: %s %v", string(resp.Bytes()), resp.Response.StatusCode)
		return errors.New("deleteRule: invalid response code")
	}
	return nil
}

// AddRulesBranch builds a tree for rules
func AddRulesBranch(client *req.Client, org string, tree treeprint.Tree) error {
	rules, err := getRules(client, org)
	if err != nil {
		logrus.Errorf("Unable to get rules from insights: %v", err)
		return err
	}
	rulesBranch := tree.AddBranch("rules")
	for _, rule := range rules {
		rulesNode := rulesBranch.AddBranch(rule.Name)
		if rule.Cluster != "" {
			value := fmt.Sprintf("Cluster: %s", rule.Cluster)
			rulesNode.AddNode(value)
		}
		if rule.Context != "" {
			value := fmt.Sprintf("Repository: %s", rule.Repository)
			rulesNode.AddNode(value)
		}
		if rule.Repository != "" {
			value := fmt.Sprintf("Repository: %s", rule.Repository)
			rulesNode.AddNode(value)
		}
		if rule.ReportType != "" {
			value := fmt.Sprintf("Report Type: %s", rule.ReportType)
			rulesNode.AddNode(value)
		}
	}
	return nil
}

func getRulesFromFiles(files map[string][]string) ([]Rule, error) {
	var rules []Rule
	for _, ruleFiles := range files {
		for _, filePath := range ruleFiles {
			fileContents, err := os.ReadFile(filePath)
			if err != nil {
				logrus.Error(err, "Error reading file", filePath)
				return nil, err
			}
			var rule Rule
			err = yaml.Unmarshal(fileContents, &rule)
			if err != nil {
				logrus.Error(err, "Error unmarshalling check YAML", filePath)
				return nil, err
			}
			if rule.Name == "" {
				return nil, fmt.Errorf("Rule name is empty in file: %s", filePath)
			}
			if rule.Action == "" {
				return nil, fmt.Errorf("Rule action is empty in file: %s", filePath)
			}
			rules = append(rules, rule)
		}
	}
	return rules, nil
}

func ruleNeedsUpdate(fileRule, existingRule Rule) bool {
	if fileRule.Description != existingRule.Description {
		return true
	}
	if fileRule.Context != existingRule.Context {
		return true
	}
	if fileRule.Cluster != existingRule.Cluster {
		return true
	}
	if fileRule.ReportType != existingRule.ReportType {
		return true
	}
	if fileRule.Repository != existingRule.Repository {
		return true
	}
	if fileRule.Action != existingRule.Action {
		return true
	}
	return false
}

func getRuleDifferences(fileRules, existingRules []Rule) CompareResults {
	mappedFileRules := lo.KeyBy(fileRules, func(i Rule) string { return i.Name })
	mappedExistingRules := lo.KeyBy(existingRules, func(i Rule) string { return i.Name })
	var results CompareResults
	for ruleName, fileRule := range mappedFileRules {
		if existingRule, ok := mappedExistingRules[ruleName]; ok {
			if ruleNeedsUpdate(fileRule, existingRule) {
				fileRule.ID = existingRule.ID
				results.RuleUpdate = append(results.RuleUpdate, fileRule)
			}
		} else {
			results.RuleInsert = append(results.RuleInsert, fileRule)
		}
	}

	for ruleName, existingRule := range mappedExistingRules {
		if _, ok := mappedFileRules[ruleName]; !ok {
			results.RuleDelete = append(results.RuleDelete, existingRule)
		}
	}
	return results
}

// compareRules compares a folder vs the rules returned by the API.
func compareRules(client *req.Client, folder, org string) (CompareResults, error) {
	var results CompareResults
	files, err := directory.ScanFolder(folder)
	if err != nil {
		logrus.Error("Error scanning directory")
		return results, err
	}

	fileRules, err := getRulesFromFiles(files)
	if err != nil {
		logrus.Error("Error reading checks from files")
		return results, err
	}
	existingRules, err := getRules(client, org)
	if err != nil {
		logrus.Error("Error during API call")
		return results, err
	}

	results = getRuleDifferences(fileRules, existingRules)
	return results, nil
}

// PushRules pushes automation rules to insights
func PushRules(client *req.Client, pushDir, org string, deleteMissing, dryrun bool) error {
	logrus.Debugln("Pushing automation rules")
	_, err := os.Stat(pushDir)
	if err != nil {
		return err
	}
	results, err := compareRules(client, pushDir, org)
	if err != nil {
		logrus.Errorf("unable to compare and push rules to Insights: %v", err)
		return err
	}

	for _, ruleForInsert := range results.RuleInsert {
		logrus.Infof("Adding automation rule: %s", ruleForInsert.Name)
		if !dryrun {
			err = insertRule(client, org, ruleForInsert)
			if err != nil {
				logrus.Errorf("Error while adding rule %s to insights: %v", ruleForInsert.Name, err)
				return err
			}
		}
	}

	for _, ruleForUpdate := range results.RuleUpdate {
		logrus.Infof("Updating automation rule: %s", ruleForUpdate.Name)
		if !dryrun {
			err = updateRule(client, org, ruleForUpdate)
			if err != nil {
				logrus.Errorf("Error while updating rule %s to insights: %v", ruleForUpdate.Name, err)
				return err
			}
		}
	}

	if deleteMissing {
		for _, ruleForDelete := range results.RuleDelete {
			logrus.Infof("Deleting automation rule: %s", ruleForDelete.Name)
			if !dryrun {
				err = deleteRule(client, org, ruleForDelete)
				if err != nil {
					logrus.Errorf("Error while deleting rule %s from insights: %v", ruleForDelete.Name, err)
					return err
				}
			}
		}
	}
	logrus.Debugln("Done pushing automation rules")
	return nil
}

func getHeaders() map[string]string {
	return map[string]string{
		"Accept":       "application/json",
		"Content-Type": "application/json",
	}
}
