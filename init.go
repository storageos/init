package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/storageos/init/info"
	"github.com/storageos/init/info/k8s"
	"github.com/storageos/init/script"
	"github.com/storageos/init/script/runner"

	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

const (
	daemonSetNameEnvVar      = "DAEMONSET_NAME"
	daemonSetNamespaceEnvVar = "DAEMONSET_NAMESPACE"
	nodeImageEnvVar          = "NODE_IMAGE"
)

const timeFormat = time.RFC3339

type logWriter struct{}

func (lw *logWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(time.Now().UTC().Format(timeFormat), " ", string(bytes))
}

func main() {
	log := log.New(new(logWriter), "", log.LstdFlags)
	log.SetFlags(0)

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
		var imageInfo info.ImageInfoer

		// This is in k8s environment.
		kubeclient, err := newK8SClient()
		if err != nil {
			log.Fatal(err)
		}

		// Create a k8s image info.
		name, namespace := getParamsForK8SImageInfo(*dsName, *dsNamespace)
		imageInfo = k8s.NewImageInfo(kubeclient).SetDaemonSet(name, namespace)

		// Get image.
		storageosImage, err = imageInfo.GetContainerImage(k8s.DefaultContainerName)
		if err != nil {
			log.Fatal(err)
		}
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
	run := runner.NewRun(log)

	// Run all the scripts.
	if err := runScripts(run, allScripts, scriptEnvVar); err != nil {
		log.Fatalf("init failed: %v", err)
	}
}

// NewK8SClient attempts to get k8s cluster configuration and return a new
// kubernetes client.
func newK8SClient() (kubernetes.Interface, error) {
	cfg, err := restclient.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(cfg)
}

// getParamsForK8SImageInfo returns the name and namespace to be used in k8s
// ImageInfo.
func getParamsForK8SImageInfo(dsName, dsNamespace string) (name, namespace string) {
	// If DaemonSet name is not provided, read from env var.
	if dsName == "" {
		dsName = os.Getenv(daemonSetNameEnvVar)

		// If DaemonSet name is still empty, use the default DaemonSet name.
		if dsName == "" {
			dsName = k8s.DefaultDaemonSetName
		}
	}

	// If DaemonSet namespace is not provided, read from env var.
	if dsNamespace == "" {
		dsNamespace = os.Getenv(daemonSetNamespaceEnvVar)

		// If DaemonSet namespace still empty, use the default StorageOS
		// deployment namespace.
		if dsNamespace == "" {
			dsNamespace = k8s.DefaultDaemonSetNamespace
		}
	}

	return dsName, dsNamespace
}

// runScripts takes a list of scripts and env vars, and runs the scripts
// sequentially. The error returned by the script execution is logged as k8s pod
// event.
// Any preliminary checks that need to be performed before running a script can
// be performed here.
func runScripts(run script.Runner, scripts []string, envVars map[string]string) error {
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
