package workerserver

import (
	"errors"
	"net"
	"sync"

	v1 "github.com/iver-wharf/wharf-cmd/api/workerapi/v1"
	"github.com/iver-wharf/wharf-cmd/pkg/resultstore"
	"google.golang.org/grpc"
)

type rpcServer struct {
	bindAddress string

	grpcServer   *grpc.Server
	workerServer *workerRPCServer
	serveMutex   *sync.Mutex
}

// NewRPCServer creates a new server that can be started by calling Serve.
func NewRPCServer(bindAddress string, store resultstore.Store) Server {
	return &rpcServer{
		bindAddress:  bindAddress,
		workerServer: newWorkerRPCServer(store),
		serveMutex:   &sync.Mutex{},
	}
}

func (s *rpcServer) Serve() error {
	s.ForceStop()
	s.serveMutex.Lock()
	defer s.serveMutex.Unlock()
	listener, err := net.Listen("tcp", s.bindAddress)
	if err != nil {
		return err
	}

	var opts []grpc.ServerOption
	s.grpcServer = grpc.NewServer(opts...)
	v1.RegisterWorkerServer(s.grpcServer, s.workerServer)

	log.Info().Messagef("Listening and serving gRPC on %s", s.bindAddress)
	err = s.grpcServer.Serve(listener)
	if errors.Is(err, grpc.ErrServerStopped) {
		log.Info().Message("gRPC server stopped.")
		return nil
	}
	return err
}

func (s *rpcServer) GracefulStop() error {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
		s.grpcServer = nil
	}
	return nil
}

func (s *rpcServer) ForceStop() error {
	if s.grpcServer != nil {
		s.grpcServer.Stop()
		s.grpcServer = nil
	}
	return nil
}

func (s *rpcServer) IsRunning() bool {
	return s.grpcServer != nil
}
