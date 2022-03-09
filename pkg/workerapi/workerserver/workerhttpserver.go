package workerserver

import (
	"io"

	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml"
	"github.com/iver-wharf/wharf-cmd/pkg/workerapi/workerserver/model/response"
)

// StepLister is an interface that provides a way to list build steps.
type StepLister interface {
	ListAllSteps() []wharfyml.Step
}

// ArtifactLister is an interface that provides a way to list artifacts.
type ArtifactLister interface {
	ListArtifacts() []response.Artifact
}

// ArtifactDownloader is an interface that provides a way to download
// an artifact.
type ArtifactDownloader interface {
	DownloadArtifact(artifactID uint) (io.ReadCloser, error)
}

type workerHTTPServer struct {
	stepLister         StepLister
	artifactLister     ArtifactLister
	artifactDownloader ArtifactDownloader
}

func newWorkerHTTPServer(
	stepLister StepLister,
	artifactLister ArtifactLister,
	artifactDownloader ArtifactDownloader) *workerHTTPServer {
	return &workerHTTPServer{
		stepLister:         stepLister,
		artifactLister:     artifactLister,
		artifactDownloader: artifactDownloader,
	}
}
