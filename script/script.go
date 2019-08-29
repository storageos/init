// Package script provides helpers to interact with scripts.
package script

import (
	"log"
	"os"
	"path/filepath"
)

// GetAllScripts takes a scripts directory path (absolute path) and scans it for
// script files, returning a list of all the scripts. It ignores files with docs
// extensions(.md, .txt).
func GetAllScripts(scriptsDir string) ([]string, error) {
	var ignoreFileExt = map[string]bool{
		".md":  true,
		".txt": true,
	}

	allScripts := []string{}

	err := filepath.Walk(scriptsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Do nothing with a directory.
			return nil
		}

		// Ignore non-script files.
		if _, exists := ignoreFileExt[filepath.Ext(path)]; exists {
			return nil
		}

		allScripts = append(allScripts, path)
		return nil
	})
	if err != nil {
		log.Printf("error walking the scripts dir path %v\n", err)
		return allScripts, err
	}

	return allScripts, nil
}
