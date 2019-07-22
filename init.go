package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
)

func main() {
	scriptsDir := flag.String("scripts", "", "absolute path of the scripts directory")

	flag.Parse()

	if *scriptsDir == "" {
		log.Println("no scripts directory specified, pass scripts dir with -scripts flag.")
		os.Exit(1)
	}

	allScripts, err := getAllScripts(*scriptsDir)
	if err != nil {
		log.Fatalf("failed to get list of scripts: %v", err)
	}

	log.Println("scripts:", allScripts)

	if err := runScripts(allScripts); err != nil {
		log.Fatalf("init failed: %v", err)
	}
}

// getAllScripts takes a scripts directory path (absolute path) and scans it for
// script files, returning a list of all the scripts. It ignores files with docs
// extensions.
func getAllScripts(scriptsDir string) ([]string, error) {
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

// runScripts takes a list of scripts and runs then sequentially. The error
// returned by the script execution is logged as k8s pod event.
// Any preliminary checks that need to be performed before running a script can
// be performed here.
func runScripts(scripts []string) error {
	for _, script := range scripts {
		// TODO: Check if the script has any preliminary checks to be performed
		// before execution.

		log.Printf("exec: %s", script)

		_, stderr, err := runScript(script)

		// If stderr contains message, log and issue warning event.
		if len(stderr) > 0 {
			// log.Printf("[STDERR] %s: \n%s\n", script, string(stderr))
			// Create k8s warning event.
			// Issue a warning event with the stderr log.
		}

		if err != nil {
			// Create a k8s failure events.

			return fmt.Errorf("script %q failed: %v", script, err)
		}
	}

	return nil
}

// runScript runs a given script with arguments if specified, and attaches a
// multiwriter to the stdout and stderr to write to the system pipes and to a
// buffer to collect the messages. The captured stdout and stderr messages are
// returned along with any error.
// The actual logs of the script are still written to the stdout and stderr.
func runScript(script string, arg ...string) ([]byte, []byte, error) {
	cmd := exec.Command(script, arg...)

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

	// Get the script execution exit status from the returned error.
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
