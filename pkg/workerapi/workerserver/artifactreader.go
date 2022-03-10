package workerserver

import (
	"io"
)

// ArtifactReader is an interface that provides a way to read an artifact's
// data using the artifact's ID.
type ArtifactReader interface {
	// Get gets an io.ReadCloser for the artifact with the given ID.
	Get(artifactID uint) (io.ReadCloser, error)
}
