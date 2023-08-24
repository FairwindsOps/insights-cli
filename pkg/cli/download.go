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

var downloadDir string

func init() {
	downloadCmd.PersistentFlags().StringVarP(&downloadDir, "download-directory", "d", ".", "target directory to download content pulled from Insights.")
	rootCmd.AddCommand(downloadCmd)
}

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download custom configuration resources from Insights",
	Long:  "Download custom configuration such as App Groups from Insights.",
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Error("Please specify a sub-command.")
		err := cmd.Help()
		if err != nil {
			logrus.Error(err)
		}
		os.Exit(1)
	},
}
