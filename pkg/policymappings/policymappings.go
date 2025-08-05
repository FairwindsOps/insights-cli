package policymappings

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

// AddPolicyMappingsBranch builds a tree for policy-mapping
func AddPolicyMappingsBranch(tree treeprint.Tree, policyMappings []PolicyMapping) error {
	policyMappingsBranch := tree.AddBranch("policy-mappings")
	for _, policyMapping := range policyMappings {
		enabled := "enabled"
		if policyMapping.Spec.Enabled != nil && !*policyMapping.Spec.Enabled {
			enabled = "disabled"
		}
		block := "Block based on policy settings"
		if policyMapping.Spec.Block != nil {
			if *policyMapping.Spec.Block {
				block = "Always block on these policies"
			} else {
				block = "Never block on these policies"
			}
		}
		policyMappingsBranch.AddBranch(fmt.Sprintf("%s (%s) (%s)", policyMapping.Name, block, enabled))
	}
	return nil
}

// PushPolicyMappings pushes policy-mapping to insights
func PushPolicyMappings(client *req.Client, pushDir, org string, deleteMissing, dryRun bool) error {
	logrus.Debugln("Pushing policy-mapping")
	_, err := os.Stat(pushDir)
	if err != nil {
		return err
	}

	existingPolicyMappings, err := FetchPolicyMappings(client, org)
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
			err = upsertPolicyMapping(client, org, policyMapping)
			if err != nil {
				return fmt.Errorf("error while upsert policy-mapping %s to Fairwinds Insights: %w", policyMapping.Name, err)
			}
		}
	}

	if deleteMissing {
		for _, policyMappingForDelete := range deletes {
			logrus.Infof("Deleting policy-mapping: %s", policyMappingForDelete.Name)
			if !dryRun {
				err = deletePolicyMapping(client, org, policyMappingForDelete)
				if err != nil {
					return fmt.Errorf("error while deleting policy-mapping %s from insights: %w", policyMappingForDelete.Name, err)
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

func getHeaders() map[string]string {
	return map[string]string{
		"Accept":       "application/json",
		"Content-Type": "application/yaml",
	}
}
