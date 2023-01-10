package main

import (
	"fmt"
	"io"
	"os"

	"github.com/fairwindsops/insights-cli/pkg/rules"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var automationRuleFile, actionItemFile, expectedActionItem, reportType, insightsContext string

func init() {
	verifyRuleCmd.PersistentFlags().StringVarP(&automationRuleFile, "automation-rule-file", "r", "./rule.js", "Automation rule JS file path")
	verifyRuleCmd.PersistentFlags().StringVarP(&actionItemFile, "action-item-file", "a", "./action-item.yaml", "Action Item file path")
	verifyRuleCmd.PersistentFlags().StringVarP(&insightsContext, "insights-context", "t", "", "Insights context: [AdmissionController/Agent/CI/CD]")
	verifyRuleCmd.PersistentFlags().StringVarP(&reportType, "report-type", "R", "", "Report type")
	verifyRuleCmd.PersistentFlags().StringVarP(&expectedActionItem, "expected-action-item", "i", "", "Expected file path")
	validateCmd.AddCommand(verifyRuleCmd)
}

var verifyRuleCmd = &cobra.Command{
	Use:   "rule -t  <insights context> {-r <rule file> -a <action item file>} [-i <expected output file>]",
	Short: "Validates an automation rule",
	Long:  "Validates an automation rule by applying it against the specified action item",
	Run: func(cmd *cobra.Command, args []string) {
		var actionItem rules.ActionItem
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
		err = yaml.Unmarshal(aiBytes, &actionItem)
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
			ActionItem: actionItem,
			Context:    rules.RuleExecutionContext(insightsContext),
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
		if expectedActionItem == "" {
			fmt.Println(string(b))
		} else {
			expectedFile, err := os.Open(expectedActionItem)
			if err != nil {
				exitWithError(fmt.Sprintf("Error when trying to open expected file %s", expectedActionItem), err)
			}
			expectedContent, err := io.ReadAll(expectedFile)
			if err != nil {
				exitWithError("Failed to read output file", err)
			}
			var expectedActionItem rules.ActionItem
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

func compareVerifyActionItem(response, expected rules.ActionItem) []string {
	msgs := []string{}
	msg := maybeReturnErrorMsgForStringPtr(expected.Description, response.Description, "description")
	if msg != nil {
		msgs = append(msgs, *msg)
	}
	msg = maybeReturnErrorMsgForString(expected.Title, response.Title, "title")
	if msg != nil {
		msgs = append(msgs, *msg)
	}
	return msgs
}

func maybeReturnErrorMsgForStringPtr(expected, actual *string, fieldName string) *string {
	if expected == nil && actual == nil {
		return nil
	}
	if expected == nil && actual != nil {
		return lo.ToPtr(fmt.Sprintf("Expected %s is NULL\nActual %s: %s", fieldName, fieldName, *actual))
	}
	if expected != nil && actual == nil {
		return lo.ToPtr(fmt.Sprintf("Expected %s: %s\nActual %s is NULL", fieldName, *expected, fieldName))
	}
	return maybeReturnErrorMsgForString(*expected, *actual, fieldName)
}

func maybeReturnErrorMsgForString(expected, actual, fieldName string) *string {
	if expected != actual {
		return lo.ToPtr(fmt.Sprintf("Expected %s: %s\nActual %s: %s", fieldName, expected, fieldName, actual))
	}
	return nil
}
