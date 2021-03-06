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
	"os"

	"github.com/fairwindsops/insights-cli/pkg/opa"
	"github.com/fairwindsops/insights-cli/pkg/rules"
)

var syncDir string
var fullsync bool
var dryrun bool
var forRules bool
var forChecks bool

func init() {
	syncCmd.PersistentFlags().StringVarP(&syncDir, "directory", "d", ".", "Directory to sync.")
	syncCmd.PersistentFlags().BoolVarP(&fullsync, "fullsync", "", false, "Delete any checks not found in this repository.")
	syncCmd.PersistentFlags().BoolVarP(&dryrun, "dry-run", "", false, "Simulates a sync.")
	syncCmd.PersistentFlags().BoolVarP(&forRules, "rules", "", false, "Sync only rules.")
	syncCmd.PersistentFlags().BoolVarP(&forChecks, "checks", "", false, "Sync only OPA checks.")
	policyCmd.AddCommand(syncCmd)
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize policies",
	Long:  "Synchronize OPA policies to Fairwinds Insights",
	Run: func(cmd *cobra.Command, args []string) {
		org := configurationObject.Options.Organization
		host := configurationObject.Options.Hostname
		if !forRules || forChecks {
			opaDir := syncDir + "/checks"
			_, err := os.Stat(opaDir)
			if os.IsNotExist(err) {
				logrus.Fatal("Folder for OPA checks doesn't exist. All OPA checks must be in a folder wih the name checks.")
			}
			err = opa.SyncOPAChecks(opaDir, org, insightsToken, host, fullsync, dryrun)
			if err != nil {
				logrus.Fatalf("Unable to sync OPA Checks: %v", err)
			}

		}
		if !forChecks || forRules {
			rulesDir := syncDir + "/rules"
			_, err := os.Stat(rulesDir)
			if os.IsNotExist(err) {
				logrus.Fatal("Folder for rules doesn't exist. All rules must be in a folder wih the name rules.")
			}
			err = rules.SyncRules(rulesDir, org, insightsToken, host, fullsync, dryrun)
			if err != nil {
				logrus.Fatalf("Unable to sync rules: %v", err)
			}
		}

	},
}
