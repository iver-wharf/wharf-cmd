package workerclient

import (
	"context"

	v1 "github.com/iver-wharf/wharf-cmd/api/workerapi/v1"
	"google.golang.org/grpc"
)

// RPC is used to communicate with the Worker gRPC server.
type RPC interface {
	Open() error
	Close() error
	StreamLogs(ctx context.Context, req *LogsRequest, opts ...grpc.CallOption) (v1.Worker_StreamLogsClient, error)
	StreamStatusEvents(ctx context.Context, req *StatusEventsRequest, opts ...grpc.CallOption) (v1.Worker_StreamStatusEventsClient, error)
	StreamArtifactEvents(ctx context.Context, req *ArtifactEventsRequest, opts ...grpc.CallOption) (v1.Worker_StreamArtifactEventsClient, error)
}

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
