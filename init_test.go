package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// update flag to update the golden files.
var update = flag.Bool("update", false, "update script stdout and stderr golden files")

// TestRunScript executes scripts and captures the stdout, stderr and exit
// status of the execution and compares them with the golden files in testdata
// dir. To update the golden files, pass -update flag to the test command.
// Stdout and stderr golden files have the extension .stdout and .stderr
// respectively.
// go test -v github.com/storageos/init -run TestRunScript -update
func TestRunScript(t *testing.T) {
	testcases := []struct {
		name           string
		scriptName     string
		scriptArg      string
		wantExitStatus int
	}{
		{
			name:           "successful script execution",
			scriptName:     "success.sh",
			wantExitStatus: 0,
		},
		{
			name:           "script with error",
			scriptName:     "error.sh",
			wantExitStatus: 3,
		},
		{
			name:           "script with an argument",
			scriptName:     "argument.sh",
			scriptArg:      "foo",
			wantExitStatus: 0,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// Construct file path of the script and the golden files.
			scriptPath := fmt.Sprintf("testdata/%s", tc.scriptName)
			stdoutPath := fmt.Sprintf("%s.%s", scriptPath, "stdout")
			stderrPath := fmt.Sprintf("%s.%s", scriptPath, "stderr")

			stdout, stderr, runErr := runScript(scriptPath, tc.scriptArg)

			// Update the golden files if update flag is specified.
			if *update {
				t.Log("update golden files")
				if err := ioutil.WriteFile(stdoutPath, stdout, 0644); err != nil {
					t.Fatalf("failed to update stdout golden file: %s", err)
				}
				if err := ioutil.WriteFile(stderrPath, stderr, 0644); err != nil {
					t.Fatalf("failed to update stderr golden file: %s", err)
				}
			}

			// Read the content of golden files.
			wantStdout, err := ioutil.ReadFile(stdoutPath)
			if err != nil {
				t.Fatalf("failed to read stdout golden file: %s", err)
			}

			wantStderr, err := ioutil.ReadFile(stderrPath)
			if err != nil {
				t.Fatalf("failed to read stderr golden file: %s", err)
			}

			// Compare the obtained results with the content of the golden
			// files.
			if !bytes.Equal(stdout, wantStdout) {
				t.Errorf("unexpected stdout:\n\t(WNT) %s\n\t(GOT) %s", string(wantStdout), string(stdout))
			}

			if !bytes.Equal(stderr, wantStderr) {
				t.Errorf("unexpected stderr:\n\t(WNT) %s\n\t(GOT) %s", string(wantStderr), string(stderr))
			}

			// Compare the exit status of the script.
			exitStatus := getExitStatus(runErr)
			if tc.wantExitStatus != exitStatus {
				t.Errorf("unexpected exit status:\n\t(WNT) %d\n\t(GOT) %d", tc.wantExitStatus, exitStatus)
			}
		})
	}
}

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

			scripts, err := getAllScripts(scriptsDir)
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
