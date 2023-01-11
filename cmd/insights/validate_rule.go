package main

import (
	"fmt"
	"io"
	"os"

	"github.com/fairwindsops/insights-cli/pkg/rules"
	"github.com/google/go-cmp/cmp"
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
	verifyRuleCmd.PersistentFlags().StringVarP(&expectedActionItem, "expected-action-item", "i", "", "Optional file containing the action item that the automation rule is expected to produce")
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
			exitWithError(fmt.Sprintf("Error when trying to open action item file %s", actionItemFile), err)
		}
		aiBytes, err := io.ReadAll(aiInput)
		if err != nil {
			exitWithError(fmt.Sprintf("Could not read action item file %s", actionItemFile), err)
		}
		err = yaml.Unmarshal(aiBytes, &actionItem)
		if err != nil {
			exitWithError(fmt.Sprintf("Could not parse action item file %s", actionItemFile), err)
		}

		ruleInput, err := os.Open(automationRuleFile)
		if err != nil {
			exitWithError(fmt.Sprintf("Error when trying to open file %s", automationRuleFile), err)
		}
		ruleBytes, err := io.ReadAll(ruleInput)
		if err != nil {
			exitWithError(fmt.Sprintf("Could not read rule file %s", automationRuleFile), err)
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
			diff := cmp.Diff(&expectedActionItem, response)
			if len(diff) == 0 {
				logrus.Infoln("Success - actual response matches expected response")
			} else {
				logrus.Errorln("Test failed:")
				fmt.Println(diff)
			}
		}
	},
}
