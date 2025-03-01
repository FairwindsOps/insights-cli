/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cli

import (
	"fmt"
	"os"

	"github.com/fairwindsops/insights-cli/pkg/opavalidation"
	fwrego "github.com/fairwindsops/insights-plugins/plugins/opa/pkg/rego"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var regoFileName, objectFileName, batchDir, libsDir, objectNamespaceOverride, insightsInfoCluster, insightsInfoContext, regoVersion string
var expectActionItem opavalidation.ExpectActionItemOptions

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
			_, err := opavalidation.Run(regoVersion, regoFileName, objectFileName, expectActionItem, fwrego.InsightsInfo{InsightsContext: insightsInfoContext, Cluster: insightsInfoCluster}, objectNamespaceOverride, libsDir)
			if err != nil {
				fmt.Printf("OPA policy failed validation: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("OPA policy validated successfully.")
		}

		if batchDir != "" {
			_, failedPolicies, err := opavalidation.RunBatch(regoVersion, batchDir, expectActionItem, fwrego.InsightsInfo{InsightsContext: insightsInfoContext, Cluster: insightsInfoCluster}, objectNamespaceOverride, libsDir)
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
	OPACmd.Flags().StringVarP(&objectFileName, "kube-object-file", "k", "", "A Kubernetes manifest to provide as input when validating a single OPA policy. This option is mutually exclusive with the batch-directory option. A manifest file ending in a .success.yaml extension is expected to return 0 action items. A manifest file ending in a .failure.yaml extension is expected to output one action item. See also the --expect-action-item option.")
	OPACmd.Flags().StringVarP(&expectActionItem.SuccessFileExtension, "kube-manifest-success-ext", "e", ".success.yaml", "The extension for a Kubernetes manifest file name which, if found, indicates an OPA policy is NOT expected to return an action item.")
	OPACmd.Flags().StringVarP(&expectActionItem.FailureFileExtension, "kube-manifest-failure-ext", "E", ".failure.yaml", "The extension for a Kubernetes manifest file name which, if found, indicates an OPA policy is expected to return an action item.")
	OPACmd.Flags().StringVarP(&objectNamespaceOverride, "object-namespace", "N", "", "A Kubernetes namespace to override any defined in the Kubernetes object being passed as input to an OPA policy.")
	OPACmd.Flags().StringVarP(&insightsInfoCluster, "insightsinfo-cluster", "l", "test", "A Kubernetes cluster name returned by the Insights-provided insightsinfo() rego function.")
	OPACmd.Flags().StringVarP(&insightsInfoContext, "insightsinfo-context", "t", "Agent", "An Insights context returned by the Insights-provided insightsinfo() rego function. The context returned by Insights plugins is typically one of: CI/CD, Admission, or Agent.")
	OPACmd.Flags().StringVarP(&libsDir, "libs-dir", "L", "", "A directory containing additional rego libraries to load. This option is not required, but can be used to load additional rego libraries.")
	OPACmd.Flags().BoolVarP(&expectActionItem.Default, "expect-action-item", "i", true, "Whether to expect the OPA policy to output one action item (true) or 0 action items (false). This option is applied to Kubernetes manifest files with no .success.yaml nor .failure.yaml extension.")
	OPACmd.Flags().StringVarP(&regoVersion, "rego-version", "v", "v0", "The version of the rego policy to validate. This option is not required, but can be used to specify the rego version to validate. Version can be v0 or v1")
}
