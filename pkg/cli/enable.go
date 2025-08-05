package cli

import (
	"os"

	"github.com/fairwindsops/insights-cli/pkg/organizations"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var enable = &cobra.Command{
	Use:   "enable",
	Short: "Enable Insights Features",
	Long:  "Enable Insights Features",
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Error("Please specify a sub-command.")
		err := cmd.Help()
		if err != nil {
			logrus.Error(err)
		}
		os.Exit(1)
	},
}

var enablePolicyMappings = &cobra.Command{
	Use:    "policy-mappings",
	Short:  "Enable policy mappings",
	Long:   "Enable policy mappings",
	PreRun: validateAndLoadInsightsAPIConfigWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		err := organizations.ManageOrganizationPolicyMappings(client, configurationObject.Options.Organization, true)
		if err != nil {
			logrus.Fatal(err)
		}
		logrus.Info("Successfully enabled policy mappings")
	},
}

func init() {
	rootCmd.AddCommand(enable)
	enable.AddCommand(enablePolicyMappings)
}
