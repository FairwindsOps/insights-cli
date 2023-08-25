package policymappings

import (
	"bytes"
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

// BuildPolicyMappingsTree builds a tree for policy-mapping
func BuildPolicyMappingsTree(policyMappings []PolicyMapping) (treeprint.Tree, error) {
	tree := treeprint.New()
	policyMappingsBranch := tree.AddBranch("policy-mappings")
	for _, policyMapping := range policyMappings {
		enabled := "enabled"
		if policyMapping.Spec.Enabled != nil && !*policyMapping.Spec.Enabled {
			enabled = "disabled"
		}
		policyMappingsBranch.AddBranch(fmt.Sprintf("%s (%s)", policyMapping.Name, enabled))
	}
	return tree, nil
}

// PushPolicyMappings pushes policy-mapping to insights
func PushPolicyMappings(pushDir, org, insightsToken, host string, deleteMissing, dryRun bool) error {
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
		if !dryRun {
			err = upsertPolicyMapping(org, insightsToken, host, policyMapping)
			if err != nil {
				return fmt.Errorf("Error while upsert policy-mapping %s to Fairwinds Insights: %w", policyMapping.Name, err)
			}
		}
	}

	if deleteMissing {
		for _, policyMappingForDelete := range deletes {
			logrus.Infof("Deleting policy-mapping: %s", policyMappingForDelete.Name)
			if !dryRun {
				err = deletePolicyMapping(org, insightsToken, host, policyMappingForDelete)
				if err != nil {
					return fmt.Errorf("Error while deleting policy-mapping %s from insights: %w", policyMappingForDelete.Name, err)
				}
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
