// Package workerserver has an HTTP server and a gRPC server implementation.
package workerserver

import "github.com/iver-wharf/wharf-core/pkg/logger"

var log = logger.NewScoped("WORKER-SERVER")

// Server is an interface that server implementations should
// adhere to, containing methods for:
//   Serving
//   Stopping
//   Setting an error handler.
type Server interface {
	// Serve starts the server.
	//
	// Also functions as a force-restart by calling ForceStop if the server is
	// already running, followed by attempting to launch it again.
	//
	// To stop the server you may use GracefulStop or ForceStop.
	Serve() error
	// ForceStop forcefully stops the server, not promising to take care of
	// existing connections.
	ForceStop() error
	// GracefulStop gracefully stops the server, rejecting new connections
	// and trying to let existing connections finish what they are doing.
	GracefulStop() error
	// IsRunning returns true if the server is currently running.
	IsRunning() bool
}
