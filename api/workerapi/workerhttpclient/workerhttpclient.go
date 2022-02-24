package workerhttpclient

import (
	"github.com/iver-wharf/wharf-cmd/api/workerapi/workerhttpserver/model/response"
	"github.com/iver-wharf/wharf-core/pkg/logger"
)

var log = logger.NewScoped("WORKER-HTTP-CLIENT")

type Client interface {
	GetBuildSteps() ([]response.Step, error)
}
