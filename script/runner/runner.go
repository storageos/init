// Package runner implements script runner.
package runner

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

const timeFormat = time.RFC3339

type logWriter struct{}

func (lw *logWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(time.Now().UTC().Format(timeFormat), " ", string(bytes))
}

// Run implements Runner interface.
type Run struct {
	lw *logWriter
}

// NewRun returns an initialized Run.
func NewRun() *Run {
	return &Run{}
}

// RunScript runs a given script with arguments if specified, and attaches a
// multiwriter to the stdout and stderr to write to the system pipes and to a
// buffer to collect the messages. The captured stdout and stderr messages are
// returned along with any error.
// The actual logs of the script are still written to the stdout and stderr.
func (r Run) RunScript(script string, env map[string]string, arg ...string) ([]byte, []byte, error) {
	cmd := exec.Command(script, arg...)
	// Add all env vars.
	for k, v := range env {
		cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%s", k, v))
	}

	var stdoutBuf, stderrBuf bytes.Buffer

	// Connect to commands stdout and stderr.
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()

	// Setup multi writer to write to stdout/stderr and the buffers.
	var errStdout, errStderr error
	stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
	stderr := io.MultiWriter(os.Stderr, &stderrBuf)

	if err := cmd.Start(); err != nil {
		log.Printf("Error while starting %q: %v", script, err)
		return nil, nil, err
	}

	var wg sync.WaitGroup
	wg.Add(1)

	// Wait until the stdout is completely copied.
	go func() {
		_, errStdout = io.Copy(stdout, stdoutIn)
		wg.Done()
	}()

	_, errStderr = io.Copy(stderr, stderrIn)
	wg.Wait()

	// Wait for the command to complete.
	if err := cmd.Wait(); err != nil {
		return stdoutBuf.Bytes(), stderrBuf.Bytes(), err
	}

	if errStdout != nil || errStderr != nil {
		log.Fatalf("failed to capture stdout and stderr\n")
		return nil, nil, fmt.Errorf("failed to capture stdout and stderr: %v, %v", errStdout, errStderr)
	}

	return stdoutBuf.Bytes(), stderrBuf.Bytes(), nil
}

// getExitStatus tries to extract exit status from error by type assertions.
// Returns 0 when no exit status value can be determined.
func getExitStatus(err error) int {
	if exiterr, ok := err.(*exec.ExitError); ok {
		// Extract exit information from the exit error.
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus()
		}
	}
	return 0
}
