package provisionerapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iver-wharf/wharf-cmd/pkg/provisioner"
	"github.com/iver-wharf/wharf-core/pkg/ginutil"
	"gopkg.in/yaml.v3"
)

type workerModule struct {
	prov provisioner.Provisioner
}

func (m workerModule) register(g *gin.RouterGroup) {
	g.GET("/worker", m.listWorkersHandler)
	g.POST("/worker", m.createWorkerHandler)
	g.DELETE("/worker/:workerId", m.deleteWorkerHandler)
}

// listWorkersHandler godoc
// @id listWorkers
// @summary List provisioned wharf-cmd-workers
// @description Added in v0.8.0.
// @tags worker
// @produce json
// @success 200 {object} []provisioner.Worker
// @failure 500 {object} string "Failed"
// @router /api/worker [get]
func (m workerModule) listWorkersHandler(c *gin.Context) {
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
// @produce json
// @param BUILD_REF        query uint   true  "Build reference ID" example(123) minimum(0)
// @param ENVIRONMENT      query string false "Which Wharf environment to use, as defined in the `.wharf-ci.yml` file."
// @param GIT_BRANCH       query string false "Git branch" example(master)
// @param GIT_FULLURL      query string true  "Full Git clone'able URL" example(ssh://git@github.com/iver-wharf/wharf-cmd.git)
// @param GIT_TOKEN        query string false "Name of Git clone credentials secret"
// @param REPO_BRANCH      query string false "Git branch" example(master)
// @param REPO_GROUP       query string false "Repository group name" example(iver-wharf)
// @param REPO_NAME        query string false "Repository name" example(wharf-cmd)
// @param RUN_STAGES       query string false "Which stages to run" default(ALL) example(deploy)
// @param VARS             query string false "Input variable values, as a JSON or YAML formatted map of variable names (as defined in the project's `.wharf-ci.yml` file) as keys paired with their string, boolean, or numeric value."
// @param WHARF_INSTANCE   query string false "Wharf instance ID"
// @param WHARF_PROJECT_ID query uint   false "Wharf project ID" minimum(0) example(456)
// @success 201 {object} provisioner.Worker
// @failure 500 {object} string "Failed"
// @router /api/worker [post]
func (m workerModule) createWorkerHandler(c *gin.Context) {
	var params = struct {
		BuildRef       uint   `form:"BUILD_REF"`
		Environment    string `form:"ENVIRONMENT"`
		GitBranch      string `form:"GIT_BRANCH"`
		GitFullURL     string `form:"GIT_FULLURL"`
		GitToken       string `form:"GIT_TOKEN"`
		RepoBranch     string `form:"REPO_BRANCH"`
		RepoGroup      string `form:"REPO_GROUP"`
		RepoName       string `form:"REPO_NAME"`
		RunStages      string `form:"RUN_STAGES"`
		Vars           string `form:"VARS"`
		WharfInstance  string `form:"WHARF_INSTANCE"`
		WharfProjectID uint   `form:"WHARF_PROJECT_ID"`
	}{}
	if err := c.ShouldBindQuery(&params); err != nil {
		ginutil.WriteInvalidBindError(c, err, "One or more parameters failed to parse when reading query parameters.")
		return
	}
	if params.RunStages == "ALL" {
		params.RunStages = ""
	}
	if params.RepoBranch == "" {
		params.RepoBranch = params.GitBranch
	}
	args := provisioner.WorkerArgs{
		GitCloneBranch: params.GitBranch,
		GitCloneURL:    params.GitFullURL,

		Environment: params.Environment,
		Stage:       params.RunStages,

		WharfInstanceID: params.WharfInstance,
		ProjectID:       params.WharfProjectID,
		BuildID:         params.BuildRef,

		AdditionalVars: map[string]any{
			"BUILD_REF":   params.BuildRef,
			"GIT_BRANCH":  params.RepoBranch,
			"REPO_BRANCH": params.RepoBranch,
			"REPO_GROUP":  params.RepoGroup,
			"REPO_NAME":   params.RepoName,
		},
	}
	if params.Vars != "" {
		var inputVars map[string]any
		if err := yaml.Unmarshal([]byte(params.Vars), &inputVars); err != nil {
			ginutil.WriteInvalidParamError(c, err, "VARS",
				"Failed to parse input variables. They should be a JSON or YAML formatted map of key-values.")
			return
		}
		args.Inputs = inputVars
	}

	worker, err := m.prov.CreateWorker(c.Request.Context(), args)
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
// @produce json
// @param workerId path uint true "ID of worker to delete" minimum(0)
// @success 204 "OK"
// @failure 500 {object} string "Failed"
// @router /api/worker/{workerId} [delete]
func (m workerModule) deleteWorkerHandler(c *gin.Context) {
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
