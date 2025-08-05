package appgroups

import (
	"bytes"
	"fmt"
	"os"
	"reflect"

	"github.com/fairwindsops/insights-cli/pkg/directory"
	"github.com/imroc/req/v3"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/xlab/treeprint"
	"gopkg.in/yaml.v2"
)

// AddAppGroupsBranch builds a tree for app-groups
func AddAppGroupsBranch(tree treeprint.Tree, appGroups []AppGroup) error {
	appGroupsBranch := tree.AddBranch("app-groups")
	for _, appGroup := range appGroups {
		appGroupsBranch.AddBranch(appGroup.Name)
	}
	return nil
}

// PushAppGroups pushes app-groups to insights
func PushAppGroups(client *req.Client, pushDir, org string, deleteMissing, dryRun bool) error {
	logrus.Debugln("Pushing app-groups")
	_, err := os.Stat(pushDir)
	if err != nil {
		return err
	}

	existingAppGroups, err := FetchAppGroups(client, org)
	if err != nil {
		return fmt.Errorf("error during API call: %w", err)
	}

	upserts, deletes, err := compareAppGroups(pushDir, existingAppGroups)
	if err != nil {
		return fmt.Errorf("unable to compare and push app-groups to Insights: %w", err)
	}

	for _, appGroup := range upserts {
		logrus.Infof("upsert app-group: %s", appGroup.Name)
		if !dryRun {
			err = upsertAppGroup(client, org, appGroup)
			if err != nil {
				return fmt.Errorf("error while upsert app-group %s to Fairwinds Insights: %w", appGroup.Name, err)
			}
		}
	}

	if deleteMissing {
		for _, appGroupForDelete := range deletes {
			logrus.Infof("Deleting app-group: %s", appGroupForDelete.Name)
			if !dryRun {
				err = deleteAppGroup(client, org, appGroupForDelete)
				if err != nil {
					return fmt.Errorf("error while deleting app-group %s from insights: %w", appGroupForDelete.Name, err)
				}
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
		return nil, nil, fmt.Errorf("error scanning directory: %w", err)

	}
	fileAppGroups, err := getAppGroupsFromFiles(files)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading app-groups from files: %w", err)

	}
	upserts, deletes = getAppGroupsDifferences(fileAppGroups, existingAppGroups)
	return upserts, deletes, nil
}

func getAppGroupsFromFiles(files map[string][]string) ([]AppGroup, error) {
	var appGroups []AppGroup
	for _, appGroupsFiles := range files {
		for _, filePath := range appGroupsFiles {
			content, err := os.ReadFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("error reading file %s: %w", filePath, err)
			}
			r := bytes.NewReader(content)
			dec := yaml.NewDecoder(r)
			var appGroup AppGroup
			for dec.Decode(&appGroup) == nil {
				if appGroup.Name == "" {
					return nil, fmt.Errorf("name is required in file %s", filePath)
				}
				if appGroup.Type == "" {
					return nil, fmt.Errorf("type is empty in file %s", filePath)
				}
				appGroups = append(appGroups, appGroup)
			}
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
