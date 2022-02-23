package workerhttpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iver-wharf/wharf-cmd/api/workerapi/workerhttpserver/modelconv"
)

type buildModule struct {
	*workerServer
}

func (m *buildModule) register(g *gin.RouterGroup) {
	build := g.Group("/build")
	{
		build.GET("/step", m.listStepsHandler)
	}
}

func (m *buildModule) listStepsHandler(c *gin.Context) {
	steps := m.builder.GetBuildSteps()
	responseSteps := modelconv.StepsToResponseSteps(steps)
	c.JSON(http.StatusOK, responseSteps)
}
