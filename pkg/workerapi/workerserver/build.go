package workerserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iver-wharf/wharf-cmd/pkg/workerapi/workerserver/modelconv"
)

type buildModule struct {
	*workerHTTPServer
}

func (m *buildModule) register(g *gin.RouterGroup) {
	build := g.Group("/build")
	{
		build.GET("/step", m.listBuildStepsHandler)
	}
}

func (m *buildModule) listBuildStepsHandler(c *gin.Context) {
	steps := m.buildStepLister.ListBuildSteps()
	responseSteps := modelconv.StepsToResponseSteps(steps)
	c.JSON(http.StatusOK, responseSteps)
}
