package main

import (
	"fmt"
	"io"
	"os"

	"github.com/fairwindsops/insights-cli/pkg/rules"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var inputFile, outputFile string

func init() {
	verifyRuleCmd.PersistentFlags().StringVarP(&inputFile, "input", "i", "./input.yaml", "Input file")
	verifyRuleCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "./output.yaml", "Output file")
	rootCmd.AddCommand(verifyRuleCmd)
}

var verifyRuleCmd = &cobra.Command{
	Use:   "verify-rule",
	Short: "Verifies a rule against Insights",
	Long:  "Verifies a rule to provided action item and make sure it will work fine testing it against Insights",
	Run: func(cmd *cobra.Command, args []string) {
		var verifyRule rules.VerifyRule
		org := configurationObject.Options.Organization
		host := configurationObject.Options.Hostname
		in, err := os.Open(inputFile)
		if err != nil {
			exitWithError(fmt.Sprintf("Error when trying to read file %s", inputFile), err)
		}
		inputContents, err := io.ReadAll(in)
		if err != nil {
			exitWithError(fmt.Sprintf("Could not read file %s", inputFile), err)
		}
		err = yaml.Unmarshal(inputContents, &verifyRule)
		if err != nil {
			exitWithError("Could not parse input file", err)
		}
		result, err := rules.RunVerifyRule(org, insightsToken, host, verifyRule)
		if err != nil {
			exitWithError("Unable to verify rule:", err)
		}
		b, err := yaml.Marshal(result)
		if err != nil {
			exitWithError("could not marshal verify result", err)
		}
		err = os.WriteFile(outputFile, b, 0644)
		if err != nil {
			exitWithError("Failed to write output file", err)
		}
		logrus.Infoln("Success")
	},
}
