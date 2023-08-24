package policymappings

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

// BuildPolicyMappingsTree builds a tree for policy-mapping
func BuildPolicyMappingsTree(policyMappings []PolicyMapping) (treeprint.Tree, error) {
	tree := treeprint.New()
	policyMappingsBranch := tree.AddBranch("policy-mappings")
	for _, policyMapping := range policyMappings {
		enabled := "enabled"
		if policyMapping.Spec.Enabled != nil && *policyMapping.Spec.Enabled == false {
			enabled = "disabled"
		}
		policyMappingsBranch.AddBranch(fmt.Sprintf("%s (%s)", policyMapping.Name, enabled))
	}
	return tree, nil
}

// PushPolicyMappings pushes policy-mapping to insights
func PushPolicyMappings(pushDir, org, insightsToken, host string, deleteMissing bool) error {
	logrus.Debugln("Pushing policy-mapping")
	_, err := os.Stat(pushDir)
	if err != nil {
		return err
	}

	existingPolicyMappings, err := FetchPolicyMappings(org, insightsToken, host)
	if err != nil {
		return fmt.Errorf("error during API call: %w", err)
	}

	upserts, deletes, err := comparePolicyMappings(pushDir, existingPolicyMappings)
	if err != nil {
		return fmt.Errorf("unable to compare and push policy-mapping to Insights: %w", err)
	}

	for _, policyMapping := range upserts {
		logrus.Infof("upsert policy-mapping: %s", policyMapping.Name)
		err = upsertPolicyMapping(org, insightsToken, host, policyMapping)
		if err != nil {
			return fmt.Errorf("Error while upsert policy-mapping %s to Fairwinds Insights: %w", policyMapping.Name, err)
		}
	}

	if deleteMissing {
		for _, policyMappingForDelete := range deletes {
			logrus.Infof("Deleting policy-mapping: %s", policyMappingForDelete.Name)
			err = deletePolicyMapping(org, insightsToken, host, policyMappingForDelete)
			if err != nil {
				return fmt.Errorf("Error while deleting policy-mapping %s from insights: %w", policyMappingForDelete.Name, err)
			}
		}
	}
	logrus.Debugln("Done pushing policy-mapping")
	return nil
}

// comparePolicyMappings compares a folder vs the policy-mapping returned by the API.
func comparePolicyMappings(folder string, existingPolicyMappings []PolicyMapping) (upserts, deletes []PolicyMapping, err error) {
	files, err := directory.ScanFolder(folder)
	if err != nil {
		return nil, nil, fmt.Errorf("error scanning directory: %w", err)

	}
	filePolicyMappings, err := getPolicyMappingsFromFiles(files)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading policy-mapping from files: %w", err)

	}
	upserts, deletes = getPolicyMappingsDifferences(filePolicyMappings, existingPolicyMappings)
	return upserts, deletes, nil
}

func getPolicyMappingsFromFiles(files map[string][]string) ([]PolicyMapping, error) {
	var policyMappings []PolicyMapping
	for _, policyMappingsFiles := range files {
		for _, filePath := range policyMappingsFiles {
			content, err := os.ReadFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("error reading file %s: %w", filePath, err)
			}
			r := bytes.NewReader(content)
			dec := yaml.NewDecoder(r)
			var policyMapping PolicyMapping
			for dec.Decode(&policyMapping) == nil {
				if policyMapping.Name == "" {
					return nil, fmt.Errorf("name is required in file %s", filePath)
				}
				if policyMapping.Type == "" {
					return nil, fmt.Errorf("type is empty in file %s", filePath)
				}
				policyMappings = append(policyMappings, policyMapping)
			}
		}
	}
	return policyMappings, nil
}

func getPolicyMappingsDifferences(filePolicyMappings, existingPolicyMappings []PolicyMapping) (upserts, deletes []PolicyMapping) {
	filePolicyMappingsByName := lo.KeyBy(filePolicyMappings, func(i PolicyMapping) string { return i.Name })
	existingPolicyMappingsByName := lo.KeyBy(existingPolicyMappings, func(i PolicyMapping) string { return i.Name })

	for name, filePolicyMapping := range filePolicyMappingsByName {
		if existingPolicyMapping, found := existingPolicyMappingsByName[name]; found {
			if !reflect.DeepEqual(filePolicyMapping, existingPolicyMapping) {
				// only update if the policy-mapping has changed
				upserts = append(upserts, filePolicyMapping)
			}
		} else {
			upserts = append(upserts, filePolicyMapping)
		}
	}

	for name, existingPolicyMapping := range existingPolicyMappingsByName {
		if _, ok := filePolicyMappingsByName[name]; !ok {
			deletes = append(deletes, existingPolicyMapping)
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

func SavePolicyMappingsLocally(saveDir string, policyMappings []PolicyMapping, overrideLocalFiles bool) (int, error) {
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

	err = purgeDirectory(saveDir)
	if err != nil {
		return 0, fmt.Errorf("could not purge directory %s: %w", saveDir, err)
	}

	var saved int
	for _, policyMapping := range policyMappings {
		filename := formatFilename(policyMapping.Name)
		filePath := saveDir + "/" + filename

		b, err := yaml.Marshal(policyMapping)
		if err != nil {
			return saved, fmt.Errorf("error marshalling policy-mapping %s: %w", policyMapping.Name, err)
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
