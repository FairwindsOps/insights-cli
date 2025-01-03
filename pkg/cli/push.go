// Copyright 2023 FairwindsOps Inc
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

package cli

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var pushDir string
var pushDryRun bool
var pushDeleteNonProvidedTeams bool

func init() {
	pushCmd.PersistentFlags().StringVarP(&pushDir, "push-directory", "d", ".", "Directory of content to push to Insights.")
	pushCmd.PersistentFlags().BoolVarP(&pushDryRun, "dry-run", "z", false, "Explains what would be pushed to Insights, without making changes.")
	pushCmd.PersistentFlags().BoolVarP(&pushDeleteNonProvidedTeams, "delete", "x", false, "Deletes teams that are not provided in the push directory.")
	rootCmd.AddCommand(pushCmd)
}

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push local file-based custom configuration resources into Insights.",
	Long:  "Commands to push custom configuration resources, such as OPA policies and automation rules, into Insights.",
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Error("Please specify a sub-command.")
		err := cmd.Help()
		if err != nil {
			logrus.Error(err)
		}
		os.Exit(1)
	},
}
