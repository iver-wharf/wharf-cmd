package provisionerapi

import (
	"github.com/gin-gonic/gin"
	"github.com/iver-wharf/wharf-core/pkg/ginutil"
	"github.com/iver-wharf/wharf-core/pkg/logger"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"

	// Load in swagger docs
	"github.com/iver-wharf/wharf-cmd/pkg/provisioner"
	_ "github.com/iver-wharf/wharf-cmd/pkg/provisionerapi/docs"
)

var log = logger.NewScoped("PROVISIONER-API")

// Serve starts an HTTP server.
//
// @title Wharf provisioner API
// @description REST API for wharf-cmd to provision wharf-cmd-workers.
// @license.name MIT
// @license.url https://github.com/iver-wharf/wharf-cmd/blob/master/LICENSE
// @contact.name Iver wharf-cmd support
// @contact.url https://github.com/iver-wharf/wharf-cmd/issues
// @contact.email wharf@iver.se
// @basePath /api
// @query.collection.format multi
func Serve(prov provisioner.Provisioner) error {
	gin.DefaultWriter = ginutil.DefaultLoggerWriter
	gin.DefaultErrorWriter = ginutil.DefaultLoggerWriter

	r := gin.New()
	r.Use(
		ginutil.DefaultLoggerHandler,
		ginutil.RecoverProblem,
	)

	g := r.Group("/api")
	g.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	g.GET("", pingHandler)

	workerModule := workerModule{prov: prov}
	workerModule.register(g)

	const bindAddress = "0.0.0.0:5009"
	log.Info().WithString("address", bindAddress).Message("Starting server.")
	if err := r.Run(bindAddress); err != nil {
		log.Error().
			WithError(err).
			WithString("address", bindAddress).
			Message("Failed to start web server.")
		return err
	}

	return nil
}

// Ping is the response from a GET /api/ request.
type Ping struct {
	Message string `json:"message" example:"pong"`
}

// pingHandler godoc
// @id ping
// @summary Ping
// @description Pong.
// @description Added in v0.8.0.
// @tags meta
// @produce json
// @success 200 {object} Ping
// @router / [get]
func pingHandler(c *gin.Context) {
	c.JSON(200, Ping{Message: "pong"})
}
