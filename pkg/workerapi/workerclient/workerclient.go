// Package workerclient provides a way to communicate with a Wharf worker
// server.
package workerclient

import (
	"context"
	"fmt"
	"io"

	v1 "github.com/iver-wharf/wharf-cmd/api/workerapi/v1"
	"github.com/iver-wharf/wharf-core/pkg/logger"
	"google.golang.org/grpc"
)

var log = logger.NewScoped("WORKER-CLIENT")

// LogLine is an alias for workerapi/v1.StreamLogsResponse.
type LogLine = v1.StreamLogsResponse

// LogsRequest is an alias for workerapi/v1.LogsRequest.
type LogsRequest = v1.StreamLogsRequest

// StatusEvent is an alias for workerapi/v1.StreamStatusEventsResponse.
type StatusEvent = v1.StreamStatusEventsResponse

// StatusEventsRequest is an alias for workerapi/v1.StreamStatusEventsRequest.
type StatusEventsRequest = v1.StreamStatusEventsRequest

// ArtifactEvent is an alias for workerapi/v1.StreamArtifactEventsResponse.
type ArtifactEvent = v1.StreamArtifactEventsResponse

// ArtifactEventsRequest is an alias for workerapi/v1.StreamArtifactEventsResponse.
type ArtifactEventsRequest = v1.StreamArtifactEventsRequest

// Client is an interface with methods to communicate with a Wharf worker server.
type Client interface {
	StreamLogs(ctx context.Context, req *LogsRequest, opts ...grpc.CallOption) (v1.Worker_StreamLogsClient, error)
	StreamStatusEvents(ctx context.Context, req *StatusEventsRequest, opts ...grpc.CallOption) (v1.Worker_StreamStatusEventsClient, error)
	StreamArtifactEvents(ctx context.Context, req *ArtifactEventsRequest, opts ...grpc.CallOption) (v1.Worker_StreamArtifactEventsClient, error)
	DownloadArtifact(ctx context.Context, artifactID uint) (io.ReadCloser, error)
	Ping(ctx context.Context) error

	Close() error
}

// New creates a new client that can communicate with a Wharf worker server.
//
// Implements the Closer interface.
func New(baseURL string, opts Options) (Client, error) {
	rest, err := newRestClient(opts)
	if err != nil {
		return nil, err
	}
	return &client{
		rest:    rest,
		grpc:    newGRPCClient(baseURL, opts),
		baseURL: baseURL,
	}, nil
}

type client struct {
	rest    *restClient
	grpc    *grpcClient
	baseURL string
}

// Options contains options that can be used in the creation
// of a new client.
type Options struct {
	// InsecureSkipVerify disables cert verification if set to true.
	//
	// Should NOT be true in a production environment.
	InsecureSkipVerify bool
}

// StreamLogs returns a stream that will receive log lines from the worker.
func (c *client) StreamLogs(ctx context.Context, req *LogsRequest, opts ...grpc.CallOption) (v1.Worker_StreamLogsClient, error) {
	if err := c.grpc.ensureOpen(); err != nil {
		return nil, err
	}
	return c.grpc.client.StreamLogs(ctx, req, opts...)
}

// StreamStatusEvents returns a stream that will receive status events from the
// worker.
func (c *client) StreamStatusEvents(ctx context.Context, req *StatusEventsRequest, opts ...grpc.CallOption) (v1.Worker_StreamStatusEventsClient, error) {
	if err := c.grpc.ensureOpen(); err != nil {
		return nil, err
	}
	return c.grpc.client.StreamStatusEvents(ctx, req, opts...)
}

// StreamArtifactEvents returns a stream that will receive status events from the
// worker.
func (c *client) StreamArtifactEvents(ctx context.Context, req *ArtifactEventsRequest, opts ...grpc.CallOption) (v1.Worker_StreamArtifactEventsClient, error) {
	if err := c.grpc.ensureOpen(); err != nil {
		return nil, err
	}
	return c.grpc.client.StreamArtifactEvents(ctx, req, opts...)
}

func (c *client) DownloadArtifact(ctx context.Context, artifactID uint) (io.ReadCloser, error) {
	res, err := c.rest.get(ctx, fmt.Sprintf("%s/api/artifact/%d/download", c.baseURL, artifactID))
	if err := assertResponseOK(res, err); err != nil {
		return nil, err
	}
	return res.Body, nil
}

func (c *client) Ping(ctx context.Context) error {
	res, err := c.rest.get(ctx, fmt.Sprintf("%s/api", c.baseURL))
	return assertResponseOK(res, err)
}

func (c *client) Close() error {
	return c.grpc.close()
}
