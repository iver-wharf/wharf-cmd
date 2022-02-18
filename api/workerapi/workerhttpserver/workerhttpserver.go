package main

import (
	"github.com/gin-gonic/gin"
	"github.com/iver-wharf/wharf-core/pkg/ginutil"
	"github.com/iver-wharf/wharf-core/pkg/logger"
)

var log = logger.NewScoped("WORKER-HTTP-SERVER")

func main() {
	gin.DefaultWriter = ginutil.DefaultLoggerWriter
	gin.DefaultErrorWriter = ginutil.DefaultLoggerWriter

	r := gin.New()
	r.Use(
		ginutil.DefaultLoggerHandler,
		ginutil.RecoverProblem,
	)

	g := r.Group("/api")
	g.GET("/", func(c *gin.Context) { c.JSON(200, gin.H{"message": "pong"}) })

	{
		buildModule := buildModule{}
		artifactModule := artifactModule{}
		buildModule.register(g)
		artifactModule.register(g)
	}

	const bindAddress = "0.0.0.0:8082"
	if err := r.Run(bindAddress); err != nil {
		log.Error().
			WithError(err).
			WithString("address", bindAddress).
			Message("Failed to start web server.")
		return
	}

	return
}
