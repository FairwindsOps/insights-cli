/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package main

import (
	"github.com/spf13/cobra"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:              "validate",
	Short:            "Validate files for use with Insights",
	Long:             `Validate files used with Insights, before submitting them to the Insights API`,
	TraverseChildren: true,
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
