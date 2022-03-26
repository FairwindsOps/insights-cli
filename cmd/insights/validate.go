/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package main

import (
	"fmt"
	"os"

	"github.com/fairwindsops/insights-cli/pkg/opavalidation"
	"github.com/spf13/cobra"
)

var regoFileName, objectFileName, objectNamespaceOverride string

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the syntax and output of a V2 Insights OPA policy",
	Long:  `Validate checks a V2 format Insights OPA policy for rego syntax and proper action item output.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := opavalidation.Run(regoFileName, objectFileName, objectNamespaceOverride)
		if err != nil {
			fmt.Printf("Policy failed validation: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Policy validated successfully.")
	},
}

func init() {
	policyCmd.AddCommand(validateCmd)
	// validateCmd.PersistentFlags().String("foo", "", "A help for foo")
	validateCmd.Flags().StringVarP(&regoFileName, "rego-file", "f", "policy.rego", "The rego file to validate.")
	validateCmd.Flags().StringVarP(&objectFileName, "object-file", "i", "", "A Kubernetes manifest to provide as input to the rego policy.")
	validateCmd.Flags().StringVarP(&objectNamespaceOverride, "object-namespace", "N", "", "A Kubernetes namespace to override any defined in the object-file.")
	validateCmd.MarkFlagRequired("object-file")
}
