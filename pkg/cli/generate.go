package cli

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// generateCmd represents the validate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate resources",
	Long:  `Generate resources using different strategies`,
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
	rootCmd.AddCommand(generateCmd)
}
