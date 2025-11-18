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

package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var downloadDir string
var overrideLocalFiles bool

func init() {
	downloadCmd.PersistentFlags().BoolVar(&overrideLocalFiles, "override", false, "use this flag to override local files.")
	downloadCmd.PersistentFlags().StringVarP(&downloadDir, "download-directory", "d", ".", "target directory to download content pulled from Insights.")
	rootCmd.AddCommand(downloadCmd)
}

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download custom configuration resources from Insights",
	Long:  "Download custom configuration such as App Groups from Insights.",
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Error("Please specify a sub-command.")
		err := cmd.Help()
		if err != nil {
			logrus.Error(err)
		}
		os.Exit(1)
	},
}

type nameable interface {
	GetName() string
	GetYamlBytes() ([]byte, error)
}

var filenameRegex = regexp.MustCompile("[^A-Za-z0-9]+")

func saveEntitiesLocally[T nameable](saveDir string, entities []T, overrideLocalFiles bool, skipFilesPatterns []string) (int, error) {
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

	err = purgeDirectory(saveDir, skipFilesPatterns)
	if err != nil {
		return 0, fmt.Errorf("could not purge directory %s: %w", saveDir, err)
	}

	var saved int
	for _, e := range entities {
		filename := formatFilename(e.GetName())
		filePath := saveDir + "/" + filename

		b, err := e.GetYamlBytes()
		if err != nil {
			return saved, fmt.Errorf("error getting yaml bytes for entity %s: %w", e.GetName(), err)
		}
		err = os.WriteFile(filePath, b, 0644)
		if err != nil {
			return saved, fmt.Errorf("error writing file %s: %w", filePath, err)
		}
		saved++
	}
	return saved, nil
}

// remove all contents of a directory except files matching skipFilesPatterns
func purgeDirectory(saveDir string, skipFilesPatterns []string) error {
	// If no skip patterns, remove everything
	if len(skipFilesPatterns) == 0 {
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

	// Read all files in the directory
	entries, err := os.ReadDir(saveDir)
	if err != nil {
		return fmt.Errorf("error reading directory %s: %w", saveDir, err)
	}

	// Check each entry and delete if it doesn't match any skip pattern
	for _, entry := range entries {
		entryPath := filepath.Join(saveDir, entry.Name())

		// Check if the entry matches any skip pattern
		shouldSkip := false
		for _, pattern := range skipFilesPatterns {
			if strings.Contains(entry.Name(), pattern) {
				shouldSkip = true
				break
			}
		}

		// Delete if it doesn't match any skip pattern
		if !shouldSkip {
			err := os.RemoveAll(entryPath)
			if err != nil {
				return fmt.Errorf("error removing %s: %w", entryPath, err)
			}
		}
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
	defer func() {
		if err := f.Close(); err != nil {
			logrus.Errorf("error closing file %s: %v", path, err)
		}
	}()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}
