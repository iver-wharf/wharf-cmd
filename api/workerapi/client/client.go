package client

import (
	"context"
	"io"
	"log"
	"time"

	v1 "github.com/iver-wharf/wharf-cmd/api/workerapi/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client is used to communicate with the Worker gRPC server.
type Client struct {
	client v1.WorkerClient
	conn   *grpc.ClientConn
}

// NewClient creates a new gRPC Client that can communicate with the Worker
// gRPC server.
func NewClient() (*Client, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, err := grpc.Dial("127.0.0.1:8081", opts...)
	if err != nil {
		return nil, err
	}
	c := Client{v1.NewWorkerClient(conn), conn}
	return &c, nil
}

// Close tears down the connection to the server.
func (c Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

// PrintStreamedLogs prints logs received from the server.
//
// TEMPORARY: To test functionality.
func (c Client) PrintStreamedLogs() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := c.client.StreamLogs(ctx, &v1.StreamLogsRequest{})
	if err != nil {
		log.Fatalf("%v.Logs(_) = _, %v: ", c, err)
	}

	for {
		logLines, err := stream.Recv()
		if err == io.EOF {
			return
		} else if err != nil {
			log.Fatalf("%v.ListFeatures(_) = _, %v", c, err)
		}
		for _, v := range logLines.Logs {
			log.Printf("%v\n", v)
		}
	}
}

// PrintLogs prints logs received from the server.
//
// TEMPORARY: To test functionality.
func (c Client) PrintLogs() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := c.client.Logs(ctx, &v1.LogsRequest{})
	if err != nil {
		log.Fatalf("%v.Logs(_) = _, %v: ", c, err)
	}

	for {
		logLine, err := stream.Recv()
		if err == io.EOF {
			return
		} else if err != nil {
			log.Fatalf("%v.ListFeatures(_) = _, %v", c, err)
		}
		log.Printf("%v", logLine)
	}
}

// PrintStatusEvents prints status events received from the server.
//
// TEMPORARY: To test functionality.
func (c Client) PrintStatusEvents() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := c.client.StatusEvents(ctx, &v1.StatusEventsRequest{})
	if err != nil {
		log.Fatalf("%v.StatusEvents(_) = _, %v: ", c, err)
	}

	for {
		statusEvent, err := stream.Recv()
		if err == io.EOF {
			return
		} else if err != nil {
			log.Fatalf("%v.ListFeatures(_) = _, %v", c, err)
		}
		log.Printf("%v", statusEvent)
	}
}

// PrintArtifactEvents prints artifact events received from the server.
//
// TEMPORARY: To test functionality.
func (c Client) PrintArtifactEvents() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := c.client.ArtifactEvents(ctx, &v1.ArtifactEventsRequest{})
	if err != nil {
		log.Fatalf("%v.ArtifactEvents(_) = _, %v: ", c, err)
	}

	for {
		artifactEvent, err := stream.Recv()
		if err == io.EOF {
			return
		} else if err != nil {
			log.Fatalf("%v.ListFeatures(_) = _, %v", c, err)
		}
		log.Printf("%v", artifactEvent)
	}
}
