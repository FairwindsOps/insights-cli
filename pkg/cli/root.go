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
	"errors"
	"fmt"
	"io"
	"os"

	cliversion "github.com/fairwindsops/insights-cli/pkg/version"
	"github.com/imroc/req/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/xlab/treeprint"
	"gopkg.in/yaml.v3"
)

var client = req.C()

var logLevel string
var configFile string
var organization string
var noDecoration bool

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

// SetDefaults sets configuration defaults
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

// RUn executes the cobra root command, and returns an exit value depending on
// whether an error occurred.
func Run() (exitValue int) {
	if err := rootCmd.Execute(); err != nil {
		logrus.Error(err)
		return 1
	}
	return 0
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
	rootCmd.PersistentFlags().BoolVarP(&noDecoration, "no-decoration", "", false, "Do not include decorative characters in output, such as tree visualization.")

	logrus.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true, DisableLevelTruncation: true})
}

func validateAndLoadInsightsAPIConfigWrapper(cmd *cobra.Command, args []string) {
	err := validateAndLoadInsightsAPIConfig(client)
	if err != nil {
		logrus.Fatal(err)
	}
}

// validateAndLoadInsightsAPIConfig checks to make sure the user has set the FAIRWINDS_TOKEN environment variable and has a valid config file.
func validateAndLoadInsightsAPIConfig(client *req.Client) error {
	insightsToken := os.Getenv("FAIRWINDS_TOKEN")
	if insightsToken == "" {
		return errors.New("FAIRWINDS_TOKEN must be set")
	}
	configHandler, err := os.Open(configFile)
	if err == nil {
		configContents, err := io.ReadAll(configHandler)
		if err != nil {
			return fmt.Errorf("could not read fairwinds-insights.yaml: %v", err)
		}
		err = yaml.Unmarshal(configContents, &configurationObject)
		if err != nil {
			return fmt.Errorf("could not parse fairwinds-insights.yaml: %v", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("could not open fairwinds-insights.yaml: %v", err)
	} else if organization == "" {
		return fmt.Errorf("please add fairwinds-insights.yaml to the base of your repository: %v", err)
	}
	configurationObject.SetDefaults()
	if organization != "" {
		configurationObject.Options.Organization = organization
	}
	err = configurationObject.CheckForErrors()
	if err != nil {
		return fmt.Errorf("error parsing fairwinds-insights.yaml: %v", err)
	}

	// common client configuration
	client.SetCommonHeaders(map[string]string{
		"Authorization":           fmt.Sprintf("Bearer %s", insightsToken),
		"X-Fairwinds-CLI-Version": cliversion.GetVersion(),
	})

	client.SetBaseURL(configurationObject.Options.Hostname)

	return nil
}

func preRun(cmd *cobra.Command, args []string) {
	if noDecoration {
		treeprint.EdgeTypeLink = " "
		treeprint.EdgeTypeMid = "  "
		treeprint.EdgeTypeEnd = "  "
	}
	parsedLevel, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logrus.Errorf("log-level flag has invalid value %s", logLevel)
	} else {
		logrus.SetLevel(parsedLevel)
	}
}

var rootCmd = &cobra.Command{
	Use:              "insights-cli",
	Short:            "insights",
	Long:             `Interact with Fairwinds Insights.`,
	PersistentPreRun: preRun,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Error("Please specify a command.")
		err := cmd.Help()
		if err != nil {
			logrus.Error(err)
		}
		os.Exit(1)
	},
}
