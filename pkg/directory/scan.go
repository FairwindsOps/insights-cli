// Copyright 2020 FairwindsOps Inc
//
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

package directory

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

// ScanFolder looks through a given folder and returns a map[string][]string
// keyed on the OPA policy name, and the value listing files providing rego
// and V1 yaml instances for that policy.
func ScanFolder(folder string) (map[string][]string, error) {
	fileMap := map[string][]string{}
	regoFiles, err := getRegoFiles(folder)
	if err != nil {
		return nil, err
	}
	if len(regoFiles) > 0 {
		logrus.Debugf("processing all top-level .rego files in %s as 2 Insights OPA policies", folder)
		for _, rf := range regoFiles {
			rfWithoutExt := strings.TrimSuffix(rf, filepath.Ext(rf))
			rfWithDirectory := folder + "/" + rf
			fileMap[rfWithoutExt] = append(fileMap[rfWithoutExt], rfWithDirectory)
		}
		// return fileMap, nil
	}
	// logrus.Debugf("treating %s as a mix of V1 and V2 Insights OPA policies as there are no top-level .rego files", folder)
	err = filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".yaml" && !strings.HasPrefix(filepath.Base(path), "policy") {
			return nil
		}
		if filepath.Dir(path) == folder { // Ignore files not in a per-policy sub-dir
			return fmt.Errorf("File %s will not be processed because it is not in a per-policy sub-directory", path)
		}
		directoryName := filepath.Base(filepath.Dir(path))
		fileMap[directoryName] = append(fileMap[directoryName], path)
		return nil
	})
	logrus.Debugf("fileScan returning file-matp: %#v\n", fileMap)
	return fileMap, err
}

// getRegoFiles returns a slice of .rego files in the specified directory.
// This does not parse sub-directories.
func getRegoFiles(dirName string) ([]string, error) {
	regoFiles := make([]string, 0)
	files, err := ioutil.ReadDir(dirName)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if strings.ToLower(filepath.Ext(file.Name())) == ".rego" {
			regoFiles = append(regoFiles, file.Name())
		}
	}
	return regoFiles, nil
}
