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
	g.GET("/build/step", m.listBuildStepsHandler)
}

func (m *buildModule) listBuildStepsHandler(c *gin.Context) {
	steps := m.stepLister.ListAllSteps()
	responseSteps := modelconv.StepsToResponseSteps(steps)
	c.JSON(http.StatusOK, responseSteps)
}
