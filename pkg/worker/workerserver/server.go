// Package workerserver provides a way for a worker to set up servers for both
// gRPC and HTTP communication.
package workerserver

import (
	"errors"
	"net"
	"time"

	"github.com/iver-wharf/wharf-cmd/pkg/resultstore"
	"github.com/iver-wharf/wharf-core/v2/pkg/logger"
	"github.com/soheilhy/cmux"
	"gopkg.in/typ.v4"
)

var log = logger.NewScoped("WORKER-SERVER")

// Server contains both a gRPC server and an HTTP server.
//
// It provides an easy way to start both of these servers simultaneously,
// listening and responding to requests to either.
type Server interface {
	Serve(bindAddress string) error
	Close() error
}

type server struct {
	rest *restServer
	grpc *grpcWorkerServer

	listener net.Listener
}

// New creates a new server that can handle both HTTP and gRPC requests.
func New(store resultstore.Store, artifactOpener ArtifactFileOpener) Server {
	return &server{
		rest: newRestServer(artifactOpener),
		grpc: newGRPCServer(store),
	}
}

// Serve starts the gRPC and HTTP servers.
func (s *server) Serve(bindAddress string) error {
	var err error
	s.listener, err = net.Listen("tcp", bindAddress)
	if err != nil {
		return err
	}
	mux := cmux.New(s.listener)
	grpcListener := mux.MatchWithWriters(
		cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
	httpListener := mux.Match(cmux.Any())

	logIfErrored := func(protocol string, f func() error) {
		if err := f(); err != nil && !errors.Is(err, cmux.ErrListenerClosed) {
			log.Error().WithError(err).Messagef("Error during serving %s.", protocol)
		}
	}

	go logIfErrored("gRPC", func() error { return serveGRPC(s.grpc, grpcListener) })
	go logIfErrored("REST", func() error { return serveHTTP(s.rest, httpListener) })
	err = mux.Serve()
	return typ.Tern(errors.Is(err, net.ErrClosed), nil, err)
}

// Close closes the server.
//
// Tries to gracefully stop gRPC requests and connections.
// Abruptly stops active HTTP requests.
func (s *server) Close() error {
	if s.grpc != nil && s.grpc.grpc != nil {
		const timeout = 5 * time.Second
		log.Debug().WithDuration("timeout", timeout).
			Message("Attempting to shut down gracefully.")

		timer := time.AfterFunc(timeout, func() {
			log.Debug().Message("Timeout exceeded. Shutting down immediately.")
			s.grpc.grpc.Stop()
		})
		s.grpc.grpc.GracefulStop()
		timer.Stop()
	}
	return s.listener.Close()
}
