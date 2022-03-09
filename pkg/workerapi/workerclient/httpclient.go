package workerclient

import (
	"io"

	"github.com/iver-wharf/wharf-cmd/pkg/workerapi/workerserver/model/response"
)

// HTTP is an interface with methods to communicate with a Wharf HTTP server.
type HTTP interface {
	ListBuildSteps() ([]response.Step, error)
	ListArtifacts() ([]response.Artifact, error)
	DownloadArtifact(artifactID uint) (io.ReadCloser, error)
}
