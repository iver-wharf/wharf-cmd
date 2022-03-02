package main

import (
	"time"

	"github.com/iver-wharf/wharf-cmd/pkg/workerapi/workerclient"
	"github.com/iver-wharf/wharf-cmd/pkg/workerapi/workerserver"
	"github.com/iver-wharf/wharf-core/pkg/logger"
	"github.com/iver-wharf/wharf-core/pkg/logger/consolepretty"
)

var log = logger.NewScoped("CLIENT-SERVER-TEST")

func main() {
	logger.AddOutput(logger.LevelDebug, consolepretty.New(consolepretty.DefaultConfig))

	server := launchServer()
	if server == nil {
		return
	}
	defer server.GracefulStop()

	client := launchClient()
	if client == nil {
		return
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Error().WithError(err).Message("Closing client returned error.")
		}
	}()

	testClientCalls(client)
}

func launchServer() workerserver.Server {
	server := workerserver.NewRPCServer("0.0.0.0:8081", &mockStore{})
	go func() {
		if err := server.Serve(); err != nil {
			log.Error().WithError(err).Message("Starting server failed.")
		}
	}()
	for !server.IsRunning() {
		time.Sleep(time.Millisecond)
	}
	return server
}

func launchClient() *workerclient.RPCClient {
	client := workerclient.NewRPCClient("127.0.0.1:8081")
	err := client.Open()
	if err != nil {
		log.Error().WithError(err).Message("Creating client failed.")
		return nil
	}
	return client
}

func testClientCalls(client *workerclient.RPCClient) {
	if err := client.HandleLogStream(func(line *workerclient.LogLine) {
		log.Info().Messagef("%v", line)
	}); err != nil {
		log.Error().WithError(err).Message("")
	}
	if err := client.HandleStatusEventStream(func(statusEvent *workerclient.StatusEvent) {
		log.Info().Messagef("%v", statusEvent)
	}); err != nil {
		log.Error().WithError(err).Message("")
	}
	if err := client.HandleArtifactEventStream(func(artifactEvent *workerclient.ArtifactEvent) {
		log.Info().Messagef("%v", artifactEvent)
	}); err != nil {
		log.Error().WithError(err).Message("")
	}
}
