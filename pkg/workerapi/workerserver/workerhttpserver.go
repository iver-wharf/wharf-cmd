package workerserver

import (
	"github.com/iver-wharf/wharf-cmd/pkg/worker"
)

type workerHTTPServer struct {
	builder worker.Builder
}

func newWorkerHTTPServer(builder worker.Builder) *workerHTTPServer {
	return &workerHTTPServer{
		builder: builder,
	}
}
