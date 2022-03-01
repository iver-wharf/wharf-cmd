package workerhttpclient

import (
	"io"

	"github.com/iver-wharf/wharf-cmd/api/workerapi/workerhttpserver/model/response"
	"github.com/iver-wharf/wharf-core/pkg/logger"
)

var log = logger.NewScoped("WORKER-HTTP-CLIENT")

type Client interface {
	ListBuildSteps() ([]response.Step, error)
	ListArtifacts() ([]response.Artifact, error)
	DownloadArtifact(artifactID uint) (io.ReadCloser, error)
}
