package provisionerapi

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/iver-wharf/wharf-core/pkg/ginutil"
	"github.com/iver-wharf/wharf-core/pkg/logger"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	"k8s.io/client-go/rest"

	// Load in swagger docs
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
func Serve(config Config, cfg *rest.Config) error {
	gin.DefaultWriter = ginutil.DefaultLoggerWriter
	gin.DefaultErrorWriter = ginutil.DefaultLoggerWriter

	r := gin.New()
	r.Use(
		ginutil.DefaultLoggerHandler,
		ginutil.RecoverProblem,
	)

	applyCORS(r, config.HTTP.CORS)

	g := r.Group("/api")
	g.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	g.GET("", pingHandler)

	workerModule := workerModule{}
	if err := workerModule.init(config.K8s, cfg); err != nil {
		log.Error().
			WithError(err).
			Message("Failed to initialize worker module.")
		return err
	}
	workerModule.register(g)

	log.Info().WithString("address", config.HTTP.BindAddress).Message("Starting server.")
	if err := r.Run(config.HTTP.BindAddress); err != nil {
		log.Error().
			WithError(err).
			WithString("address", config.HTTP.BindAddress).
			Message("Failed to start web server.")
		return err
	}

	return nil
}

func applyCORS(r *gin.Engine, c CORSConfig) {
	if c.AllowAllOrigins {
		log.Info().Message("Allowing all origins in CORS.")
		corsConfig := cors.DefaultConfig()
		corsConfig.AllowAllOrigins = true
		r.Use(cors.New(corsConfig))
	} else if len(c.AllowOrigins) > 0 {
		log.Info().
			WithStringf("origin", "%v", c.AllowOrigins).
			Message("Allowing origins in CORS.")
		corsConfig := cors.DefaultConfig()
		corsConfig.AllowOrigins = c.AllowOrigins
		corsConfig.AddAllowHeaders("Authorization")
		corsConfig.AllowCredentials = true
		r.Use(cors.New(corsConfig))
	}
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
// @success 200 {object} Ping
// @router / [get]
func pingHandler(c *gin.Context) {
	c.JSON(200, Ping{Message: "pong"})
}
