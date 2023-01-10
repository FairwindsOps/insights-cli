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
	"net/http"
	"os"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/imroc/req"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"github.com/xlab/treeprint"

	"github.com/fairwindsops/insights-cli/pkg/directory"
)

const rulesURLFormat = "%s/v0/organizations/%s/rules"
const rulesURLFormatCreate = "%s/v0/organizations/%s/rules/create"
const rulesURLFormatUpdateDelete = "%s/v0/organizations/%s/rules/%d"
const rulesURLVerify = "%s/v0/organizations/%s/rules/verify"

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

type VerifyActionItemTicketProvider string

// Defines values for VerifyActionItemTicketProvider.
const (
	VerifyActionItemTicketProviderAzure  VerifyActionItemTicketProvider = "Azure"
	VerifyActionItemTicketProviderGitHub VerifyActionItemTicketProvider = "GitHub"
	VerifyActionItemTicketProviderJira   VerifyActionItemTicketProvider = "Jira"
)

type ActionItem struct {
	TicketCreatedAt   *time.Time                      `yaml:"TicketCreatedAt,omitempty"`
	TicketLink        *string                         `yaml:"TicketLink,omitempty"`
	TicketProvider    *VerifyActionItemTicketProvider `yaml:"TicketProvider,omitempty"`
	AssigneeEmail     *string                         `yaml:"assigneeEmail,omitempty"`
	Category          *string                         `yaml:"category,omitempty"`
	Cluster           *string                         `yaml:"cluster,omitempty"`
	DeletedAt         *time.Time                      `yaml:"deletedAt,omitempty"`
	Description       *string                         `yaml:"description,omitempty"`
	EventType         *string                         `yaml:"eventType,omitempty"`
	FirstSeen         *time.Time                      `yaml:"firstSeen,omitempty"`
	Fixed             *bool                           `yaml:"fixed,omitempty"`
	IsCustom          *bool                           `yaml:"isCustom,omitempty"`
	LastReportedAt    *time.Time                      `yaml:"lastReportedAt,omitempty"`
	Notes             *string                         `yaml:"notes,omitempty"`
	Organization      *string                         `yaml:"organization,omitempty"`
	Remediation       *string                         `yaml:"remediation,omitempty"`
	ReportType        *string                         `yaml:"reportType,omitempty"`
	Resolution        *string                         `yaml:"resolution,omitempty"`
	ResourceContainer *string                         `yaml:"resourceContainer,omitempty"`
	ResourceKind      *string                         `yaml:"resourceKind,omitempty"`
	ResourceLabels    map[string]string               `yaml:"resourceLabels,omitempty"`
	ResourceName      *string                         `yaml:"resourceName,omitempty"`
	ResourceNamespace *string                         `yaml:"resourceNamespace,omitempty"`
	Severity          *float32                        `yaml:"severity,omitempty"`
	Tags              []string                        `yaml:"tags"`
	Title             string                          `yaml:"title"`
}

// RuleExecutionContext defines model for RuleExecutionContext.
type RuleExecutionContext string

// Defines values for RuleExecutionContext.
const (
	RuleExecutionContextAdmissionController RuleExecutionContext = "AdmissionController"
	RuleExecutionContextAgent               RuleExecutionContext = "Agent"
	RuleExecutionContextCICD                RuleExecutionContext = "CI/CD"
)

type VerifyRule struct {
	ActionItem ActionItem           `yaml:"actionItem"`
	Context    RuleExecutionContext `yaml:"context"`
	ReportType string               `yaml:"reportType"`
	Script     string               `yaml:"script"`
}

// getRules queries Fairwinds Insights to retrieve all of the Rules for an organization
func getRules(org, token, hostName string) ([]Rule, error) {
	url := fmt.Sprintf(rulesURLFormat, hostName, org)
	logrus.Debugf("Rules URL: %s", url)
	resp, err := req.Get(url, getHeaders(token))
	if err != nil {
		logrus.Errorf("Unable to get rules from insights: %v", err)
		return nil, err
	}
	var rules []Rule
	if resp.Response().StatusCode != http.StatusOK {
		logrus.Errorf("Invalid response code: %s %v", string(resp.Bytes()), resp.Response().StatusCode)
		return nil, errors.New("invalid response code")
	}
	err = resp.ToJSON(&rules)
	if err != nil {
		logrus.Errorf("Unable to convert response to json for rules: %v", err)
		return nil, err
	}
	return rules, nil
}

// insertRule adds a new rule
func insertRule(org, token, hostName string, rule Rule) error {
	url := fmt.Sprintf(rulesURLFormatCreate, hostName, org)
	resp, err := req.Post(url, getHeaders(token), req.BodyJSON(&rule))
	if err != nil {
		logrus.Errorf("Unable to add rule %s to insights: %v", rule.Name, err)
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
		logrus.Errorf("Unable to update rule %s to insights: %v", rule.Name, err)
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
		logrus.Errorf("Unable to delete rule %s from insights: %v", rule.Name, err)
		return err
	}
	if resp.Response().StatusCode != http.StatusOK {
		logrus.Errorf("Invalid response code: %s %v", string(resp.Bytes()), resp.Response().StatusCode)
		return errors.New("invalid response code")
	}
	return nil
}

// RunVerifyRule verifies rule against one action item
func RunVerifyRule(org, token, hostName string, rule VerifyRule) (*ActionItem, error) {
	url := fmt.Sprintf(rulesURLVerify, hostName, org)
	resp, err := req.Post(url, getRuleVerifyHeaders(token), req.BodyJSON(&rule))
	if err != nil {
		logrus.Errorf("error verifying rule %v in insights: %v", rule, err)
		return nil, err
	}
	if resp.Response().StatusCode != http.StatusOK {
		logrus.Errorf("invalid response code: %s %v", string(resp.Bytes()), resp.Response().StatusCode)
		return nil, errors.New("invalid response code")
	}
	var verify *ActionItem
	err = resp.ToJSON(&verify)
	if err != nil {
		logrus.Errorf("unable to convert response to json to VerifyActionItem: %v", err)
		return nil, err
	}
	return verify, nil
}

// BuildRulesTree builds a tree for rules
func BuildRulesTree(org, token, hostName string, tree treeprint.Tree) error {
	rules, err := getRules(org, token, hostName)
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
	files, err := directory.ScanRulesFolder(folder)
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

// PushRules pushes automation rules to insights
func PushRules(pushDir, org, insightsToken, host string, deleteMissing, dryrun bool) error {
	logrus.Debugln("Pushing automation rules")
	_, err := os.Stat(pushDir)
	if err != nil {
		return err
	}
	results, err := compareRules(pushDir, org, insightsToken, host)
	if err != nil {
		logrus.Errorf("unable to compare and push rules to Insights: %v", err)
		return err
	}

	for _, ruleForInsert := range results.RuleInsert {
		logrus.Infof("Adding automation rule: %s", ruleForInsert.Name)
		if !dryrun {
			err = insertRule(org, insightsToken, host, ruleForInsert)
			if err != nil {
				logrus.Errorf("Error while adding rule %s to insights: %v", ruleForInsert.Name, err)
				return err
			}
		}
	}

	for _, ruleForUpdate := range results.RuleUpdate {
		logrus.Infof("Updating automation rule: %s", ruleForUpdate.Name)
		if !dryrun {
			err = updateRule(org, insightsToken, host, ruleForUpdate)
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
				err = deleteRule(org, insightsToken, host, ruleForDelete)
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

func getHeaders(token string) req.Header {
	return req.Header{
		"Authorization": fmt.Sprintf("Bearer %s", token),
		"Accept":        "application/json",
	}
}

func getRuleVerifyHeaders(token string) req.Header {
	return req.Header{
		"Authorization": fmt.Sprintf("Bearer %s", token),
		"Accept":        "application/json",
		"Content-Type":  "application/yaml",
	}
}
