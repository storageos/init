package runner

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"testing"
)

// update flag to update the golden files.
var update = flag.Bool("update", false, "update script stdout and stderr golden files")

// TestRunScript executes scripts and captures the stdout, stderr and exit
// status of the execution and compares them with the golden files in testdata
// dir. To update the golden files, pass -update flag to the test command.
// Stdout and stderr golden files have the extension .stdout and .stderr
// respectively.
// go test -v github.com/storageos/init/script/runner -run TestRunScript -update
func TestRunScript(t *testing.T) {
	testcases := []struct {
		name           string
		scriptName     string
		scriptArg      string
		envvars        map[string]string
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
		{
			name:       "script with env vars",
			scriptName: "envvar.sh",
			envvars: map[string]string{
				"FOOVAR": "fooval",
			},
			wantExitStatus: 0,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// Construct file path of the script and the golden files.
			scriptPath := fmt.Sprintf("testdata/%s", tc.scriptName)
			stdoutPath := fmt.Sprintf("%s.%s", scriptPath, "stdout")
			stderrPath := fmt.Sprintf("%s.%s", scriptPath, "stderr")

			run := NewRun()
			stdout, stderr, runErr := run.RunScript(scriptPath, tc.envvars, tc.scriptArg)

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
