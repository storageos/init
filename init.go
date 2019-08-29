package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/storageos/init/k8s"
	"github.com/storageos/init/k8s/inspector"
	"github.com/storageos/init/script"
	"github.com/storageos/init/script/runner"
)

const (
	daemonSetNameEnvVar       = "DAEMONSET_NAME"
	daemonSetNamespaceEnvVar  = "DAEMONSET_NAMESPACE"
	nodeImageEnvVar           = "NODE_IMAGE"
	defaultDaemonSetName      = "storageos-daemonset"
	defaultDaemonSetNamespace = "kube-system"
	defaultContainerName      = "storageos"
)

func main() {
	scriptsDir := flag.String("scripts", "", "absolute path of the scripts directory")
	dsName := flag.String("dsName", "", "name of the StorageOS DaemonSet")
	dsNamespace := flag.String("dsNamespace", "", "namespace of the StorageOS DaemonSet")
	nodeImage := flag.String("nodeImage", "", "container image of StorageOS Node, use when running out of k8s")

	flag.Parse()

	// StorageOS node container image.
	var storageosImage string

	// Abort if no scripts directory is provided.
	if *scriptsDir == "" {
		log.Println("no scripts directory specified, pass scripts dir with -scripts flag.")
		os.Exit(1)
	}

	// Attempt to get storageos node image.

	if *nodeImage == "" {
		// This is in k8s environment. Get the node image from StorageOS k8s
		// resources.
		kubeclient, err := k8s.NewK8SClient()
		if err != nil {
			log.Fatal(err)
		}

		// Create a k8s inspector.
		inspect := inspector.NewInspect(kubeclient)

		img, err := getImageFromK8S(inspect, *dsName, *dsNamespace)
		if err != nil {
			log.Fatal(err)
		}
		storageosImage = img
	} else {
		storageosImage = *nodeImage
	}

	// Abort if storageos node image is still unknown.
	if storageosImage == "" {
		log.Println("unknown storageos node image, pass node image with -nodeImage flag.")
		os.Exit(1)
	}

	// scriptEnvVar is the env vars passed to all the scripts.
	scriptEnvVar := map[string]string{}

	scriptEnvVar[nodeImageEnvVar] = storageosImage

	// Get list of all the scripts.
	allScripts, err := script.GetAllScripts(*scriptsDir)
	if err != nil {
		log.Fatalf("failed to get list of scripts: %v", err)
	}

	log.Println("scripts:", allScripts)

	// Create a script runner.
	run := runner.NewRun()

	// Run all the scripts.
	if err := runScripts(run, allScripts, scriptEnvVar); err != nil {
		log.Fatalf("init failed: %v", err)
	}
}

// getImageFromK8S fetches the StorageOS node container image from StorageOS
// running on k8s cluster.
func getImageFromK8S(inspect inspector.Inspector, dsName, dsNamespace string) (string, error) {
	// If DaemonSet name is not provided, read from env var.
	if dsName == "" {
		dsName = os.Getenv(daemonSetNameEnvVar)

		// If DaemonSet name is still empty, use the default DaemonSet name.
		if dsName == "" {
			dsName = defaultDaemonSetName
		}
	}

	// If DaemonSet namespace is not provided, read from env var.
	if dsNamespace == "" {
		dsNamespace = os.Getenv(daemonSetNamespaceEnvVar)

		// If DaemonSet namespace still empty, use the default StorageOS
		// deployment namespace.
		if dsNamespace == "" {
			dsNamespace = defaultDaemonSetNamespace
		}
	}

	return inspect.GetDaemonSetContainerImage(dsName, dsNamespace, defaultContainerName)
}

// runScripts takes a list of scripts and env vars, and runs the scripts
// sequentially. The error returned by the script execution is logged as k8s pod
// event.
// Any preliminary checks that need to be performed before running a script can
// be performed here.
func runScripts(run runner.Runner, scripts []string, envVars map[string]string) error {
	for _, script := range scripts {
		// TODO: Check if the script has any preliminary checks to be performed
		// before execution.

		log.Printf("exec: %s", script)

		_, stderr, err := run.RunScript(script, envVars)

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
