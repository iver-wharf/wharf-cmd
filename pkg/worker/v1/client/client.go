package v1

import (
	"context"
	"log"
	"time"

	v1 "github.com/iver-wharf/wharf-cmd/pkg/worker/v1/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client is used to communicate with the Worker gRPC server.
type Client struct {
	v1.WorkerClient
	conn *grpc.ClientConn
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

func (c Client) printLogs() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	logs, err := c.Logs(ctx, &v1.LogsRequest{})
	if err != nil {
		log.Fatalf("%v.Logs(_) = _, %v: ", c, err)
	}
	log.Println(logs)
}

func (c Client) printStatusEvents() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	statusEvents, err := c.StatusEvents(ctx, &v1.StatusEventsRequest{})
	if err != nil {
		log.Fatalf("%v.StatusEvents(_) = _, %v: ", c, err)
	}
	log.Println(statusEvents)
}

func (c Client) printArtifactEvents() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	artifactEvents, err := c.ArtifactEvents(ctx, &v1.ArtifactEventsRequest{})
	if err != nil {
		log.Fatalf("%v.ArtifactEvents(_) = _, %v: ", c, err)
	}
	log.Println(artifactEvents)
}
