// Copyright 2024 FairwindsOps Inc //
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package policies provides an interface to Insights policy settings and
// enforcement.
package teams

import (
	"errors"
	"fmt"
	"os"

	"github.com/imroc/req/v3"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"
)

const teamsPutURLFormat = "/v0/organizations/%s/teams-bulk"
const teamsGetURLFormat = "/v0/organizations/%s/teams"

type TeamInput struct {
	Name                   string   `json:"name" yaml:"name"`
	Clusters               []string `json:"clusters" yaml:"clusters"`
	Namespaces             []string `json:"namespaces" yaml:"namespaces"`
	Repositories           []string `json:"repositories" yaml:"repositories"`
	DisallowedClusters     []string `json:"disallowedClusters" yaml:"disallowedClusters"`
	DisallowedNamespaces   []string `json:"disallowedNamespaces" yaml:"disallowedNamespaces"`
	DisallowedRepositories []string `json:"disallowedRepositories" yaml:"disallowedRepositories"`
	AppGroups              []string `json:"appGroups" yaml:"appGroups"`
}

type TeamOutput = TeamInput

func PostTeams(client *req.Client, teamInput []TeamInput, deleteNonProvidedTeams bool, org string) error {
	url := fmt.Sprintf(teamsPutURLFormat, org)
	if deleteNonProvidedTeams {
		url += "?deleteNonProvidedTeams=true"
	}
	resp, err := client.R().SetHeaders(getHeaders()).SetBody(&teamInput).Post(url)
	if err != nil {
		return err
	}
	if resp.IsErrorState() {
		return fmt.Errorf("invalid HTTP response %d %s", resp.StatusCode, string(resp.Bytes()))
	}
	return nil
}

func ListTeams(client *req.Client, org string) ([]TeamOutput, error) {
	url := fmt.Sprintf(teamsGetURLFormat, org)
	resp, err := client.R().SetHeaders(getHeaders()).Get(url)
	if err != nil {
		return nil, err
	}
	if resp.IsErrorState() {
		return nil, fmt.Errorf("invalid HTTP response %d %s", resp.StatusCode, string(resp.Bytes()))
	}
	teams := []TeamOutput{}
	err = resp.Unmarshal(&teams)
	if err != nil {
		return nil, fmt.Errorf("unable to convert response to json for teams: %w", err)
	}
	return teams, nil
}

func PushTeams(client *req.Client, pushDir, org string, deleteNonProvidedTeams, dryRun bool) error {
	if pushDir == "" {
		return errors.New("pushDir cannot be empty")
	}

	teamsFileName := pushDir + "/teams.yaml"
	logrus.Infof("Pushing teams configuration from %s", teamsFileName)
	_, err := os.Stat(teamsFileName)
	if err != nil {
		return err
	}

	teamsFile, err := os.Open(teamsFileName)
	if err != nil {
		return err
	}

	defer func() {
		if err := teamsFile.Close(); err != nil {
			logrus.Errorf("error closing teams file: %v", err)
		}
	}()

	localTeams := []TeamInput{}
	b, err := os.ReadFile(teamsFileName)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(b, &localTeams)
	if err != nil {
		return err
	}

	if dryRun {
		logrus.Infof("Dry run: Would have pushed the following teams configuration:")
		for _, team := range localTeams {
			logrus.Infof("Team: %s", team.Name)
			logrus.Infof("  Clusters: %v", team.Clusters)
			logrus.Infof("  Namespaces: %v", team.Namespaces)
			logrus.Infof("  Repositories: %v", team.Repositories)
			logrus.Infof("  DisallowedClusters: %v", team.DisallowedClusters)
			logrus.Infof("  DisallowedNamespaces: %v", team.DisallowedNamespaces)
			logrus.Infof("  DisallowedRepositories: %v", team.DisallowedRepositories)
			logrus.Infof("  AppGroups: %v", team.AppGroups)
		}

		if deleteNonProvidedTeams {
			logrus.Infof("--------------------------------")

			remoteTeams, err := ListTeams(client, org)
			if err != nil {
				return fmt.Errorf("error listing teams: %w", err)
			}
			remoteTeamsByName := lo.KeyBy(remoteTeams, func(i TeamOutput) string { return i.Name })
			localTeamsByName := lo.KeyBy(localTeams, func(i TeamInput) string { return i.Name })

			teamsToBeDeleted := []string{}
			for name := range remoteTeamsByName {
				if _, ok := localTeamsByName[name]; !ok {
					teamsToBeDeleted = append(teamsToBeDeleted, name)
				}
			}

			if len(teamsToBeDeleted) == 0 {
				logrus.Infof("Dry run: Would have deleted no teams")
				return nil
			}

			logrus.Infof("Dry run: Would have deleted the following teams:")
			for _, team := range teamsToBeDeleted {
				logrus.Infof("Team: %s", team)
			}
		}

		return nil
	}

	err = PostTeams(client, localTeams, deleteNonProvidedTeams, org)
	if err != nil {
		return err
	}

	logrus.Debugln("Done pushing teams configuration")
	return nil
}

func getHeaders() map[string]string {
	return map[string]string{
		"Content-Type": "application/yaml",
		"Accept":       "application/json",
	}
}
