package script

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestGetAllScripts(t *testing.T) {
	testcases := []struct {
		name               string
		files              []string
		wantScriptsInOrder []string
	}{
		{
			name: "dir with files only",
			files: []string{
				"script3.sh",
				"script1.sh",
				"script4.sh",
				"script2.sh",
			},
			wantScriptsInOrder: []string{
				"script1.sh",
				"script2.sh",
				"script3.sh",
				"script4.sh",
			},
		},
		{
			// The scripts are sorted based on the parent dir name. Scripts
			// under foo/ will appear first in the script list and scripts under
			// zoo/ will appear towards the end.
			name: "scripts under subdirectory",
			files: []string{
				"foo/zzzscript100.sh",
				"script4.sh",
				"zoo/a1.sh",
				"script3.sh",
				"zoo/a0.sh",
				"foo/zzzscript99.sh",
			},
			wantScriptsInOrder: []string{
				"zzzscript100.sh",
				"zzzscript99.sh",
				"script3.sh",
				"script4.sh",
				"a0.sh",
				"a1.sh",
			},
		},
		{
			name:               "no script",
			files:              []string{},
			wantScriptsInOrder: []string{},
		},
		{
			name: "ignore non-script files",
			files: []string{
				"foo/script10.sh",
				"foo/readme.md",
				"foo/license.txt",
				"aaa/aa-script.sh",
				"aaa/warning.txt",
			},
			wantScriptsInOrder: []string{
				"aa-script.sh",
				"script10.sh",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// Create scripts dir.
			scriptsDir, err := ioutil.TempDir("", "init-scripts-test")
			if err != nil {
				t.Fatalf("failed to create directory: %v", err)
			}

			// Create the test script files.
			for _, script := range tc.files {
				// Check if the script is in a subdirectory and create the
				// directory.
				parentDir := filepath.Dir(script)
				if parentDir != "" {
					// Join path with the temp dir and create the dir.
					absPath := filepath.Join(scriptsDir, parentDir)
					if err := os.MkdirAll(absPath, 0777); err != nil {
						t.Fatalf("failed to create sub directory: %v", err)
					}
				}

				// Create the script file.
				_, err := os.Create(filepath.Join(scriptsDir, script))
				if err != nil {
					t.Fatalf("failed to create script file: %v", err)
				}
			}

			scripts, err := GetAllScripts(scriptsDir)
			if err != nil {
				t.Fatalf("failed to get all scripts: %v", err)
			}

			if len(scripts) != len(tc.wantScriptsInOrder) {
				t.Errorf("unexpected number of scripts:\n\t(WNT) %d\n\t(GOT) %d", len(tc.wantScriptsInOrder), len(scripts))
			}

			// Check the scripts are in the expected order.
			for index, script := range scripts {
				if tc.wantScriptsInOrder[index] != filepath.Base(script) {
					t.Errorf("unexpected script order at position %d:\n\t(WNT) %s\n\t(GOT) %s", index, tc.wantScriptsInOrder[index], filepath.Base(script))
				}
			}

			// Cleanup.
			defer os.RemoveAll(scriptsDir)
		})
	}
}
