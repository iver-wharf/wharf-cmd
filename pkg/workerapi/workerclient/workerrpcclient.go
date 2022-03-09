package workerclient

import (
	"context"
	"fmt"

	v1 "github.com/iver-wharf/wharf-cmd/api/workerapi/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type workerRPC struct {
	address string

	client v1.WorkerClient
	conn   *grpc.ClientConn
}

// NewRPCClient creates a new gRPC Client that can communicate with the Worker
// gRPC server.
func NewRPCClient(address string) RPC {
	return &workerRPC{
		address: address,
	}
}

// Open initializes the connection to the server.
func (c *workerRPC) Open() error {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.Dial(c.address, opts...)
	if err != nil {
		return fmt.Errorf("failed connecting to server: %v", err)
	}
	c.client = v1.NewWorkerClient(conn)
	c.conn = conn
	return nil
}

// Close tears down the connection to the server.
func (c *workerRPC) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// StreamLogs returns a stream that will receive log lines from the worker.
func (c *workerRPC) StreamLogs(ctx context.Context, req *LogsRequest, opts ...grpc.CallOption) (v1.Worker_StreamLogsClient, error) {
	return c.client.StreamLogs(ctx, req, opts...)
}

// StreamStatusEvents returns a stream that will receive status events from the
// worker.
func (c *workerRPC) StreamStatusEvents(ctx context.Context, req *StatusEventsRequest, opts ...grpc.CallOption) (v1.Worker_StreamStatusEventsClient, error) {
	return c.client.StreamStatusEvents(ctx, req, opts...)
}

// StreamArtifactEvents returns a stream that will receive status events from the
// worker.
func (c *workerRPC) StreamArtifactEvents(ctx context.Context, req *ArtifactEventsRequest, opts ...grpc.CallOption) (v1.Worker_StreamArtifactEventsClient, error) {
	return c.client.StreamArtifactEvents(ctx, req, opts...)
}
