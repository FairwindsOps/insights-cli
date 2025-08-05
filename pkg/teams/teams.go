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
	"net/http"
	"os"

	"github.com/imroc/req/v3"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"

	cliversion "github.com/fairwindsops/insights-cli/pkg/version"
)

const teamsPutURLFormat = "%s/v0/organizations/%s/teams-bulk"

type TeamInput struct {
	Clusters               []string `json:"clusters" yaml:"clusters"`
	DisallowedClusters     []string `json:"disallowedClusters" yaml:"disallowedClusters"`
	DisallowedNamespaces   []string `json:"disallowedNamespaces" yaml:"disallowedNamespaces"`
	DisallowedRepositories []string `json:"disallowedRepositories" yaml:"disallowedRepositories"`
	Name                   string   `json:"name" yaml:"name"`
	Namespaces             []string `json:"namespaces" yaml:"namespaces"`
	Repositories           []string `json:"repositories" yaml:"repositories"`
	AppGroups              []string `json:"appGroups" yaml:"appGroups"`
}

func PostTeams(client *req.Client, teamInput []TeamInput, deleteNonProvidedTeams bool, org, token, hostName string) error {
	url := fmt.Sprintf(teamsPutURLFormat, hostName, org)
	if deleteNonProvidedTeams {
		url += "?deleteNonProvidedTeams=true"
	}
	resp, err := client.R().SetHeaders(getHeaders(token)).SetBody(&teamInput).Post(url)
	if err != nil {
		return err
	}
	if resp.Response.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid HTTP response %d %s", resp.Response.StatusCode, string(resp.Bytes()))
	}
	return nil
}

func PushTeams(client *req.Client, pushDir, org, insightsToken, host string, deleteNonProvidedTeams, dryrun bool) error {
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
	defer teamsFile.Close()
	teams := []TeamInput{}
	b, err := os.ReadFile(teamsFileName)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(b, &teams)
	if err != nil {
		return err
	}
	if dryrun {
		logrus.Infof("Dry run: Would have pushed the following teams configuration:")
		for _, team := range teams {
			logrus.Infof("Team: %s", team.Name)
			logrus.Infof("  Clusters: %v", team.Clusters)
			logrus.Infof("  Namespaces: %v", team.Namespaces)
			logrus.Infof("  Repositories: %v", team.Repositories)
			logrus.Infof("  DisallowedClusters: %v", team.DisallowedClusters)
			logrus.Infof("  DisallowedNamespaces: %v", team.DisallowedNamespaces)
			logrus.Infof("  DisallowedRepositories: %v", team.DisallowedRepositories)
			logrus.Infof("  AppGroups: %v", team.AppGroups)
		}
		return nil
	}
	err = PostTeams(client, teams, deleteNonProvidedTeams, org, insightsToken, host)
	if err != nil {
		return err
	}
	logrus.Debugln("Done pushing teams configuration")
	return nil
}

func getHeaders(token string) map[string]string {
	return map[string]string{
		"Content-Type":            "application/yaml",
		"X-Fairwinds-CLI-Version": cliversion.GetVersion(),
		"Authorization":           fmt.Sprintf("Bearer %s", token),
		"Accept":                  "application/json",
	}
}
