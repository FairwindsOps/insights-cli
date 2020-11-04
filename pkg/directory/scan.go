package directory

import (
	"os"
	"path/filepath"
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
		directoryName := filepath.Base(filepath.Dir(path))
		fileMap[directoryName] = append(fileMap[directoryName], path)
		return nil
	})
	return fileMap, err
}
