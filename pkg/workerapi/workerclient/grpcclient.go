package workerclient

import (
	"fmt"
	"net/url"
	"strings"

	v1 "github.com/iver-wharf/wharf-cmd/api/workerapi/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type grpcClient struct {
	address string
	opts    Options

	client v1.WorkerClient
	conn   *grpc.ClientConn
}

// NewRPCClient creates a new gRPC Client that can communicate with the Worker
// gRPC server.
func newGRPCClient(address string, opts Options) *grpcClient {
	return &grpcClient{
		address: address,
		opts:    opts,
	}
}

func (c *grpcClient) ensureOpen() error {
	if c.conn != nil {
		return nil
	}
	var opts []grpc.DialOption
	if c.opts.InsecureSkipVerify {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	conn, err := grpc.Dial(trimProtocol(c.address), opts...)
	if err != nil {
		return fmt.Errorf("failed connecting to server: %v", err)
	}
	c.client = v1.NewWorkerClient(conn)
	c.conn = conn
	return nil
}

func (c *grpcClient) close() error {
	if c.conn != nil {
		err := c.conn.Close()
		if err == nil {
			c.conn = nil
		}
		return err
	}
	return nil
}

func trimProtocol(addr string) string {
	u, err := url.Parse(addr)
	if err != nil {
		return addr
	}
	return strings.TrimPrefix(addr, u.Scheme+"://")
}
