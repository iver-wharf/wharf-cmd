package provisionerapi

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/iver-wharf/wharf-cmd/pkg/config"
	"github.com/iver-wharf/wharf-cmd/pkg/provisioner"
	"github.com/iver-wharf/wharf-cmd/pkg/provisioner/provisionerserver/docs"
	"github.com/iver-wharf/wharf-core/v2/pkg/ginutil"
	"github.com/iver-wharf/wharf-core/v2/pkg/logger"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

var log = logger.NewScoped("PROV-SERVER")

// Serve starts an HTTP server.
//
// @title Wharf provisioner API
// @version v0.9.1
// @description REST API for wharf-cmd to provision wharf-cmd-workers.
// @license.name MIT
// @license.url https://github.com/iver-wharf/wharf-cmd/blob/master/LICENSE
// @contact.name Iver wharf-cmd support
// @contact.url https://github.com/iver-wharf/wharf-cmd/issues
// @contact.email wharf@iver.se
// @query.collection.format multi
func Serve(prov provisioner.Provisioner, config config.ProvisionerConfig) error {
	gin.DefaultWriter = ginutil.DefaultLoggerWriter
	gin.DefaultErrorWriter = ginutil.DefaultLoggerWriter

	r := gin.New()
	r.Use(
		ginutil.DefaultLoggerHandler,
		ginutil.RecoverProblem,
	)

	applyCORS(r, config.HTTP.CORS)

	r.GET("", pingHandler)
	api := r.Group("/api")
	api.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, func(c *ginSwagger.Config) {
		c.InstanceName = docs.SwaggerInfoprovisionerapi.InstanceName()
	}))

	workerModule{prov: prov}.register(api)

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

func applyCORS(r *gin.Engine, cfg config.CORSConfig) {
	if cfg.AllowAllOrigins {
		log.Info().Message("Allowing all origins in CORS.")
		corsConfig := cors.DefaultConfig()
		corsConfig.AllowAllOrigins = true
		r.Use(cors.New(corsConfig))
	} else if len(cfg.AllowOrigins) > 0 {
		log.Info().
			WithStringf("origin", "%v", cfg.AllowOrigins).
			Message("Allowing origins in CORS.")
		corsConfig := cors.DefaultConfig()
		corsConfig.AllowOrigins = cfg.AllowOrigins
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
// @produce json
// @success 200 {object} Ping
// @router / [get]
func pingHandler(c *gin.Context) {
	c.JSON(200, Ping{Message: "pong"})
}
