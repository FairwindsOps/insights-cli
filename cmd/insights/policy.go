package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(policyCmd)
}

var policyCmd = &cobra.Command{
	Use:   "policy",
	Short: "Commands related to Policies",
	Long:  "Commands for managing OPA policies or Rules in Fairwinds Insights",
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Error("You must specify a sub-command.")
		err := cmd.Help()
		if err != nil {
			logrus.Error(err)
		}
		os.Exit(1)

	},
}
