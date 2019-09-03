package k8s

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	// DefaultContainerName is the default StorageOS node container name.
	DefaultContainerName = "storageos"
	// DefaultDaemonSetName is the name of the default DaemonSet used to deploy
	// StorageOS.
	DefaultDaemonSetName = "storageos-daemonset"
	// DefaultDaemonSetNamespace is the namespace where StorageOS is deployed by
	// default.
	DefaultDaemonSetNamespace = "kube-system"
)

// ImageInfo implements ImageInfoer interface for k8s DaemonSet that's used to
// deploy StorageOS node. DaemonSet name and namespace must be provided for the
// proper selection of DaemonSet that contains the required container.
type ImageInfo struct {
	client             kubernetes.Interface
	daemonSetName      string
	daemonSetNamespace string
	containerName      string
}

// NewImageInfo returns an initialized ImageInfo.
func NewImageInfo(client kubernetes.Interface) *ImageInfo {
	return &ImageInfo{
		client: client,
	}
}

// SetDaemonSet sets the k8s DaemonSet name and namespace of ImageInfo.
func (i *ImageInfo) SetDaemonSet(name, namespace string) *ImageInfo {
	i.daemonSetName = name
	i.daemonSetNamespace = namespace
	return i
}

// GetContainerImage returns the container image name of a given container in a
// DaemonSet Pod. SetDaemonSet() must be used before calling this to set
// DaemonSet attributes if not using the default StorageOS deployment.
func (i *ImageInfo) GetContainerImage(containerName string) (string, error) {
	if i.daemonSetName == "" {
		i.daemonSetNamespace = DefaultDaemonSetName
	}
	if i.daemonSetNamespace == "" {
		i.daemonSetNamespace = DefaultDaemonSetNamespace
	}

	daemonset, err := i.client.AppsV1().DaemonSets(i.daemonSetNamespace).Get(i.daemonSetName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	// Find the target container by name.
	for _, container := range daemonset.Spec.Template.Spec.Containers {
		if container.Name == containerName {
			return container.Image, nil
		}
	}

	return "", fmt.Errorf("failed to find container %q in daemonset %q", i.containerName, i.daemonSetName)
}
