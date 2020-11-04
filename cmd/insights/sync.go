package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/fairwindsops/insights-cli/pkg/opa"
)

var syncDir string

func init() {
	syncCmd.PersistentFlags().StringVarP(&syncDir, "directory", "d", ".", "Directory to sync.")
	policyCmd.AddCommand(syncCmd)
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize policies",
	Long:  "Synchronize OPA policies to Fairwinds Insights",
	Run: func(cmd *cobra.Command, args []string) {

		// TODO Finish
		results, err := opa.CompareChecks(syncDir, "centaurus", insightsToken, "https://staging.insights.fairwinds.com", true)
		if err != nil {
			panic(err)
		}
		logrus.Infof("%+v", results)
	},
}
