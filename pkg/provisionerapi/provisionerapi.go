package provisionerapi

import (
	"github.com/gin-gonic/gin"
	"github.com/iver-wharf/wharf-core/pkg/ginutil"
	"github.com/iver-wharf/wharf-core/pkg/logger"
	"k8s.io/client-go/rest"
)

var log = logger.NewScoped("PROVISIONER-API")

// Serve starts an HTTP server.
func Serve(ns string, cfg *rest.Config) error {
	gin.DefaultWriter = ginutil.DefaultLoggerWriter
	gin.DefaultErrorWriter = ginutil.DefaultLoggerWriter

	r := gin.New()
	r.Use(
		ginutil.DefaultLoggerHandler,
		ginutil.RecoverProblem,
	)

	g := r.Group("/api")
	g.GET("/", func(c *gin.Context) { c.JSON(200, gin.H{"message": "pong"}) })

	workerModule := workerModule{}
	if err := workerModule.init(ns, cfg); err != nil {
		log.Error().
			WithError(err).
			Message("Failed to initialize worker module.")
		return err
	}
	workerModule.register(g)

	const bindAddress = "0.0.0.0:8080"
	if err := r.Run(bindAddress); err != nil {
		log.Error().
			WithError(err).
			WithString("address", bindAddress).
			Message("Failed to start web server.")
		return err
	}

	return nil
}
