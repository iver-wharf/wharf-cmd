package workerserver

import (
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml"
)

type workerHTTPServer struct {
	buildStepLister wharfyml.BuildStepLister
}

func newWorkerHTTPServer(buildStepLister wharfyml.BuildStepLister) *workerHTTPServer {
	return &workerHTTPServer{
		buildStepLister: buildStepLister,
	}
}
