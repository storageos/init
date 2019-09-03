// Package info provides interfaces for information gatherers. The info
// gatherers are the information sources.
package info

// ImageInfoer is an interface that can be implemented by an information source
// to return container image information.
type ImageInfoer interface {
	GetContainerImage(containerName string) (string, error)
}
