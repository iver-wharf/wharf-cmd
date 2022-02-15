package provisionerapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iver-wharf/wharf-cmd/pkg/provisioner"
	"github.com/iver-wharf/wharf-core/pkg/ginutil"
	"k8s.io/client-go/rest"
)

type workerModule struct {
	prov provisioner.Provisioner
}

func (m *workerModule) init(cfg *rest.Config) error {
	p, err := provisioner.NewK8sProvisioner("default", cfg)
	if err != nil {
		return err
	}
	m.prov = p
	return nil
}

func (m *workerModule) register(g *gin.RouterGroup) {
	g.GET("/worker", m.listWorkersHandler)
	g.POST("/worker", m.createWorkerHandler)
	g.DELETE("/worker/:workerId", m.deleteWorkerHandler)
}

func (m *workerModule) listWorkersHandler(c *gin.Context) {
	workers, err := m.prov.ListWorkers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, workers)
}

func (m *workerModule) createWorkerHandler(c *gin.Context) {
	worker, err := m.prov.CreateWorker(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, worker)
}

func (m *workerModule) deleteWorkerHandler(c *gin.Context) {
	workerID, ok := ginutil.RequireParamString(c, "workerId")
	if !ok {
		return
	}

	if err := m.prov.DeleteWorker(c.Request.Context(), workerID); err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusNoContent, "OK")
}
