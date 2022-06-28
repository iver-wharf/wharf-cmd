package workerserver

import (
	"net"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/iver-wharf/wharf-cmd/pkg/worker/workerserver/docs"
	"github.com/iver-wharf/wharf-core/v2/pkg/ginutil"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

type restServer struct {
	artifactOpener ArtifactFileOpener
}

func newRestServer(artifactOpener ArtifactFileOpener) *restServer {
	return &restServer{
		artifactOpener: artifactOpener,
	}
}

// serveHTTP godoc
// @title Wharf worker API
// @version v0.9.1-rc.1
// @description REST API for wharf-cmd to access build results.
// @description Please refer to the gRPC API for more endpoints.
// @license.name MIT
// @license.url https://github.com/iver-wharf/wharf-cmd/blob/master/LICENSE
// @contact.name Iver wharf-cmd support
// @contact.url https://github.com/iver-wharf/wharf-cmd/issues
// @contact.email wharf@iver.se
// @query.collection.format multi
func serveHTTP(s *restServer, listener net.Listener) error {
	r := gin.New()
	applyGinHandlers(r)
	applyCORSConfig(r)

	r.GET("", pingHandler)

	api := r.Group("/api")
	api.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, func(c *ginSwagger.Config) {
		c.InstanceName = docs.SwaggerInfoworkerapi.InstanceName()
	}))
	artifactModule{s.artifactOpener}.register(api)
	return r.RunListener(listener)
}

func applyGinHandlers(r *gin.Engine) {
	gin.DefaultWriter = ginutil.DefaultLoggerWriter
	gin.DefaultErrorWriter = ginutil.DefaultLoggerWriter
	r.Use(
		ginutil.DefaultLoggerHandler,
		ginutil.RecoverProblem,
	)
}

func applyCORSConfig(r *gin.Engine) {
	log.Warn().Message("Allowing all origins in CORS.")
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	r.Use(cors.New(corsConfig))
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
