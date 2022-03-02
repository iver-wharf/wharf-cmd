package workerclient

import (
	"context"
	"fmt"
	"io"

	v1 "github.com/iver-wharf/wharf-cmd/api/workerapi/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// RPCClient is used to communicate with the Worker gRPC server.
type RPCClient struct {
	targetAddress string
	targetPort    string

	Client v1.WorkerClient
	conn   *grpc.ClientConn
}

// NewRPCClient creates a new gRPC Client that can communicate with the Worker
// gRPC server.
func NewRPCClient(targetAddress, targetPort string) *RPCClient {
	return &RPCClient{
		targetAddress: targetAddress,
		targetPort:    targetPort,
	}
}

// Open initializes the connection to the server.
func (c *RPCClient) Open() error {
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
func (c *RPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// StreamLogs returns a channel that will receive log lines from the worker.
func (c *RPCClient) StreamLogs() (<-chan *v1.LogLine, <-chan error) {
	ch := make(chan *v1.LogLine)
	errCh := make(chan error)
	stream, err := c.Client.StreamLogs(context.Background(), &v1.StreamLogLineRequest{})
	if err != nil {
		log.Error().WithError(err).Message("Error fetching stream for logs.")
		errCh <- err
		close(errCh)
		close(ch)
	} else {
		go func() {
			for {
				logLine, err := stream.Recv()
				if err != nil {
					close(ch)
					if err == io.EOF {
						errCh <- nil
					} else {
						log.Error().WithError(err).Message("Error fetching from logs stream.")
						errCh <- err
					}
					break
				}
				ch <- logLine
			}
			close(errCh)
		}()
	}
	return ch, errCh
}

// HandleLogStream is the functional equivalent of calling StreamLogs and
// passing each log to the callback.
func (c *RPCClient) HandleLogStream(onLogLine func(*v1.LogLine)) error {
	ch, errCh := c.StreamLogs()
	for line, ok := <-ch; ok; line, ok = <-ch {
		onLogLine(line)
	}
	for err := range errCh {
		return err
	}
	return nil
}

// StreamStatusEvents returns a channel that will receive status events from
// the worker.
func (c *RPCClient) StreamStatusEvents() (<-chan *v1.StatusEvent, <-chan error) {
	ch := make(chan *v1.StatusEvent)
	errCh := make(chan error)
	stream, err := c.Client.StreamStatusEvents(context.Background(), &v1.StreamStatusEventRequest{})
	if err != nil {
		log.Error().WithError(err).Message("Error fetching stream for batched logs.")
		errCh <- err
		close(errCh)
		close(ch)
	} else {
		go func() {
			for {
				statusEvent, err := stream.Recv()
				if err != nil {
					close(ch)
					if err == io.EOF {
						errCh <- nil
					} else {
						log.Error().WithError(err).Message("Error fetching from batched logs stream.")
						errCh <- err
					}
					break
				}
				ch <- statusEvent
			}
			close(errCh)
		}()
	}
	return ch, errCh
}

// HandleStatusEventStream is the functional equivalent of calling
// StreamStatusEvents and passing each event to the callback.
func (c *RPCClient) HandleStatusEventStream(onStatusEvent func(*v1.StatusEvent)) error {
	ch, errCh := c.StreamStatusEvents()
	for statusEvent, ok := <-ch; ok; statusEvent, ok = <-ch {
		onStatusEvent(statusEvent)
	}
	for err := range errCh {
		return err
	}
	return nil
}

// StreamArtifactEvents returns a channel that will receive artifact events
// from the worker.
func (c *RPCClient) StreamArtifactEvents() (<-chan *v1.ArtifactEvent, <-chan error) {
	ch := make(chan *v1.ArtifactEvent)
	errCh := make(chan error)
	stream, err := c.Client.StreamArtifactEvents(context.Background(), &v1.StreamArtifactEventRequest{})
	if err != nil {
		log.Error().WithError(err).Message("Error fetching stream for batched logs.")
		errCh <- err
		close(errCh)
		close(ch)
	} else {
		go func() {
			for {
				artifactEvent, err := stream.Recv()
				if err != nil {
					close(ch)
					if err == io.EOF {
						errCh <- nil
					} else {
						log.Error().WithError(err).Message("Error fetching from batched logs stream.")
						errCh <- err
					}
					break
				}
				ch <- artifactEvent
			}
			close(errCh)
		}()
	}
	return ch, errCh
}

// HandleArtifactEventStream is the functional equivalent of calling
// StreamArtifactEvents and passing each event to the callback.
func (c *RPCClient) HandleArtifactEventStream(onArtifactEvent func(*v1.ArtifactEvent)) error {
	ch, errCh := c.StreamArtifactEvents()
	for artifactEvent, ok := <-ch; ok; artifactEvent, ok = <-ch {
		onArtifactEvent(artifactEvent)
	}
	for err := range errCh {
		return err
	}
	return nil
}
