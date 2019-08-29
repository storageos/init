// Package inspector helps inspect k8s cluster resources.
package inspector

//go:generate mockgen -destination=../../mocks/mock_inspector.go -package=mocks github.com/storageos/init/k8s/inspector Inspector

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Inspector is an interface for k8s inspector. This can be used to query
// various k8s related informations.
type Inspector interface {
	GetDaemonSetContainerImage(daemonSetName, daemonSetNamespace, containerName string) (string, error)
}

// Inspect implements the Inspector interface.
type Inspect struct {
	client kubernetes.Interface
}

// NewInspect returns an initialized Inspect.
func NewInspect(client kubernetes.Interface) *Inspect {
	return &Inspect{client: client}
}

// GetDaemonSetContainerImage returns the container image name of a given
// container in a DaemonSet Pod.
func (i Inspect) GetDaemonSetContainerImage(dsName, dsNamespace, containerName string) (string, error) {
	daemonset, err := i.client.AppsV1().DaemonSets(dsNamespace).Get(dsName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	// Find the target container by name.
	for _, container := range daemonset.Spec.Template.Spec.Containers {
		if container.Name == containerName {
			return container.Image, nil
		}
	}

	return "", fmt.Errorf("failed to find container %q in daemonset %q", containerName, dsName)
}
