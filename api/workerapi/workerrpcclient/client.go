package workerrpcclient

import (
	"context"
	"fmt"
	"io"

	v1 "github.com/iver-wharf/wharf-cmd/api/workerapi/v1"
	"github.com/iver-wharf/wharf-core/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client is used to communicate with the Worker gRPC server.
type Client struct {
	targetAddress string
	targetPort    string

	Client v1.WorkerClient
	conn   *grpc.ClientConn
}

var log = logger.NewScoped("WORKER-RPC-CLIENT")

// NewClient creates a new gRPC Client that can communicate with the Worker
// gRPC server.
func NewClient(targetAddress, targetPort string) *Client {
	return &Client{
		targetAddress: targetAddress,
		targetPort:    targetPort,
	}
}

// Open initializes the connection to the server.
func (c *Client) Open() error {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", c.targetAddress, c.targetPort), opts...)
	if err != nil {
		return fmt.Errorf("failed connecting to server: %v", err)
	}
	c.Client = v1.NewWorkerClient(conn)
	c.conn = conn
	return nil
}

// Close tears down the connection to the server.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// PrintStreamedLogs prints logs received from the server.
//
// TEMPORARY: To test functionality.
func (c *Client) PrintStreamedLogs() error {
	stream, err := c.Client.StreamLogs(context.Background(), &v1.StreamLogsRequest{ChunkSize: 50})
	if err != nil {
		log.Error().WithError(err).Message("Error fetching stream for batched logs.")
		return err
	}

	for {
		logLines, err := stream.Recv()
		if err == io.EOF {
			return nil
		} else if err != nil {
			log.Error().WithError(err).Message("Error fetching from batched logs stream.")
			return err
		}
		for _, v := range logLines.Logs {
			log.Info().WithStringf("logLine", "%v", v).Message("")
		}
	}
}

// PrintLogs prints logs received from the server.
//
// TEMPORARY: To test functionality.
func (c *Client) PrintLogs() error {
	stream, err := c.Client.Log(context.Background(), &v1.LogRequest{}, grpc.FailFastCallOption{FailFast: true})
	if err != nil {
		log.Error().WithError(err).Message("Error fetching stream for logs.")
		return err
	}

	for {
		logLine, err := stream.Recv()
		if err == io.EOF {
			return nil
		} else if err != nil {
			log.Error().WithError(err).Message("Error fetching from logs stream.")
			return err
		}
		log.Info().WithStringf("logLine", "%v", logLine).Message("")
	}
}

// PrintStatusEvents prints status events received from the server.
//
// TEMPORARY: To test functionality.
func (c *Client) PrintStatusEvents() error {
	stream, err := c.Client.StatusEvent(context.Background(), &v1.StatusEventRequest{})
	if err != nil {
		log.Error().WithError(err).Message("Error fetching stream for status events.")
		return err
	}

	for {
		statusEvent, err := stream.Recv()
		if err == io.EOF {
			return nil
		} else if err != nil {
			log.Error().WithError(err).Message("Error fetching from status events stream.")
			return err
		}
		log.Info().WithStringf("statusEvent", "%v", statusEvent).Message("")
	}
}

// PrintArtifactEvents prints artifact events received from the server.
//
// TEMPORARY: To test functionality.
func (c *Client) PrintArtifactEvents() error {
	stream, err := c.Client.ArtifactEvent(context.Background(), &v1.ArtifactEventRequest{})
	if err != nil {
		log.Error().WithError(err).Message("Error fetching stream for artifact events.")
	}

	for {
		artifactEvent, err := stream.Recv()
		if err == io.EOF {
			return nil
		} else if err != nil {
			log.Error().WithError(err).Message("Error fetching from artifact events stream.")
			return err
		}
		log.Info().WithStringf("artifactEvent", "%v", artifactEvent).Message("")
	}
}
