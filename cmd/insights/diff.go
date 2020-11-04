package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/fairwindsops/insights-cli/pkg/opa"
)

func init() {
	diffCmd.PersistentFlags().StringVarP(&syncDir, "directory", "d", ".", "Directory to sync.")
	policyCmd.AddCommand(diffCmd)
}

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Should changes in policies",
	Long:  "Shows diffs with OPA policies to Fairwinds Insights",
	Run: func(cmd *cobra.Command, args []string) {

		// TODO remove hardcoded Org and Host
		// TODO remove hardcoded GitOPs param
		// TODO Finish
		results, err := opa.CompareChecks(syncDir, "centaurus", insightsToken, "https://staging.insights.fairwinds.com", true)
		if err != nil {
			panic(err)
		}
		logrus.Infof("%+v", results)
	},
}
