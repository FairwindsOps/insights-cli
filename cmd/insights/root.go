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
	"errors"
	"flag"
	"io/ioutil"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

var logLevel string
var insightsToken string
var configFile string
var organization string

var configurationObject configuration

type configuration struct {
	Options optionConfig `yaml:"options"`
}

type optionConfig struct {
	BaseBranch     string `yaml:"baseBranch"`
	Hostname       string `yaml:"hostname"`
	Organization   string `yaml:"organization"`
	RepositoryName string `yaml:"repositoryName"`
}

// SetDefaults sets configurationd defaults
func (c *configuration) SetDefaults() {
	if c.Options.BaseBranch == "" {
		c.Options.BaseBranch = "main"
	}
	if c.Options.Hostname == "" {
		c.Options.Hostname = "https://insights.fairwinds.com"
	}
}

// CheckForErrors checks to make sure the configuration is valid
func (c configuration) CheckForErrors() error {
	if c.Options.Organization == "" {
		return errors.New("options.organization not set")
	}
	return nil
}

func exitWithError(message string, err error) {
	if err != nil {
		logrus.Fatalf("%s: %s", message, err.Error())
	} else {
		logrus.Fatal(message)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "", logrus.InfoLevel.String(), "Logrus log level.")
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "./fairwinds-insights.yaml", "Configuration file")
	rootCmd.PersistentFlags().StringVarP(&organization, "organization", "", "", "Fairwinds Insights Organization name")
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flag.Parse()
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
}

func preRun(cmd *cobra.Command, args []string) {
	parsedLevel, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logrus.Errorf("log-level flag has invalid value %s", logLevel)
	} else {
		logrus.SetLevel(parsedLevel)
	}
	if cmd.Use == "opa" {
		return
	}
	insightsToken = os.Getenv("FAIRWINDS_TOKEN")
	if insightsToken == "" {
		exitWithError("FAIRWINDS_TOKEN must be set.", nil)
	}

	configHandler, err := os.Open(configFile)
	if err == nil {
		configContents, err := ioutil.ReadAll(configHandler)
		if err != nil {
			exitWithError("Could not read fairwinds-insights.yaml", err)
		}
		err = yaml.Unmarshal(configContents, &configurationObject)
		if err != nil {
			exitWithError("Could not parse fairwinds-insights.yaml", err)
		}
	} else if !os.IsNotExist(err) {
		exitWithError("Could not open fairwinds-insights.yaml", err)
	} else if organization == "" {
		exitWithError("Please add fairwinds-insights.yaml to the base of your repository.", nil)
	}
	configurationObject.SetDefaults()
	if organization != "" {
		configurationObject.Options.Organization = organization
	}
	err = configurationObject.CheckForErrors()
	if err != nil {
		exitWithError("Error parsing fairwinds-insights.yaml", err)
	}
}

var rootCmd = &cobra.Command{
	Use:              "insights",
	Short:            "insights",
	Long:             `Interact with Fairwinds Insights.`,
	PersistentPreRun: preRun,
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
