// Package workerserver provides a way for a worker to set up servers for both
// gRPC and HTTP communication.
package workerserver

import (
	"net"

	"github.com/iver-wharf/wharf-cmd/pkg/resultstore"
	"github.com/iver-wharf/wharf-core/pkg/logger"
	"github.com/soheilhy/cmux"
)

var log = logger.NewScoped("WORKER-SERVER")

// Server contains both a gRPC server and an HTTP server.
//
// It provides an easy way to start both of these servers simultaneously,
// listening and responding to requests to either.
type Server struct {
	rest *restServer
	grpc *grpcWorkerServer

	listener net.Listener
}

// Serve starts the gRPC and HTTP servers.
func (s *Server) Serve(bindAddress string, store resultstore.Store, artifactReader ArtifactReader) error {
	s.rest = newRestServer(artifactReader)
	s.grpc = newGRPCServer(store)

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
		if err := f(); err != nil {
			log.Error().WithError(err).Messagef("Error during serving %s.", protocol)
		}
	}

	go logIfErrored("gRPC", func() error { return serveGRPC(s.grpc, grpcListener) })
	go logIfErrored("HTTP", func() error { return serveHTTP(s.rest, httpListener) })

	return mux.Serve()
}

// Close closes the listener.
//
// Any blocked Accept operations will be unblocked and return errors.
//
// Immediately closes all gRPC connections and listeners, and any active RPCs
// on both the client and server side will be notified by connection errors.
func (s *Server) Close() error {
	if s.grpc != nil && s.grpc.grpc != nil {
		s.grpc.grpc.Stop()
	}
	return s.listener.Close()
}
