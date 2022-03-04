package workerserver

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iver-wharf/wharf-core/pkg/ginutil"
	"github.com/iver-wharf/wharf-core/pkg/problem"
)

type artifactModule struct {
	*workerHTTPServer
}

func (m *artifactModule) register(g *gin.RouterGroup) {
	artifact := g.Group("/artifact")
	{
		artifact.GET("", m.listArtifactsHandler)
		{
			artifactID := artifact.Group(":artifactId")
			{
				artifactID.GET("/download", m.downloadArtifactHandler)
			}
		}
	}
}

func (m *artifactModule) listArtifactsHandler(c *gin.Context) {
	artifacts := m.artifactLister.ListArtifacts()
	c.JSON(http.StatusOK, artifacts)
}

func (m *artifactModule) downloadArtifactHandler(c *gin.Context) {
	artifactID, ok := ginutil.ParseParamUint(c, "artifactId")
	if !ok {
		return
	}

	ioBody, err := m.artifactDownloader.DownloadArtifact(artifactID)
	if err != nil {
		ginutil.WriteProblemError(c, err, problem.Response{
			Type:   "/prob/workerapi/record-not-found",
			Title:  "Record not found.",
			Status: http.StatusNotFound,
			Detail: fmt.Sprintf("Artifact with ID %d was not found.", artifactID),
		})
		return
	}

	c.Header("Content-Type", "application/octet-stream")
	written, err := io.Copy(c.Writer, ioBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	log.Debug().WithInt64("written", written).Message("Successfully wrote from server to client.")
	c.Status(http.StatusOK)
}
