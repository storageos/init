// Package k8s implements k8s related helpers to interact with a k8s cluster.
package k8s

import (
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

// NewK8SClient attempts to get k8s cluster configuration and return a new
// kubernetes client.
func NewK8SClient() (kubernetes.Interface, error) {
	cfg, err := restclient.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(cfg)
}
