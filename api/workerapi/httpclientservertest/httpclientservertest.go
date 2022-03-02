package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/iver-wharf/wharf-cmd/api/workerapi/workerhttpclient"
	"github.com/iver-wharf/wharf-cmd/api/workerapi/workerhttpserver"
	"github.com/iver-wharf/wharf-core/pkg/logger"
	"github.com/iver-wharf/wharf-core/pkg/logger/consolepretty"
	"github.com/iver-wharf/wharf-core/pkg/problem"
)

var log = logger.NewScoped("HTTP-CLIENT-SERVER-TEST")

var client workerhttpclient.Client

const (
	validArtifactID1  = 34
	validArtifactID2  = 35
	invalidArtifactID = 404
)

func main() {
	logger.AddOutput(logger.LevelDebug, consolepretty.New(consolepretty.DefaultConfig))

	server := workerhttpserver.NewServer("0.0.0.0", "8080", &mockBuilder{})
	server.SetOnServeErrorHandler(func(err error) {
		log.Error().WithError(err).Message("Serve error occurred.")
		time.Sleep(1 * time.Second)
		if err := server.Serve(); err != nil {
			log.Error().WithError(err).Message("Auto-restart of server failed.")
		}
	})
	if err := server.Serve(); err != nil {
		log.Error().WithError(err).Message("Starting server failed.")
		return
	}

	var err error
	client, err = workerhttpclient.NewClient("127.0.0.1", "8080")
	if err != nil {
		log.Error().WithError(err).Message("Creating client failed.")
		os.Exit(1)
	}

	testListBuildSteps()
	testListArtifacts()
	testDownloadArtifact(validArtifactID1)
	testDownloadArtifact(validArtifactID2)
	testDownloadArtifact(invalidArtifactID)

	log.Info().Message("Finished as expected, previous log should be a problem response.")
}

func testListBuildSteps() {
	steps, err := client.ListBuildSteps()
	if err != nil {
		log.Error().WithError(err).Message("Listing build steps failed.")
		os.Exit(2)
		return
	}
	fmt.Printf("%v\n", steps)
}

func testListArtifacts() {
	artifacts, err := client.ListArtifacts()
	if err != nil {
		log.Error().WithError(err).Message("Listing artifacts failed.")
		os.Exit(3)
		return
	}
	fmt.Printf("%v\n", artifacts)
}

func testDownloadArtifact(artifactID uint) {
	ioBody, err := client.DownloadArtifact(artifactID)
	if err != nil {
		log.Error().WithError(err).Message("Downloading artifact failed.")
		var probNotFound problem.Response
		if !errors.As(err, &probNotFound) {
			os.Exit(4)
		}
		return
	}

	outFile, err := os.CreateTemp("", "httpclientservertest_download_artifact_*")
	if err != nil {
		log.Error().WithError(err).Message("Creating temporary file to write to failed.")
		os.Exit(5)
		return
	}
	written, err := io.Copy(outFile, ioBody)
	if err != nil {
		log.Error().WithError(err).Message("Writing to temporary file failed.")
		os.Exit(6)
		return
	}

	log.Debug().
		WithInt64("written", written).
		WithString("path", outFile.Name()).
		Message("Successfully downloaded artifact to temporary file.")

	if err := os.Remove(outFile.Name()); err != nil {
		log.Error().WithError(err).Message("Deleting temporary file failed.")
		os.Exit(7)
		return
	}
}
