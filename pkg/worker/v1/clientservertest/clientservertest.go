package main

import (
	"log"
	"time"

	clientv1 "github.com/iver-wharf/wharf-cmd/pkg/worker/v1/client"
	serverv1 "github.com/iver-wharf/wharf-cmd/pkg/worker/v1/server"
)

func main() {
	server := serverv1.Server{}
	server.Start()
	time.Sleep(time.Millisecond * 500)

	client, err := clientv1.NewClient()
	if err != nil {
		log.Fatalf("creating client failed: %v", err)
		return
	}

	client.PrintLogs()
	client.PrintStatusEvents()
	client.PrintArtifactEvents()

	server.GracefulStop()
}
