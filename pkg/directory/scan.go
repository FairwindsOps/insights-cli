// Copyright 2023 FairwindsOps Inc
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
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

// ScanOPAFolder looks through a given folder and returns a map[string][]string
// keyed on the OPA policy name, and the value listing files providing rego
// and V1 yaml instances for that policy.
func ScanOPAFolder(folder string) (map[string][]string, error) {
	fileMap := map[string][]string{}
	regoFiles, err := findRegoFilesOtherThanPolicy(folder)
	if err != nil {
		return nil, err
	}
	if len(regoFiles) > 0 {
		filesAreUnique, duplicateFiles := fileNamesAreUnique(regoFiles)
		if !filesAreUnique {
			duplicateFilesStr := prettyPrintDuplicateFiles(duplicateFiles)
			return nil, fmt.Errorf("rego file names must be unique when not named policy.rego, please resolve these %s", duplicateFilesStr)
		}
		logrus.Debugf("using the content of these files as V2 OPA policies: %v\n", regoFiles)
		for _, rf := range regoFiles {
			policyName := filepath.Base(strings.TrimSuffix(rf, filepath.Ext(rf)))
			fileMap[policyName] = append(fileMap[policyName], rf)
		}
	}
	// Now process policies in individual directories.
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
		if filepath.Dir(path) == folder { // Any top-level .rego files are already processed
			return nil
		}
		policyName := filepath.Base(filepath.Dir(path))
		fileMap[policyName] = append(fileMap[policyName], path)
		return nil
	})
	logrus.Debugf("OPA fileScan returning: %#v\n", fileMap)
	return fileMap, err
}

// ScanFolder looks through a given folder and returns a map[string][]string
// keyed on the directory name, and the value listing files.
func ScanFolder(folder string) (map[string][]string, error) {
	fileMap := map[string][]string{}
	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		directoryName := filepath.Base(filepath.Dir(path))
		fileMap[directoryName] = append(fileMap[directoryName], path)
		return nil
	})
	logrus.Debugf("scan-folder returning: %#v\n", fileMap)
	return fileMap, err
}

// findRegoFilesOtherThanPolicy returns a recursive list of .rego files in the
// given directory, other than the file `policy.rego`.
func findRegoFilesOtherThanPolicy(dir string) ([]string, error) {
	files := make([]string, 0)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		}
		ext := strings.ToLower(filepath.Ext(path))
		nameWithoutExt := strings.ToLower(filepath.Base(strings.TrimSuffix(path, filepath.Ext(path))))
		if ext == ".rego" && nameWithoutExt != "policy" {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// fileNamesAreUnique returns true when no file names repeat in the given
// list. For exampole, given input ["dir1/file.txt", "dir2/file.txt"] would
// return false, and duplicateNames of map["file.txt"]{"dir1", "dir2"}.
func fileNamesAreUnique(files []string) (alreadyUnique bool, duplicateNames map[string][]string) {
	duplicateNames = make(map[string][]string)
	// Initially populate duplicateNames with all filenames.
	for _, file := range files {
		fileWithoutDirectory := filepath.Base(file)
		duplicateNames[fileWithoutDirectory] = append(duplicateNames[fileWithoutDirectory], filepath.Dir(file))
	}
	// Iterate the above map and remove entries that have only one entry.
	// The result is only duplicate file names.
	for fileName, fullFileNames := range duplicateNames {
		if len(fullFileNames) == 1 {
			delete(duplicateNames, fileName)
		}
	}
	if len(duplicateNames) > 0 {
		return false, duplicateNames
	}
	return true, nil
}

// prettyPrintDuplicateFiles returns a nicely formatted string representation
// of the duplicateNames map[string][]string that is returned by the
// fileNamesAreUnique() function.
func prettyPrintDuplicateFiles(m map[string][]string) string {
	var message strings.Builder
	fmt.Fprintf(&message, "%d duplicate file", len(m))
	if len(m) > 1 {
		message.WriteString("s")
	}
	message.WriteString(": ")
	var n int = 1 // counter of keys processed
	for k, v := range m {
		if n == len(m) && len(m) == 2 {
			message.WriteString(" and ")
		}
		if n == len(m) && len(m) > 2 {
			message.WriteString("and ") // the comma logic will provide a leading space
		}
		fmt.Fprintf(&message, "%s found in %v", k, v)
		if n < len(m) && len(m) > 2 {
			message.WriteString(", ")
		}
		n++
	}
	return message.String()
}
