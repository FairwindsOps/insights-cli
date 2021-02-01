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

package rules

import (
	"errors"
	"fmt"
	"github.com/fairwindsops/insights-cli/pkg/directory"
	"github.com/imroc/req"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"github.com/xlab/treeprint"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"net/http"
)

const rulesURLFormat = "%s/v0/organizations/%s/rules"
const rulesURLFormatCreate = "%s/v0/organizations/%s/rules/create"
const rulesURLFormatUpdateDelete = "%s/v0/organizations/%s/rules/%d"

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

type CompareResults struct {
	RuleInsert []Rule
	RuleUpdate []Rule
	RuleDelete []Rule
}

// getRules queries Fairwinds Insights to retrieve all of the Rules for an organization
func getRules(org, token, hostName string) ([]Rule, error) {
	url := fmt.Sprintf(rulesURLFormat, hostName, org)
	logrus.Infof("Rules URL: %s", url)
	resp, err := req.Get(url, getHeaders(token))
	if err != nil {
		return nil, err
	}
	var rules []Rule
	if resp.Response().StatusCode != http.StatusOK {
		logrus.Errorf("Invalid response code: %s %v", string(resp.Bytes()), resp.Response().StatusCode)
		return nil, errors.New("invalid response code")
	}
	err = resp.ToJSON(&rules)
	if err != nil {
		return nil, err
	}
	return rules, nil
}

// insertRule adds a new rule
func insertRule(org, token, hostName string, rule Rule) error {
	url := fmt.Sprintf(rulesURLFormatCreate, hostName, org)
	resp, err := req.Post(url, getHeaders(token), req.BodyJSON(&rule))
	if err != nil {
		return err
	}
	if resp.Response().StatusCode != http.StatusOK {
		logrus.Errorf("Invalid response code: %s %v", string(resp.Bytes()), resp.Response().StatusCode)
		return errors.New("invalid response code")
	}
	return nil
}

// updateRule updates an existing rule
func updateRule(org, token, hostName string, rule Rule) error {
	url := fmt.Sprintf(rulesURLFormatUpdateDelete, hostName, org, rule.ID)
	resp, err := req.Post(url, getHeaders(token), req.BodyJSON(&rule))
	if err != nil {
		return err
	}
	if resp.Response().StatusCode != http.StatusOK {
		logrus.Errorf("Invalid response code: %s %v", string(resp.Bytes()), resp.Response().StatusCode)
		return errors.New("invalid response code")
	}
	return nil
}

// deleteRule deletes an existing rule
func deleteRule(org, token, hostName string, rule Rule) error {
	url := fmt.Sprintf(rulesURLFormatUpdateDelete, hostName, org, rule.ID)
	resp, err := req.Delete(url, getHeaders(token), nil)
	if err != nil {
		return err
	}
	if resp.Response().StatusCode != http.StatusOK {
		logrus.Errorf("Invalid response code: %s %v", string(resp.Bytes()), resp.Response().StatusCode)
		return errors.New("invalid response code")
	}
	return nil
}

func BuildRulesTree(org, token, hostName string, tree treeprint.Tree) error {
	rules, err := getRules(org, token, hostName)
	if err != nil {
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
			value := fmt.Sprintf("Context: %s", rule.Context)
			rulesNode.AddNode(value)
		}
		if rule.Repository != "" {
			value := fmt.Sprintf("Cluster: %s", rule.Repository)
			rulesNode.AddNode(value)
		}
		if rule.ReportType != "" {
			value := fmt.Sprintf("Cluster: %s", rule.ReportType)
			rulesNode.AddNode(value)
		}
	}

	return nil
}

func getRulesFromFiles(files map[string][]string) ([]Rule, error) {
	var rules []Rule
	for _, ruleFiles := range files {
		var rule Rule
		for _, filePath := range ruleFiles {
			fileContents, err := ioutil.ReadFile(filePath)
			if err != nil {
				logrus.Error(err, "Error reading file", filePath)
				return nil, err
			}
			err = yaml.Unmarshal(fileContents, &rule)
			if err != nil {
				logrus.Error(err, "Error Unmarshalling check YAML", filePath)
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

func mapRulesWithName(rules []Rule) map[string]Rule {
	return funk.Map(rules, func(value Rule) (string, Rule) {
		return value.Name, value
	}).(map[string]Rule)
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
	mappedFileRules := mapRulesWithName(fileRules)
	mappedExistingRules := mapRulesWithName(existingRules)
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
func compareRules(folder, org, token, hostName string) (CompareResults, error) {
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
	existingRules, err := getRules(org, token, hostName)
	if err != nil {
		logrus.Error("Error during API call")
		return results, err
	}

	results = getRuleDifferences(fileRules, existingRules)
	return results, nil
}

func SyncRules(syncDir, org, insightsToken, host string, dryRun bool) error {
	results, err := compareRules(syncDir, org, insightsToken, host)
	if err != nil {
		return err
	}

	for _, ruleForInsert := range results.RuleInsert {
		logrus.Infof("Adding rule: %s", ruleForInsert.Name)
		if !dryRun {
			err = insertRule(org, insightsToken, host, ruleForInsert)
			if err != nil {
				return err
			}
		}
	}

	for _, ruleForUpdate := range results.RuleUpdate {
		logrus.Infof("Updating rule: %s", ruleForUpdate.Name)
		if !dryRun {
			err = updateRule(org, insightsToken, host, ruleForUpdate)
			if err != nil {
				return err
			}
		}
	}

	for _, ruleForDelete := range results.RuleDelete {
		logrus.Infof("Deleting rule: %s", ruleForDelete.Name)
		if !dryRun {
			err = deleteRule(org, insightsToken, host, ruleForDelete)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func getHeaders(token string) req.Header {
	return req.Header{
		"Authorization": fmt.Sprintf("Bearer %s", token),
		"Accept":        "application/json",
	}
}
