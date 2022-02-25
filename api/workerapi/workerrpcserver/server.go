package workerrpcserver

import (
	"fmt"
	"net"
	"time"

	v1 "github.com/iver-wharf/wharf-cmd/api/workerapi/v1"
	"github.com/iver-wharf/wharf-cmd/pkg/resultstore"
	"github.com/iver-wharf/wharf-core/pkg/logger"
	"google.golang.org/grpc"
)

var log = logger.NewScoped("WORKER-RPC-SERVER")

// Server contains a gRPC server object, and lets us start the gRPC server
// asynchronously.
type Server struct {
	address             string
	port                string
	onServeErrorHandler func(error)

	grpcServer   *grpc.Server
	workerServer *workerServer
	isRunning    bool
}

// NewServer creates a new server that can be started by calling Start.
func NewServer(address, port string, store resultstore.Store) *Server {
	return &Server{
		address:      address,
		port:         port,
		workerServer: newWorkerServer(store),
	}
}

// SetOnServeErrorHandler sets the handler to call when an error occurs during
// serving.
func (s *Server) SetOnServeErrorHandler(onServeErrorHandler func(error)) {
	s.onServeErrorHandler = onServeErrorHandler
}

// Serve starts the gRPC server, and starts listening to it in a goroutine.
//
// Also functions as a force-restart by calling ForceStop if the server is
// already running, followed by attempting to launch it again.
//
// To stop the server you may use GracefulStop or ForceStop.
func (s *Server) Serve() error {
	s.ForceStop()

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", s.address, s.port))
	if err != nil {
		return err
	}

	var opts []grpc.ServerOption
	s.grpcServer = grpc.NewServer(opts...)
	v1.RegisterWorkerServer(s.grpcServer, s.workerServer)

	go func() {
		s.isRunning = true
		if err := s.grpcServer.Serve(listener); err != nil {
			log.Error().
				WithError(err).
				Message("Error during serving led to gRPC server closing unexpectedly.")
			if s.onServeErrorHandler != nil {
				s.onServeErrorHandler(err)
			}
		}
		s.isRunning = false
	}()

	for !s.isRunning {
		time.Sleep(1 * time.Millisecond)
	}

	return nil
}

// GracefulStop stops the gRPC server gracefully, blocking new connections
// and RPCs and blocks until pending ones have finished.
func (s *Server) GracefulStop() {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
		s.grpcServer = nil
	}
	s.isRunning = false
}

// ForceStop forcefully stops the gRPC server, closing all connections and pending RPCs.
func (s *Server) ForceStop() {
	if s.grpcServer != nil {
		s.grpcServer.Stop()
		s.grpcServer = nil
	}
	s.isRunning = false
}

// IsRunning returns true if the server is currently running and processing
// requests.
func (s *Server) IsRunning() bool {
	return s.grpcServer != nil && s.isRunning
}
