package cli

import (
	"fmt"

	"github.com/fairwindsops/insights-cli/pkg/appgroups"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var overrideLocalFiles bool
var downloadAppGroupsSubDir string

func init() {
	downloadAppGroupsCmd.PersistentFlags().BoolVar(&overrideLocalFiles, "override", false, "use this flag to override local app-groups files.")
	downloadAppGroupsCmd.PersistentFlags().StringVar(&downloadAppGroupsSubDir, "download-subdirectory", "app-groups", "Sub-directory within download-directory, to download app-groups.")
	downloadCmd.AddCommand(downloadAppGroupsCmd)
}

var downloadAppGroupsCmd = &cobra.Command{
	Use:    "app-groups",
	Short:  "Download App Groups to local files.",
	Long:   "Download App Groups defined in Insights to local files.",
	Hidden: hideAppGroupCommands,
	PreRun: validateAndLoadInsightsAPIConfigWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		org := configurationObject.Options.Organization
		host := configurationObject.Options.Hostname
		appGroups, err := appgroups.FetchAppGroups(org, insightsToken, host)
		if err != nil {
			logrus.Fatalf("unable to fetch app-groups from insights: %v", err)
		}
		saveDir := downloadDir + "/" + downloadAppGroupsSubDir
		c, err := appgroups.SaveAppGroupsLocally(saveDir, appGroups, overrideLocalFiles)
		if err != nil {
			logrus.Fatalf("error saving app-groups locally: %v", err)
		}
		fmt.Printf("downloaded %d app-groups to %s\n", c, saveDir)
	},
}
