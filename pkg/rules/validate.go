package rules

import (
	"fmt"
	"io"
	"os"

	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func ValidateRule(org, host, insightsToken, automationRuleFilePath, actionItemFilePath, expectedActionItemFilePath, reportType, insightsContext string, dryRun bool) error {
	aiInput, err := os.Open(actionItemFilePath)
	if err != nil {
		return fmt.Errorf("error when trying to open action item file %s: %v", actionItemFilePath, err)
	}

	aiBytes, err := io.ReadAll(aiInput)
	if err != nil {
		return fmt.Errorf("could not read action item file %s: %v", actionItemFilePath, err)
	}

	var actionItem ActionItem
	err = yaml.Unmarshal(aiBytes, &actionItem)
	if err != nil {
		return fmt.Errorf("could not parse action item file %s: %v", actionItemFilePath, err)
	}

	ruleInput, err := os.Open(automationRuleFilePath)
	if err != nil {
		return fmt.Errorf("error when trying to open file %s: %v", automationRuleFilePath, err)
	}

	ruleBytes, err := io.ReadAll(ruleInput)
	if err != nil {
		return fmt.Errorf("could not read rule file %s: %v", automationRuleFilePath, err)
	}

	verifyRule := VerifyRule{
		ActionItem: actionItem,
		Context:    RuleExecutionContext(insightsContext),
		ReportType: reportType,
		Script:     string(ruleBytes),
	}
	r, err := RunVerifyRule(org, insightsToken, host, verifyRule, dryRun)
	if err != nil {
		return fmt.Errorf("unable to verify rule: %v", err)
	}
	responseActionItem := r.ActionItem

	if len(r.Events) > 0 {
		fmt.Printf("\n-- Logs --\n\n")
		for _, e := range r.Events {
			fmt.Println(e)
		}
	}

	if expectedActionItemFilePath == "" {
		fmt.Printf("\n-- Returned Action Item --\n\n")
		b, err := yaml.Marshal(responseActionItem)
		if err != nil {
			return fmt.Errorf("could not marshal verify result: %v", err)
		}
		fmt.Println(string(b))
		return nil
	}

	expectedActionItemFile, err := os.Open(expectedActionItemFilePath)
	if err != nil {
		return fmt.Errorf("Error when trying to open expected file %s: %v", expectedActionItemFilePath, err)
	}

	expectedActionItemBytes, err := io.ReadAll(expectedActionItemFile)
	if err != nil {
		return fmt.Errorf("Failed to read output file: %v", err)
	}

	var expectedActionItem ActionItem
	err = yaml.Unmarshal(expectedActionItemBytes, &expectedActionItem)
	if err != nil {
		return fmt.Errorf("could not marshal expected response: %v", err)
	}
	diff := cmp.Diff(&expectedActionItem, responseActionItem)
	if len(diff) == 0 {
		logrus.Infoln("Success - actual response matches expected response")
	} else {
		logrus.Errorln("Test failed:")
		fmt.Println(diff)
	}
	return nil
}
