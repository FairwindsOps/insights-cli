package appgroups

import (
	"fmt"
	"os"
	"reflect"

	"github.com/fairwindsops/insights-cli/pkg/directory"
	cliversion "github.com/fairwindsops/insights-cli/pkg/version"
	"github.com/imroc/req"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/xlab/treeprint"
	"gopkg.in/yaml.v2"
)

// BuildAppGroupsTree builds a tree for app-groups
func BuildAppGroupsTree(org, token, hostName string) (treeprint.Tree, error) {
	appGroups, err := fetchAppGroups(org, token, hostName)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch app-groups from insights: %v", err)
	}
	tree := treeprint.New()
	appGroupsBranch := tree.AddBranch("app-groups")
	for _, appGroup := range appGroups {
		appGroupsNode := appGroupsBranch.AddBranch(appGroup.Name)
		if appGroup.Spec.Enabled != nil {
			appGroupsNode.AddNode(fmt.Sprintf("Enabled: %t", *appGroup.Spec.Enabled))
		}
	}
	return tree, nil
}

// PushAppGroups pushes app-groups to insights
func PushAppGroups(pushDir, org, insightsToken, host string, deleteMissing bool) error {
	logrus.Debugln("Pushing app-groups")
	_, err := os.Stat(pushDir)
	if err != nil {
		return err
	}

	existingAppGroups, err := fetchAppGroups(org, insightsToken, host)
	if err != nil {
		return fmt.Errorf("error during API call: %w", err)
	}

	upserts, deletes, err := compareAppGroups(pushDir, existingAppGroups)
	if err != nil {
		return fmt.Errorf("unable to compare and push app-groups to Insights: %w", err)
	}

	for _, appGroup := range upserts {
		logrus.Infof("upsert app-group: %s", appGroup.Name)
		err = upsertAppGroup(org, insightsToken, host, appGroup)
		if err != nil {
			return fmt.Errorf("Error while upsert app-group %s to Fairwinds Insights: %w", appGroup.Name, err)
		}
	}

	if deleteMissing {
		for _, appGroupForDelete := range deletes {
			logrus.Infof("Deleting app-group: %s", appGroupForDelete.Name)
			err = deleteAppGroup(org, insightsToken, host, appGroupForDelete)
			if err != nil {
				return fmt.Errorf("Error while deleting app-group %s from insights: %w", appGroupForDelete.Name, err)
			}
		}
	}
	logrus.Debugln("Done pushing app-groups")
	return nil
}

// compareAppGroups compares a folder vs the app-groups returned by the API.
func compareAppGroups(folder string, existingAppGroups []AppGroup) (upserts, deletes []AppGroup, err error) {
	files, err := directory.ScanFolder(folder)
	if err != nil {
		logrus.Error("Error scanning directory")
		return nil, nil, err
	}
	fileAppGroups, err := getAppGroupsFromFiles(files)
	if err != nil {
		logrus.Error("Error reading checks from files")
		return nil, nil, err
	}
	upserts, deletes = getAppGroupsDifferences(fileAppGroups, existingAppGroups)
	return upserts, deletes, nil
}

func getAppGroupsFromFiles(files map[string][]string) ([]AppGroup, error) {
	var appGroups []AppGroup
	for _, appGroupsFiles := range files {
		for _, filePath := range appGroupsFiles {
			fileContents, err := os.ReadFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("error reading file %s: %w", filePath, err)
			}
			var appGroup AppGroup
			err = yaml.Unmarshal(fileContents, &appGroup)
			if err != nil {
				return nil, fmt.Errorf("Error unmarshaling check YAML %s: %w", filePath, err)
			}
			if appGroup.Name == "" {
				return nil, fmt.Errorf("name is required in file %s", filePath)
			}
			if appGroup.Type == "" {
				return nil, fmt.Errorf("type is empty in file %s", filePath)
			}
			appGroups = append(appGroups, appGroup)
		}
	}
	return appGroups, nil
}

func getAppGroupsDifferences(fileAppGroups, existingAppGroups []AppGroup) (upserts, deletes []AppGroup) {
	fileAppGroupsByName := lo.KeyBy(fileAppGroups, func(i AppGroup) string { return i.Name })
	existingAppGroupsByName := lo.KeyBy(existingAppGroups, func(i AppGroup) string { return i.Name })

	for name, fileAppGroup := range fileAppGroupsByName {
		if existingAppGroup, found := existingAppGroupsByName[name]; found {
			if !reflect.DeepEqual(fileAppGroup, existingAppGroup) {
				// only update if the app-group has changed
				upserts = append(upserts, fileAppGroup)
			}
		} else {
			upserts = append(upserts, fileAppGroup)
		}
	}

	for name, existingAppGroup := range existingAppGroupsByName {
		if _, ok := fileAppGroupsByName[name]; !ok {
			deletes = append(deletes, existingAppGroup)
		}
	}
	return upserts, deletes
}

func getHeaders(token string) req.Header {
	return req.Header{
		"X-Fairwinds-CLI-Version": cliversion.GetVersion(),
		"Authorization":           fmt.Sprintf("Bearer %s", token),
		"Accept":                  "application/json",
	}
}
