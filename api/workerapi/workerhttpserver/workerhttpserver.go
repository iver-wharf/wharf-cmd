package workerhttpserver

import (
	"github.com/iver-wharf/wharf-cmd/pkg/worker"
	"github.com/iver-wharf/wharf-core/pkg/logger"
)

var log = logger.NewScoped("WORKER-HTTP-SERVER")

type workerServer struct {
	builder worker.Builder
}

func newWorkerServer(builder worker.Builder) *workerServer {
	return &workerServer{builder: builder}
}
