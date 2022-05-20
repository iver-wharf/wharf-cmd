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

// New creates a new client that can communicate with a Wharf worker server.
//
// Implements the Closer interface.
func New(baseURL string, opts Options) (Client, error) {
	rest, err := newRestClient(opts)
	if err != nil {
		return Client{}, err
	}
	return Client{
		rest:    rest,
		grpc:    newGRPCClient(baseURL, opts),
		baseURL: baseURL,
		buildID: opts.BuildID,
	}, nil
}

// Client is a HTTP (gRPC & REST) client that talks to wharf-cmd-worker.
// A new instance should be created via New to initiate it correctly.
type Client struct {
	rest    *restClient
	grpc    *grpcClient
	baseURL string
	buildID uint
}

// Options contains options that can be used in the creation
// of a new client.
type Options struct {
	// InsecureSkipVerify disables cert verification if set to true.
	//
	// Should NOT be true in a production environment.
	InsecureSkipVerify bool

	// BuildID is the ID of the build from wharf-api.
	BuildID uint
}

// StreamLogs returns a stream that will receive log lines from the worker.
func (c *Client) StreamLogs(ctx context.Context, req *LogsRequest, opts ...grpc.CallOption) (v1.Worker_StreamLogsClient, error) {
	if err := c.grpc.ensureOpen(); err != nil {
		return nil, err
	}
	return c.grpc.client.StreamLogs(ctx, req, opts...)
}

// StreamStatusEvents returns a stream that will receive status events from the
// worker.
func (c *Client) StreamStatusEvents(ctx context.Context, req *StatusEventsRequest, opts ...grpc.CallOption) (v1.Worker_StreamStatusEventsClient, error) {
	if err := c.grpc.ensureOpen(); err != nil {
		return nil, err
	}
	return c.grpc.client.StreamStatusEvents(ctx, req, opts...)
}

// StreamArtifactEvents returns a stream that will receive status events from the
// worker.
func (c *Client) StreamArtifactEvents(ctx context.Context, req *ArtifactEventsRequest, opts ...grpc.CallOption) (v1.Worker_StreamArtifactEventsClient, error) {
	if err := c.grpc.ensureOpen(); err != nil {
		return nil, err
	}
	return c.grpc.client.StreamArtifactEvents(ctx, req, opts...)
}

// DownloadArtifact will open a stream to download an artifact BLOB.
func (c *Client) DownloadArtifact(ctx context.Context, artifactID uint) (io.ReadCloser, error) {
	res, err := c.rest.get(ctx, fmt.Sprintf("%s/api/artifact/%d/download", c.baseURL, artifactID))
	if err := assertResponseOK(res, err); err != nil {
		return nil, err
	}
	return res.Body, nil
}

// BuildID returns the worker's build ID. The value zero
// means the worker does not have an assigned build ID.
func (c *Client) BuildID() uint {
	return c.buildID
}

// Ping pongs.
func (c *Client) Ping(ctx context.Context) error {
	res, err := c.rest.get(ctx, c.baseURL)
	return assertResponseOK(res, err)
}

// Close will terminate all active connections.
// Currently only gRPC streams are affected.
func (c *Client) Close() error {
	return c.grpc.close()
}
