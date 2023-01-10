package main

import (
	"fmt"
	"io"
	"os"

	"github.com/fairwindsops/insights-cli/pkg/rules"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var automationRuleFile, actionItemFile, expectedFileName, reportType, context string

func init() {
	verifyRuleCmd.PersistentFlags().StringVarP(&automationRuleFile, "automation-rule-file", "r", "./rule.js", "Automation rule JS file")
	verifyRuleCmd.PersistentFlags().StringVarP(&actionItemFile, "action-item-file", "a", "./action-item.yaml", "Action Item file")
	verifyRuleCmd.PersistentFlags().StringVarP(&context, "insights-context", "t", "", "Insights context")
	verifyRuleCmd.PersistentFlags().StringVarP(&reportType, "report-type", "R", "", "Report type")
	verifyRuleCmd.PersistentFlags().StringVarP(&expectedFileName, "expected-action-item", "i", "", "Expected file")
	validateCmd.AddCommand(verifyRuleCmd)
}

var verifyRuleCmd = &cobra.Command{
	Use:   "rule -t  <insights context> {-r <rule file> -a <action item file>} [-i <expected output file>]",
	Short: "Verifies a rule against Insights",
	Long:  "Verifies a rule to provided the action item and makes sure it will work fine testing it against Insights",
	Run: func(cmd *cobra.Command, args []string) {
		var verifyActionItem rules.VerifyActionItem
		org := configurationObject.Options.Organization
		host := configurationObject.Options.Hostname
		aiInput, err := os.Open(actionItemFile)
		if err != nil {
			exitWithError(fmt.Sprintf("Error when trying to open file %s", actionItemFile), err)
		}
		aiBytes, err := io.ReadAll(aiInput)
		if err != nil {
			exitWithError(fmt.Sprintf("Could not read file %s", actionItemFile), err)
		}
		err = yaml.Unmarshal(aiBytes, &verifyActionItem)
		if err != nil {
			exitWithError("Could not parse input file", err)
		}

		ruleInput, err := os.Open(automationRuleFile)
		if err != nil {
			exitWithError(fmt.Sprintf("Error when trying to open file %s", actionItemFile), err)
		}
		ruleBytes, err := io.ReadAll(ruleInput)
		if err != nil {
			exitWithError(fmt.Sprintf("Could not read file %s", actionItemFile), err)
		}
		verifyRule := rules.VerifyRule{
			ActionItem: verifyActionItem,
			Context:    rules.RuleExecutionContext(context),
			ReportType: reportType,
			Script:     string(ruleBytes),
		}
		response, err := rules.RunVerifyRule(org, insightsToken, host, verifyRule)
		if err != nil {
			exitWithError("Unable to verify rule:", err)
		}
		b, err := yaml.Marshal(response)
		if err != nil {
			exitWithError("could not marshal verify result", err)
		}
		if expectedFileName == "" {
			fmt.Println(string(b))
		} else {
			expectedFile, err := os.Open(expectedFileName)
			if err != nil {
				exitWithError(fmt.Sprintf("Error when trying to open expected file %s", expectedFileName), err)
			}
			expectedContent, err := io.ReadAll(expectedFile)
			if err != nil {
				exitWithError("Failed to read output file", err)
			}
			var expectedActionItem rules.VerifyActionItem
			err = yaml.Unmarshal(expectedContent, &expectedActionItem)
			if err != nil {
				exitWithError("could not marshal expected response", err)
			}
			testErrors := compareVerifyActionItem(*response, expectedActionItem)
			if len(testErrors) == 0 {
				logrus.Infoln("Success - actual response matches expected response")
			} else {
				logrus.Errorln("Test failed:")
				for _, msg := range testErrors {
					fmt.Println(msg)
				}
			}

		}

	},
}

func compareVerifyActionItem(response, expected rules.VerifyActionItem) []string {
	msgs := []string{}
	if *response.Description != *expected.Description {
		msgs = append(msgs, fmt.Sprintf("Expected description:%s\nActual description:%s", *response.Description, *expected.Description))
	}
	if response.Title != expected.Title {
		msgs = append(msgs, fmt.Sprintf("Expected title:%s\nActual title:%s", response.Title, expected.Title))
	}
	return msgs
}
