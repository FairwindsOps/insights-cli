package cli

import (
	"github.com/fairwindsops/insights-cli/pkg/rules"
	"github.com/spf13/cobra"
)

var automationRuleFilePath, actionItemFilePath, expectedActionItemFilePath, insightsContext string
var dryRun bool

func init() {
	verifyRuleCmd.PersistentFlags().StringVarP(&automationRuleFilePath, "automation-rule-file", "r", "./rule.js", "Automation rule JS file path")
	verifyRuleCmd.PersistentFlags().StringVarP(&actionItemFilePath, "action-item-file", "a", "./action-item.yaml", "Action Item file path")
	verifyRuleCmd.PersistentFlags().StringVarP(&insightsContext, "insights-context", "t", "", "Insights context: [AdmissionController, Agent or CI/CD]")
	verifyRuleCmd.PersistentFlags().StringVarP(&expectedActionItemFilePath, "expected-action-item", "i", "", "Optional file containing the action item that the automation rule is expected to produce")
	verifyRuleCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "", false, "Optional flag to run the rule without executing any external integration (Slack, Jira, PagerDuty, Azure, http requests).")
	err := verifyRuleCmd.MarkPersistentFlagRequired("insights-context")
	if err != nil {
		exitWithError("", err)
	}
	validateCmd.AddCommand(verifyRuleCmd)
}

var verifyRuleCmd = &cobra.Command{
	Use:    "rule -t <insights context> -r <rule file> -a <action item file> [-i <expected output file> --dry-run]",
	Short:  "Validates an automation rule",
	Long:   "Validates an automation rule by applying it against the specified action item",
	PreRun: validateAndLoadInsightsAPIConfigWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		org := configurationObject.Options.Organization
		err := rules.ValidateRule(client, org, automationRuleFilePath, actionItemFilePath, expectedActionItemFilePath, insightsContext, dryRun)
		if err != nil {
			exitWithError("", err)
		}
	},
}
