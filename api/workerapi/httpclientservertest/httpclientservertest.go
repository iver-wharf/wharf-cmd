package main

import (
	"fmt"
	"os"
	"time"

	"github.com/iver-wharf/wharf-cmd/api/workerapi/workerhttpclient"
	"github.com/iver-wharf/wharf-cmd/api/workerapi/workerhttpserver"
	"github.com/iver-wharf/wharf-core/pkg/logger"
	"github.com/iver-wharf/wharf-core/pkg/logger/consolepretty"
)

var log = logger.NewScoped("HTTP-CLIENT-SERVER-TEST")

func main() {
	logger.AddOutput(logger.LevelDebug, consolepretty.New(consolepretty.DefaultConfig))

	server := workerhttpserver.NewServer("0.0.0.0", "8080", &mockBuilder{})
	server.SetOnServeErrorHandler(func(err error) {
		log.Error().WithError(err).Message("Serve error occurred.")
		time.Sleep(1 * time.Second)
		server.Serve()
	})
	server.Serve()

	client, err := workerhttpclient.NewClient("127.0.0.1", "8080")
	if err != nil {
		log.Error().WithError(err).Message("Creating client failed.")
		os.Exit(1)
	}
	steps, err := client.GetBuildSteps()
	if err != nil {
		log.Error().WithError(err).Message("Getting build steps failed.")
		os.Exit(2)
		return
	}
	fmt.Printf("%v\n", steps)
}
