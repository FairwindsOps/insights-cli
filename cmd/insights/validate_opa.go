/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package main

import (
	"fmt"
	"os"

	"github.com/fairwindsops/insights-cli/pkg/opavalidation"
	fwrego "github.com/fairwindsops/insights-plugins/plugins/opa/pkg/rego"
	"github.com/spf13/cobra"
)

var regoFileName, objectFileName, objectNamespaceOverride, insightsInfoCluster, insightsInfoContext string

// opaCmd represents the validate opa command
var opaCmd = &cobra.Command{
	Use:              "opa",
	Short:            "Validate the syntax and output of a V2 Insights OPA policy",
	Long:             `opa checks a V2 format Insights OPA policy for rego syntax and proper action item output.`,
	TraverseChildren: true,
	Run: func(cmd *cobra.Command, args []string) {
		err := opavalidation.Run(regoFileName, objectFileName, fwrego.InsightsInfo{InsightsContext: insightsInfoContext, Cluster: insightsInfoCluster}, objectNamespaceOverride)
		if err != nil {
			fmt.Printf("OPA policy failed validation: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("OPA policy validated successfully.")
	},
}

func init() {
	validateCmd.AddCommand(opaCmd)
	opaCmd.Flags().StringVarP(&regoFileName, "rego-file", "r", "policy.rego", "The rego file to validate.")
	opaCmd.Flags().StringVarP(&objectFileName, "kube-object-file", "k", "", "A Kubernetes manifest to provide as input to the rego policy.")
	opaCmd.Flags().StringVarP(&objectNamespaceOverride, "object-namespace", "N", "", "A Kubernetes namespace to override any defined in the object-file.")
	opaCmd.Flags().StringVarP(&insightsInfoCluster, "insightsinfo-cluster", "l", "test", "A Kubernetes cluster name returned by the Insights-provided insightsinfo() rego function.")
	opaCmd.Flags().StringVarP(&insightsInfoContext, "insightsinfo-context", "t", "test", "An Insights context returned by the Insights-provided insightsinfo() rego function. The context returned by Insights is typially one of  CI/CD, Admission, or Agent.")
	err := opaCmd.MarkFlagRequired("kube-object-file")
	if err != nil {
		panic(fmt.Errorf("unable to mark flag persistent: %v", err))
	}
}
