package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type buildModule struct{}

func (m *buildModule) register(g *gin.RouterGroup) {
	build := g.Group("/build")
	{
		build.GET("/step", m.listStepsHandler)
	}
}

func (m *buildModule) listStepsHandler(c *gin.Context) {
	c.Status(http.StatusInternalServerError)
}
