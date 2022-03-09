package workerclient

import (
	"context"
	"fmt"

	v1 "github.com/iver-wharf/wharf-cmd/api/workerapi/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// LogLine is an alias for workerapi/v1.StreamLogsResponse.
type LogLine = v1.StreamLogsResponse

// LogsRequest is an alias for workerapi/v1.StreamLogsRequest.
type LogsRequest = v1.StreamLogsRequest

// StatusEvent is an alias for workerapi/v1.StreamStatusEventsResponse.
type StatusEvent = v1.StreamStatusEventsResponse

// StatusEventsRequest is an alias for workerapi/v1.StreamStatusEventsRequest.
type StatusEventsRequest = v1.StreamStatusEventsRequest

// ArtifactEvent is an alias for workerapi/v1.StreamArtifactEventsResponse.
type ArtifactEvent = v1.StreamArtifactEventsResponse

// ArtifactEventsRequest is an alias for workerapi/v1.StreamArtifactEventsResponse.
type ArtifactEventsRequest = v1.StreamArtifactEventsRequest

// RPCClient is used to communicate with the Worker gRPC server.
type RPCClient struct {
	address string

	Client v1.WorkerClient
	conn   *grpc.ClientConn
}

// NewRPCClient creates a new gRPC Client that can communicate with the Worker
// gRPC server.
func NewRPCClient(address string) *RPCClient {
	return &RPCClient{
		address: address,
	}
}

// Open initializes the connection to the server.
func (c *RPCClient) Open() error {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.Dial(c.address, opts...)
	if err != nil {
		return fmt.Errorf("failed connecting to server: %v", err)
	}
	c.Client = v1.NewWorkerClient(conn)
	c.conn = conn
	return nil
}

// Close tears down the connection to the server.
func (c *RPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// StreamLogs returns a stream that will receive log lines from the worker.
func (c *RPCClient) StreamLogs(ctx context.Context, req *LogsRequest, opts ...grpc.CallOption) (v1.Worker_StreamLogsClient, error) {
	return c.Client.StreamLogs(ctx, req, opts...)
}

// StreamStatusEvents returns a stream that will receive status events from the
// worker.
func (c *RPCClient) StreamStatusEvents(ctx context.Context, req *StatusEventsRequest, opts ...grpc.CallOption) (v1.Worker_StreamStatusEventsClient, error) {
	return c.Client.StreamStatusEvents(ctx, req, opts...)
}

// StreamArtifactEvents returns a stream that will receive status events from the
// worker.
func (c *RPCClient) StreamArtifactEvents(ctx context.Context, req *ArtifactEventsRequest, opts ...grpc.CallOption) (v1.Worker_StreamArtifactEventsClient, error) {
	return c.Client.StreamArtifactEvents(ctx, req, opts...)
}
