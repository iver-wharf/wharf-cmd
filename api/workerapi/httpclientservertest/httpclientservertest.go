package main

import (
	"fmt"

	"github.com/iver-wharf/wharf-cmd/api/workerapi/workerhttpclient"
	"github.com/iver-wharf/wharf-cmd/api/workerapi/workerhttpserver"
	"github.com/iver-wharf/wharf-core/pkg/logger"
	"github.com/iver-wharf/wharf-core/pkg/logger/consolepretty"
)

var log = logger.NewScoped("HTTP-CLIENT-SERVER-TEST")

func main() {
	logger.AddOutput(logger.LevelDebug, consolepretty.New(consolepretty.DefaultConfig))

	server := workerhttpserver.NewServer(&mockBuilder{})
	server.SetOnServeErrorHandler(func(err error) {
		log.Error().WithError(err).Message("Serve error occurred.")
	})
	server.Serve()

	client, err := workerhttpclient.NewClient()
	if err != nil {
		log.Error().WithError(err).Message("Creating client failed.")
	}

	steps, err := client.GetBuildSteps()
	if err != nil {
		log.Error().WithError(err).Message("Getting build steps failed.")
		return
	}

	fmt.Printf("%v\n", steps)
}
