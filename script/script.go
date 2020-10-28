// Package script provides helpers to interact with scripts.
package script

//go:generate go run -mod=mod github.com/golang/mock/mockgen --build_flags=--mod=vendor -destination=../mocks/mock_runner.go -package=mocks github.com/storageos/init/script Runner

import (
	"log"
	"os"
	"path/filepath"

	_ "github.com/golang/mock/mockgen/model"
)

// Runner is an interface for script runner.
type Runner interface {
	// RunScript executes a script at path script, with environment variables
	// env and arguments arg, returning stdout and stderr byte slices and any
	// execution error.
	RunScript(script string, env map[string]string, arg ...string) (stdout []byte, stderr []byte, err error)
}

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
