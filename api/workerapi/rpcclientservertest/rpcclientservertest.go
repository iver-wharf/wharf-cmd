package main

import (
	"time"

	v1 "github.com/iver-wharf/wharf-cmd/api/workerapi/v1"
	"github.com/iver-wharf/wharf-cmd/api/workerapi/workerrpcclient"
	"github.com/iver-wharf/wharf-cmd/api/workerapi/workerrpcserver"
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

func launchServer() *workerrpcserver.Server {
	bindAddress, bindPort := "0.0.0.0", "8081"
	server := workerrpcserver.NewServer(bindAddress, bindPort, &mockStore{})
	server.SetOnServeErrorHandler(func(err error) {
		log.Error().WithError(err).Message("Serve error occurred.")
		time.Sleep(1 * time.Second)
		if err := server.Serve(); err != nil {
			log.Error().WithError(err).Message("Auto-restart of server failed.")
		}
	})

	if err := server.Serve(); err != nil {
		log.Error().WithError(err).Message("Starting server failed.")
		return nil
	}
	return server
}

func launchClient() *workerrpcclient.Client {
	targetAddress, targetPort := "127.0.0.1", "8081"
	client := workerrpcclient.NewClient(targetAddress, targetPort)
	err := client.Open()
	if err != nil {
		log.Error().WithError(err).Message("Creating client failed.")
		return nil
	}
	return client
}

func testClientCalls(client *workerrpcclient.Client) {
	if err := client.HandleLogStream(func(line *v1.LogLine) {
		log.Info().Messagef("%v", line)
	}); err != nil {
		log.Error().WithError(err).Message("")
	}
	if err := client.HandleStatusEventStream(func(statusEvent *v1.StatusEvent) {
		log.Info().Messagef("%v", statusEvent)
	}); err != nil {
		log.Error().WithError(err).Message("")
	}
	if err := client.HandleArtifactEventStream(func(artifactEvent *v1.ArtifactEvent) {
		log.Info().Messagef("%v", artifactEvent)
	}); err != nil {
		log.Error().WithError(err).Message("")
	}
}
