package workerserver

import (
	"fmt"
	"net"
	"time"

	v1 "github.com/iver-wharf/wharf-cmd/api/workerapi/v1"
	"github.com/iver-wharf/wharf-cmd/pkg/resultstore"
	"google.golang.org/grpc"
)

type rpcServer struct {
	address             string
	port                string
	onServeErrorHandler func(error)

	grpcServer   *grpc.Server
	workerServer *workerRPCServer
	isRunning    bool
}

// NewRPCServer creates a new server that can be started by calling Start.
func NewRPCServer(address, port string, store resultstore.Store) Server {
	return &rpcServer{
		address:      address,
		port:         port,
		workerServer: newWorkerRPCServer(store),
	}
}

func (s *rpcServer) SetOnServeErrorHandler(onServeErrorHandler func(error)) {
	s.onServeErrorHandler = onServeErrorHandler
}

func (s *rpcServer) Serve() error {
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

func (s *rpcServer) GracefulStop() error {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
		s.grpcServer = nil
	}
	s.isRunning = false
	return nil
}

func (s *rpcServer) ForceStop() error {
	if s.grpcServer != nil {
		s.grpcServer.Stop()
		s.grpcServer = nil
	}
	s.isRunning = false
	return nil
}

func (s *rpcServer) IsRunning() bool {
	return s.grpcServer != nil && s.isRunning
}
