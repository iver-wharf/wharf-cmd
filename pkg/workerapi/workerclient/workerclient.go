// Package workerclient provides a way to communicate with a Wharf worker
// server.
package workerclient

import (
	"context"
	"io"

	v1 "github.com/iver-wharf/wharf-cmd/api/workerapi/v1"
	"github.com/iver-wharf/wharf-core/pkg/logger"
	"google.golang.org/grpc"
)

var log = logger.NewScoped("WORKER-CLIENT")

// Client is an interface with methods to communicate with a Wharf worker server.
type Client interface {
	StreamLogs(ctx context.Context, req *LogsRequest, opts ...grpc.CallOption) (v1.Worker_StreamLogsClient, error)
	StreamStatusEvents(ctx context.Context, req *StatusEventsRequest, opts ...grpc.CallOption) (v1.Worker_StreamStatusEventsClient, error)
	StreamArtifactEvents(ctx context.Context, req *ArtifactEventsRequest, opts ...grpc.CallOption) (v1.Worker_StreamArtifactEventsClient, error)
	DownloadArtifact(artifactID uint) (io.ReadCloser, error)

	Close() error
}

// ClientOptions contains options that can be used in the creation
// of a new client.
type ClientOptions struct {
	// InsecureSkipVerify disables cert verification if set to true.
	//
	// Should NOT be true in a production environment.
	InsecureSkipVerify bool
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
