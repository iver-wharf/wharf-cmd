// Package workerserver provides a way for a worker to set up servers for both
// gRPC and HTTP communication.
package workerserver

import (
	"net"
	"time"

	"github.com/iver-wharf/wharf-cmd/pkg/resultstore"
	"github.com/iver-wharf/wharf-core/pkg/logger"
	"github.com/soheilhy/cmux"
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
	time.Sleep(5 * time.Second)

	mux := cmux.New(s.listener)
	grpcListener := mux.MatchWithWriters(
		cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
	httpListener := mux.Match(cmux.Any())

	logIfErrored := func(protocol string, f func() error) {
		if err := f(); err != nil {
			log.Error().WithError(err).Messagef("Error during serving %s.", protocol)
		}
	}

	go logIfErrored("mux", func() error { return mux.Serve() })
	go logIfErrored("HTTP", func() error { return serveHTTP(s, s.rest, httpListener) })
	return serveGRPC(s.grpc, grpcListener)
}

// Close closes the server. No attempt to finish active requests is made.
//
// Any active gRPC connections will be notified by connection errors.
func (s *server) Close() error {
	if s.grpc != nil && s.grpc.grpc != nil {
		s.grpc.grpc.Stop()
	}
	return s.listener.Close()
}
