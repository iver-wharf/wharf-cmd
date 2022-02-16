package main

import (
	"log"

	"github.com/iver-wharf/wharf-cmd/api/workerapi/client"
	"github.com/iver-wharf/wharf-cmd/api/workerapi/server"
)

func main() {
	server := server.Server{}
	server.Start()

	client, err := client.NewClient()
	if err != nil {
		log.Fatalf("creating client failed: %v", err)
		return
	}

	client.PrintStreamedLogs()
	client.PrintLogs()
	client.PrintStatusEvents()
	client.PrintArtifactEvents()

	server.GracefulStop()
}
