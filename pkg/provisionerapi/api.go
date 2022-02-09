package provisionerapi

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iver-wharf/wharf-cmd/pkg/provisioner"
	"github.com/iver-wharf/wharf-core/pkg/ginutil"
	"github.com/iver-wharf/wharf-core/pkg/logger"
	"github.com/iver-wharf/wharf-core/pkg/logger/consolepretty"
	"k8s.io/client-go/rest"
)

var log = logger.NewScoped("PROVISIONER-API")
var kubeconfig *rest.Config

func Serve(cfg *rest.Config) error {
	kubeconfig = cfg
	logger.AddOutput(logger.LevelDebug, consolepretty.Default)

	gin.DefaultWriter = ginutil.DefaultLoggerWriter
	gin.DefaultErrorWriter = ginutil.DefaultLoggerWriter

	r := gin.New()
	r.Use(
		ginutil.DefaultLoggerHandler,
		ginutil.RecoverProblem,
	)

	g := r.Group("/api")
	g.GET("/", func(c *gin.Context) { c.JSON(200, gin.H{"message": "pong"}) })
	g.GET("/worker", listWorkersHandler)
	g.POST("/worker", createWorkerHandler)
	g.DELETE("/worker/:workerId", deleteWorkerHandler)

	bindAddress := "0.0.0.0:8080"
	if err := r.Run(bindAddress); err != nil {
		log.Error().
			WithError(err).
			WithString("address", bindAddress).
			Message("Failed to start web server.")
		return err
	}

	return nil
}

func listWorkersHandler(c *gin.Context) {
	p, err := provisioner.NewK8sProvisioner("default", kubeconfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	workers, err := p.ListWorkers(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, workers)
}

func createWorkerHandler(c *gin.Context) {
	p, err := provisioner.NewK8sProvisioner("default", kubeconfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	worker, err := p.CreateWorker(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, worker)
}

func deleteWorkerHandler(c *gin.Context) {
	p, err := provisioner.NewK8sProvisioner("default", kubeconfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	workerID, ok := ginutil.RequireParamString(c, "workerId")
	if !ok {
		return
	}

	if err := p.DeleteWorker(context.Background(), workerID); err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusNoContent, "OK")
}
