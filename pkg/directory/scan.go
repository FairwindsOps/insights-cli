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
	"os"
	"path/filepath"
	"strings"
)

// ScanFolder looks through a given folder and returns all of the files found
func ScanFolder(folder string) (map[string][]string, error) {
	fileMap := map[string][]string{}
	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".yaml" && !strings.HasPrefix(filepath.Base(path), "policy") {
			return nil
		}
		directoryName := filepath.Base(filepath.Dir(path))
		fileMap[directoryName] = append(fileMap[directoryName], path)
		return nil
	})
	return fileMap, err
}
