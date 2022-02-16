package main

import (
	"log"
	"time"

	"github.com/iver-wharf/wharf-cmd/api/workerapi/client"
	"github.com/iver-wharf/wharf-cmd/api/workerapi/server"
)

func main() {
	server := server.Server{}
	server.Start()
	time.Sleep(time.Millisecond * 500)

	client, err := client.NewClient()
	if err != nil {
		log.Fatalf("creating client failed: %v", err)
		return
	}

	client.PrintLogs()
	client.PrintStatusEvents()
	client.PrintArtifactEvents()

	server.GracefulStop()
}
