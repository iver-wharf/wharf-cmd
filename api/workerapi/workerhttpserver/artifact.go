package workerhttpserver

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/iver-wharf/wharf-cmd/api/workerapi/workerhttpserver/model/response"
	"github.com/iver-wharf/wharf-core/pkg/ginutil"
	"github.com/iver-wharf/wharf-core/pkg/problem"
)

type artifactModule struct {
	*workerServer
}

var artifacts = []response.Artifact{
	{
		ArtifactID: 34,
		StepID:     1,
		Name:       "Artifact-34",
	},
	{
		ArtifactID: 35,
		StepID:     1,
		Name:       "Artifact-35",
	},
}

func (m *artifactModule) register(g *gin.RouterGroup) {
	artifact := g.Group("/artifact")
	{
		artifact.GET("", m.listArtifactsHandler)
		{
			artifactID := artifact.Group(":artifactId")
			{
				artifactID.GET("/download", m.downloadArtifact)
			}
		}
	}
}

func (m *artifactModule) listArtifactsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, artifacts)
}

func (m *artifactModule) downloadArtifact(c *gin.Context) {
	artifactID, ok := ginutil.ParseParamUint(c, "artifactId")
	if !ok {
		return
	}

	if artifactID != 34 && artifactID != 35 {
		ginutil.WriteProblemError(c, fmt.Errorf("artifact not found"), problem.Response{
			Type:   "/prob/workerapi/record-not-found",
			Title:  "Record not found.",
			Status: http.StatusNotFound,
			Detail: fmt.Sprintf("Artifact with ID %d was not found.", artifactID),
		})
		return
	}

	file, ok := mockFile(c)
	stat, err := file.Stat()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	defer removeMockFile(file)

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.DataFromReader(http.StatusOK, stat.Size(), "application/octet-stream", bytes.NewReader(data), nil)
}

func mockFile(c *gin.Context) (*os.File, bool) {
	file, err := os.CreateTemp("", "file_*")
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return nil, false
	}

	if _, err := file.WriteString("Hello, Blob!\r\n"); err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return nil, false
	}

	if err := file.Sync(); err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return nil, false
	}

	if _, err := file.Seek(0, 0); err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return nil, false
	}

	return file, true
}

func removeMockFile(file *os.File) bool {
	if err := os.Remove(file.Name()); err != nil {
		log.Error().WithError(err).Message("Failed removing temporary file.")
		return false
	}
	return true
}
