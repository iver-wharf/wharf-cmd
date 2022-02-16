package server

import (
	"log"
	"net"

	v1 "github.com/iver-wharf/wharf-cmd/api/workerapi/v1/rpc"
	"google.golang.org/grpc"
)

// Server contains a gRPC server object, and lets us start the gRPC server
// asynchronously.
type Server struct {
	grpcServer *grpc.Server
}

// Start starts the gRPC server using a goroutine.
//
// To stop the server you may use Server.GracefulStop or Server.Stop.
func (s *Server) Start() {
	go func() {
		listener, err := net.Listen("tcp", "0.0.0.0:8081")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		var opts []grpc.ServerOption
		grpcServer := grpc.NewServer(opts...)
		v1.RegisterWorkerServer(grpcServer, newServer())
		s.grpcServer = grpcServer
		grpcServer.Serve(listener)
	}()
}

// GracefulStop stops the gRPC server gracefully, blocking new connections
// and RPCs and blocks until pending ones have finished.
func (s *Server) GracefulStop() {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
}

// Stop stops the gRPC server, closing all connections and pending RPCs.
func (s *Server) Stop() {
	if s.grpcServer != nil {
		s.grpcServer.Stop()
	}
}
