package workerhttpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type artifactModule struct {
	*workerServer
}

func (m *artifactModule) register(g *gin.RouterGroup) {
	artifact := g.Group("/artifact")
	{
		artifact.GET("/list", m.listArtifactsHandler)
		{
			artifactID := artifact.Group("/:artifactId")
			{
				artifactID.GET("/download")
			}
		}
	}
}

func (m *artifactModule) listArtifactsHandler(c *gin.Context) {
	c.Status(http.StatusInternalServerError)
}

func (m *artifactModule) downloadArtifact(c *gin.Context) {
	c.Status(http.StatusInternalServerError)
}
