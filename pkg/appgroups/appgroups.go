package appgroups

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"regexp"

	"github.com/fairwindsops/insights-cli/pkg/directory"
	cliversion "github.com/fairwindsops/insights-cli/pkg/version"
	"github.com/imroc/req"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/xlab/treeprint"
	"gopkg.in/yaml.v2"
)

var filenameRegex = regexp.MustCompile("[^A-Za-z0-9]+")

// BuildAppGroupsTree builds a tree for app-groups
func BuildAppGroupsTree(appGroups []AppGroup) (treeprint.Tree, error) {
	tree := treeprint.New()
	appGroupsBranch := tree.AddBranch("app-groups")
	for _, appGroup := range appGroups {
		appGroupsBranch.AddBranch(appGroup.Name)
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

	existingAppGroups, err := FetchAppGroups(org, insightsToken, host)
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

func getHeaders(token string) req.Header {
	return req.Header{
		"X-Fairwinds-CLI-Version": cliversion.GetVersion(),
		"Authorization":           fmt.Sprintf("Bearer %s", token),
		"Accept":                  "application/json",
	}
}

func SaveAppGroupsLocally(saveDir string, appGroups []AppGroup, overrideLocalFiles bool) (int, error) {
	_, err := os.Stat(saveDir)
	if err != nil {
		return 0, err
	}
	isEmpty, err := IsEmpty(saveDir)
	if err != nil {
		return 0, fmt.Errorf("error checking if directory %s is empty: %w", saveDir, err)
	}
	if !isEmpty && !overrideLocalFiles {
		logrus.Warnf("directory %s must be empty, use --override to override local files", saveDir)
		return 0, nil
	}

	purgeDirectory(saveDir)

	var saved int
	for _, appGroup := range appGroups {
		filename := formatFilename(appGroup.Name)
		filePath := saveDir + "/" + filename

		b, err := yaml.Marshal(appGroup)
		if err != nil {
			return saved, fmt.Errorf("error marshalling app-group %s: %w", appGroup.Name, err)
		}
		err = os.WriteFile(filePath, b, 0644)
		if err != nil {
			return saved, fmt.Errorf("error writing file %s: %w", filePath, err)
		}
		saved++
	}
	return saved, nil
}

// remove all contents of a directory and creates it again
func purgeDirectory(saveDir string) error {
	err := os.RemoveAll(saveDir)
	if err != nil {
		return fmt.Errorf("error clearing directory %s: %w", saveDir, err)
	}
	err = os.MkdirAll(saveDir, 0755)
	if err != nil {
		return fmt.Errorf("error creating directory %s: %w", saveDir, err)
	}
	return nil
}

func formatFilename(name string) string {
	return fmt.Sprintf("%s.yaml", filenameRegex.ReplaceAllString(name, "-"))
}

func IsEmpty(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}
