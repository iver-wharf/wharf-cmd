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
	buildStepLister    StepLister
	artifactLister     ArtifactLister
	artifactDownloader ArtifactDownloader
}

func newWorkerHTTPServer(
	buildStepLister StepLister,
	artifactLister ArtifactLister,
	artifactDownloader ArtifactDownloader) *workerHTTPServer {
	return &workerHTTPServer{
		buildStepLister:    buildStepLister,
		artifactLister:     artifactLister,
		artifactDownloader: artifactDownloader,
	}
}
