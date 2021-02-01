// Copyright 2020 FairwindsOps Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/fairwindsops/insights-cli/pkg/insights"
	"github.com/fairwindsops/insights-cli/pkg/opa"
)

var syncDir string
var gitOps bool
var dryrun bool

func init() {
	syncCmd.PersistentFlags().StringVarP(&syncDir, "directory", "d", ".", "Directory to sync.")
	syncCmd.PersistentFlags().BoolVarP(&gitOps, "fullsync", "", false, "Delete any checks not found in this repository.")
	syncCmd.PersistentFlags().BoolVarP(&dryrun, "dry-run", "", false, "Simulates a sync.")
	policyCmd.AddCommand(syncCmd)
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize policies",
	Long:  "Synchronize OPA policies to Fairwinds Insights",
	Run: func(cmd *cobra.Command, args []string) {
		org := configurationObject.Options.Organization
		host := configurationObject.Options.Hostname
		results, err := opa.CompareChecks(syncDir, org, insightsToken, host, gitOps)
		if err != nil {
			logrus.Fatalf("Unable to compare checks - %s", err)
		}
		for _, instance := range results.InstanceDelete {
			logrus.Infof("Deleting instance: %s:%s", instance.CheckName, instance.InstanceName)
			if !dryrun {
				err := insights.DeleteInstance(instance, org, insightsToken, host)
				if err != nil {
					logrus.Fatalf("Unable to delete instance %s:%s - %s", instance.CheckName, instance.InstanceName, err)
				}
			}
		}
		for _, check := range results.CheckDelete {
			logrus.Infof("Deleting check: %s", check.CheckName)
			if !dryrun {
				err := insights.DeleteCheck(check, org, insightsToken, host)
				if err != nil {
					logrus.Fatalf("Unable to delete check %s - %s", check.CheckName, err)
				}
			}
		}
		for _, check := range results.CheckInsert {
			logrus.Infof("Adding check: %s", check.CheckName)
			if !dryrun {
				err := insights.PutCheck(check, org, insightsToken, host)
				if err != nil {
					logrus.Fatalf("Unable to add check %s - %s", check.CheckName, err)
				}
			}
		}
		for _, check := range results.CheckUpdate {
			logrus.Infof("Updating check: %s", check.CheckName)
			if !dryrun {
				err := insights.PutCheck(check, org, insightsToken, host)
				if err != nil {
					logrus.Fatalf("Unable to update check %s - %s", check.CheckName, err)
				}
			}
		}
		for _, instance := range results.InstanceInsert {
			logrus.Infof("Adding instance: %s:%s", instance.CheckName, instance.InstanceName)
			if !dryrun {
				err := insights.PutInstance(instance, org, insightsToken, host)
				if err != nil {
					logrus.Fatalf("Unable to add instance %s:%s - %s", instance.CheckName, instance.InstanceName, err)
				}
			}
		}
		for _, instance := range results.InstanceUpdate {
			logrus.Infof("Updating instance: %s:%s", instance.CheckName, instance.InstanceName)
			if !dryrun {
				err := insights.PutInstance(instance, org, insightsToken, host)
				if err != nil {
					logrus.Fatalf("Unable to update instance %s:%s - %s", instance.CheckName, instance.InstanceName, err)
				}
			}
		}
	},
}
