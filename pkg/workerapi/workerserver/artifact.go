package workerserver

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iver-wharf/wharf-core/pkg/ginutil"
)

type artifactModule struct {
	artifactOpener ArtifactFileOpener
}

// ArtifactFileOpener is an interface that provides a way to read an artifact
// file's data using the artifact's ID.
type ArtifactFileOpener interface {
	// OpenArtifactFile gets an io.ReadCloser for the artifact with the given ID.
	OpenArtifactFile(artifactID uint) (io.ReadCloser, error)
}

func (m *artifactModule) register(g *gin.RouterGroup) {
	g.GET("/artifact/:artifactId/download", m.downloadArtifactHandler)
}

func (m *artifactModule) downloadArtifactHandler(c *gin.Context) {
	artifactID, ok := ginutil.ParseParamUint(c, "artifactId")
	if !ok {
		return
	}

	ioBody, err := m.artifactOpener.OpenArtifactFile(artifactID)
	if err != nil {
		ginutil.WriteDBNotFound(c, fmt.Sprintf("Unable to find artifact with ID %d.", artifactID))
		return
	}

	c.Header("Content-Type", "application/octet-stream")
	_, err = io.Copy(c.Writer, ioBody)
	if err != nil {
		ginutil.WriteAPIClientReadError(c, err,
			fmt.Sprintf("Unable to write artifact with ID %d to response.", artifactID))
		return
	}
	c.Status(http.StatusOK)
}
