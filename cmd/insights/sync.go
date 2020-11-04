package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/fairwindsops/insights-cli/pkg/insights"
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

		// TODO remove hardcoded org, host, and GitOps Param
		org := "centaurus"
		host := "https://staging.insights.fairwinds.com"
		results, err := opa.CompareChecks(syncDir, org, insightsToken, host, true)
		if err != nil {
			panic(err)
		}
		for _, instance := range results.InstanceDelete {
			logrus.Infof("Deleting instance: %s:%s", instance.CheckName, instance.InstanceName)
			err := insights.DeleteInstance(instance, org, insightsToken, host)
			if err != nil {
				panic(err)
			}
		}
		for _, check := range results.CheckDelete {
			logrus.Infof("Deleting check: %s", check.CheckName)
			err := insights.DeleteCheck(check, org, insightsToken, host)
			if err != nil {
				panic(err)
			}
		}
		for _, check := range results.CheckInsert {
			logrus.Infof("Adding check: %s", check.CheckName)
			err := insights.PutCheck(check, org, insightsToken, host)
			if err != nil {
				panic(err)
			}
		}
		for _, check := range results.CheckUpdate {
			logrus.Infof("Updating check: %s", check.CheckName)
			err := insights.PutCheck(check, org, insightsToken, host)
			if err != nil {
				panic(err)
			}
		}
		for _, instance := range results.InstanceInsert {
			logrus.Infof("Adding instance: %s:%s", instance.CheckName, instance.InstanceName)
			err := insights.PutInstance(instance, org, insightsToken, host)
			if err != nil {
				panic(err)
			}
		}
		for _, instance := range results.InstanceUpdate {
			logrus.Infof("Updating instance: %s:%s", instance.CheckName, instance.InstanceName)
			err := insights.PutInstance(instance, org, insightsToken, host)
			if err != nil {
				panic(err)
			}
		}
	},
}
