package provisionerapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iver-wharf/wharf-cmd/pkg/config"
	"github.com/iver-wharf/wharf-cmd/pkg/provisioner"
	"github.com/iver-wharf/wharf-core/pkg/ginutil"
	"k8s.io/client-go/rest"
)

type workerModule struct {
	prov provisioner.Provisioner
}

func (m *workerModule) init(config config.ProvisionerConfig, cfg *rest.Config) error {
	p, err := provisioner.NewK8sProvisioner(config, cfg)
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

// listWorkersHandler godoc
// @id listWorkers
// @summary List provisioned wharf-cmd-workers
// @description Added in v0.8.0.
// @tags worker
// @success 200 {object} []provisioner.Worker
// @failure 500 {object} string "Failed"
// @router /worker [get]
func (m *workerModule) listWorkersHandler(c *gin.Context) {
	workers, err := m.prov.ListWorkers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, workers)
}

// createWorkerHandler godoc
// @id createWorker
// @summary Creates a new wharf-cmd-worker
// @description Added in v0.8.0.
// @tags worker
// @success 201 {object} provisioner.Worker
// @failure 500 {object} string "Failed"
// @router /worker [post]
func (m *workerModule) createWorkerHandler(c *gin.Context) {
	worker, err := m.prov.CreateWorker(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, worker)
}

// deleteWorkerHandler godoc
// @id deleteWorker
// @summary Deletes a wharf-cmd-worker
// @description Added in v0.8.0.
// @tags worker
// @param workerId path uint true "ID of worker to delete" minimum(0)
// @success 204 "OK"
// @failure 500 {object} string "Failed"
// @router /worker/{workerId} [delete]
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
