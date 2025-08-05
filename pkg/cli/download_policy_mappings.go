package cli

import (
	"fmt"

	"github.com/fairwindsops/insights-cli/pkg/policymappings"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var downloadPolicyMappingsSubDir string

func init() {
	downloadPolicyMappingsCmd.PersistentFlags().StringVar(&downloadPolicyMappingsSubDir, "download-subdirectory", "policy-mappings", "Sub-directory within download-directory, to download policy-mappings.")
	downloadCmd.AddCommand(downloadPolicyMappingsCmd)
}

var downloadPolicyMappingsCmd = &cobra.Command{
	Use:    "policy-mappings",
	Short:  "Download Policy Mappings to local files.",
	Long:   "Download Policy Mappings defined in Insights to local files.",
	PreRun: validateAndLoadInsightsAPIConfigWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		org := configurationObject.Options.Organization
		host := configurationObject.Options.Hostname
		policyMappings, err := policymappings.FetchPolicyMappings(client, org, insightsToken, host)
		if err != nil {
			logrus.Fatalf("unable to fetch policy-mappings from insights: %v", err)
		}
		saveDir := downloadDir + "/" + downloadPolicyMappingsSubDir
		c, err := saveEntitiesLocally(saveDir, policyMappings, overrideLocalFiles)
		if err != nil {
			logrus.Fatalf("error saving policy-mappings locally: %v", err)
		}
		fmt.Printf("downloaded %d policy-mappings to %s\n", c, saveDir)
	},
}
