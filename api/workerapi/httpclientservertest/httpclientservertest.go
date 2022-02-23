package main

import (
	"fmt"

	"github.com/iver-wharf/wharf-cmd/api/workerapi/workerhttpclient"
	"github.com/iver-wharf/wharf-cmd/api/workerapi/workerhttpserver"
	"github.com/iver-wharf/wharf-core/pkg/logger"
	"github.com/iver-wharf/wharf-core/pkg/logger/consolepretty"
)

func main() {
	logger.AddOutput(logger.LevelDebug, consolepretty.New(consolepretty.DefaultConfig))

	server := workerhttpserver.NewServer("0.0.0.0", "8080", &mockBuilder{})
	server.Serve()

	client := workerhttpclient.NewClient("localhost", "8080")

	steps, err := client.GetBuildSteps()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%v\n", steps)
}
