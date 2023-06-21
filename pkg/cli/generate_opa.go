package cli

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// generateOPACmd represents the validate command
var generateOPACmd = &cobra.Command{
	Use:   "opa",
	Short: "Generate Insights OPA policies",
	Long:  `Generate Insights OPA policies using different strategies`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Error("Please specify a sub-command.")
		err := cmd.Help()
		if err != nil {
			logrus.Error(err)
		}
		os.Exit(1)
	},
}

func init() {
	generateCmd.AddCommand(generateOPACmd)
}
