/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package main

import (
	"fmt"
	"os"

	"github.com/fairwindsops/insights-cli/pkg/opavalidation"
	fwrego "github.com/fairwindsops/insights-plugins/plugins/opa/pkg/rego"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var regoFileName, objectFileName, batchDir, objectNamespaceOverride, insightsInfoCluster, insightsInfoContext string

// OPACmd represents the validate opa command
var OPACmd = &cobra.Command{
	Use:   "opa {-r <policy file> -k <Kube manifest file> | -b <directory of policies and manifests>} [flags]",
	Short: "Validate the syntax and output of a V2 Insights OPA policy",
	Long:  `opa runs V2 format Insights OPA policies with a Kubernetes manifest as input, validating rego syntax and proper Insights action item output.`,
	Example: `
	To validate a single policy: insights-cli validate opa -r policy.rego -k input-manifest.yaml

	To validate a directory of policies and Kubernetes manifests, with a policy and its corresponding Kubernetes manifest sharing the same base filename: insights-cli validate opa -b ./all_policies`,
	Run: func(cmd *cobra.Command, args []string) {
		if !checkValidateOPAFlags() {
			err := cmd.Help()
			if err != nil {
				logrus.Error(err)
			}
			os.Exit(1)
		}
		if regoFileName != "" {
			_, err := opavalidation.Run(regoFileName, objectFileName, fwrego.InsightsInfo{InsightsContext: insightsInfoContext, Cluster: insightsInfoCluster}, objectNamespaceOverride)
			if err != nil {
				fmt.Printf("OPA policy failed validation: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("OPA policy validated successfully.")
		}

		if batchDir != "" {
			_, failedPolicies, err := opavalidation.RunBatch(batchDir, fwrego.InsightsInfo{InsightsContext: insightsInfoContext, Cluster: insightsInfoCluster}, objectNamespaceOverride)
			fmt.Println() // separate output from RunBatch
			if err != nil {
				fmt.Printf("OPA policies failed validation: %v\n", err)
				fmt.Printf("Please check the above output for details about the %s\n", opavalidation.HumanizeStringsOutput(failedPolicies, "failure"))
				os.Exit(1)
			}
			fmt.Println("OPA policies validated successfully.")
		}
	},
}

// checkValidateOPAFlags verifies supplied flags for `validate opa` are valid.
func checkValidateOPAFlags() bool {
	if batchDir == "" && regoFileName == "" {
		logrus.Errorln("Please specify one of the --rego-file or --batch-directory options to validate one or more OPA policies.")
		return false
	}
	if batchDir != "" {
		if regoFileName != "" {
			logrus.Errorln("Please specify only one of the --batch-directory or --rego-file options.")
			return false
		}
		if objectFileName != "" {
			logrus.Errorln("The --kube-object-file option is only used with the --rego-file option, to validate a single OPA policy.")
			return false
		}
	}
	if regoFileName != "" && objectFileName == "" {
		logrus.Errorln("Please also specify a Kube object file when validating a single OPA policy, using the --kube-object-file option.")
		return false
	}
	return true
}

func init() {
	validateCmd.AddCommand(OPACmd)
	OPACmd.Flags().StringVarP(&batchDir, "batch-directory", "b", "", "A directory containing OPA policy .rego files and corresponding Kubernetes manifest .yaml input files to validate. This option validates multiple OPA policies at once, and is mutually exclusive with the rego-file option.")
	OPACmd.Flags().StringVarP(&regoFileName, "rego-file", "r", "", "An OPA policy file containing rego to validate. The --kube-object-file option is also required. This option validates a single policy, and is mutually exclusive with the batch-directory option.")
	OPACmd.Flags().StringVarP(&objectFileName, "kube-object-file", "k", "", "A Kubernetes manifest to provide as input when validating a single OPA policy. This option is mutually exclusive with the batch-directory option.")
	OPACmd.Flags().StringVarP(&objectNamespaceOverride, "object-namespace", "N", "", "A Kubernetes namespace to override any defined in the Kubernetes object being passed as input to an OPA policy.")
	OPACmd.Flags().StringVarP(&insightsInfoCluster, "insightsinfo-cluster", "l", "test", "A Kubernetes cluster name returned by the Insights-provided insightsinfo() rego function.")
	OPACmd.Flags().StringVarP(&insightsInfoContext, "insightsinfo-context", "t", "Agent", "An Insights context returned by the Insights-provided insightsinfo() rego function. The context returned by Insights plugins is typically one of: CI/CD, Admission, or Agent.")
}
