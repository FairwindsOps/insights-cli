package cli

import (
	"os"

	"github.com/fairwindsops/insights-cli/pkg/organizations"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var disable = &cobra.Command{
	Use:   "disable",
	Short: "Disable Insights Features",
	Long:  "Disable Insights Features",
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Error("Please specify a sub-command.")
		err := cmd.Help()
		if err != nil {
			logrus.Error(err)
		}
		os.Exit(1)
	},
}

var disablePolicyMappings = &cobra.Command{
	Use:    "policy-mappings",
	Short:  "Disable policy mappings",
	Long:   "Disable policy mappings",
	PreRun: validateAndLoadInsightsAPIConfigWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		err := organizations.ManageOrganizationPolicyMappings(client, configurationObject.Options.Organization, insightsToken, configurationObject.Options.Hostname, false)
		if err != nil {
			logrus.Fatal(err)
		}
		logrus.Info("Successfully disabled policy mappings")
	},
}

func init() {
	rootCmd.AddCommand(disable)
	disable.AddCommand(disablePolicyMappings)
}
