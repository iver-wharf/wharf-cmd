package main

import (
	"time"

	"github.com/iver-wharf/wharf-cmd/api/workerapi/workerclient"
	"github.com/iver-wharf/wharf-cmd/api/workerapi/workerserver"
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

	time.Sleep(1 * time.Second)
	testClientCalls(client)
}

func launchServer() *workerserver.Server {
	bindAddress, bindPort := "0.0.0.0", "8081"
	server := workerserver.NewServer(bindAddress, bindPort, &mockStore{})
	server.SetOnServeErrorHandler(func(err error) {
		log.Error().WithError(err).Message("OnServeError called. Restarting server.")
		// Try to auto-recover by restarting
		server.Serve()
	})
	err := server.Serve()
	if err != nil {
		log.Error().WithError(err).Message("Creating server failed.")
		return nil
	}
	return server
}

func launchClient() *workerclient.Client {
	targetAddress, targetPort := "127.0.0.1", "8081"
	client := workerclient.NewClient(targetAddress, targetPort)
	err := client.Open()
	if err != nil {
		log.Error().WithError(err).Message("Creating client failed.")
		return nil
	}
	return client
}

func testClientCalls(client *workerclient.Client) {
	log.Info().Message("-- PrintStreamedLogs")
	if err := client.PrintStreamedLogs(); err != nil {
		log.Error().WithError(err).Message("")
	}
	log.Info().Message("-- PrintLogs")
	if err := client.PrintLogs(); err != nil {
		log.Error().WithError(err).Message("")
	}
	log.Info().Message("-- PrintStatusEvents")
	if err := client.PrintStatusEvents(); err != nil {
		log.Error().WithError(err).Message("")
	}
	log.Info().Message("-- PrintArtifactEvents")
	if err := client.PrintArtifactEvents(); err != nil {
		log.Error().WithError(err).Message("")
	}
}