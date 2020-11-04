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
	"flag"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var logLevel string
var insightsToken string

func init() {
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "", logrus.InfoLevel.String(), "Logrus log level.")
	flag.Parse()
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
}

var rootCmd = &cobra.Command{
	Use:   "insights",
	Short: "insights",
	Long:  `Interact with Fairwinds Insights.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		parsedLevel, err := logrus.ParseLevel(logLevel)
		if err != nil {
			logrus.Errorf("log-level flag has invalid value %s", logLevel)
		} else {
			logrus.SetLevel(parsedLevel)
		}
		insightsToken = os.Getenv("FAIRWINDS_TOKEN")
		if insightsToken == "" {
			logrus.Error("FAIRWINDS_TOKEN must be set.")
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Error("You must specify a sub-command.")
		err := cmd.Help()
		if err != nil {
			logrus.Error(err)
		}
		os.Exit(1)
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
}
