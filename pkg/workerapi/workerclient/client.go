package workerclient

import (
	"context"
	"fmt"
	"io"

	v1 "github.com/iver-wharf/wharf-cmd/api/workerapi/v1"
	"google.golang.org/grpc"
)

// New creates a new client that can communicate with a Wharf worker server.
//
// Implements the Closer interface.
func New(baseURL string, opts ClientOptions) (Client, error) {
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

func (c *client) DownloadArtifact(artifactID uint) (io.ReadCloser, error) {
	res, err := c.rest.get(fmt.Sprintf("%s/api/artifact/%d/download", c.baseURL, artifactID))
	if err := assertResponseOK(res, err); err != nil {
		return nil, err
	}
	return res.Body, nil
}

func (c *client) Close() error {
	return c.grpc.close()
}
