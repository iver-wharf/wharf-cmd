package aggregator

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/model/request"
	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/wharfapi"
	"github.com/iver-wharf/wharf-cmd/internal/parallel"
	"github.com/iver-wharf/wharf-cmd/pkg/workerapi/workerclient"
)

// Aggregator pulls data from workers and sends them to Wharf API.
type Aggregator interface {
	Serve(ctx context.Context) error
}

func relayAll(ctx context.Context, wharfapi wharfapi.Client, worker workerclient.Client) error {
	r := relayer{
		wharfapi: wharfapi,
		worker:   worker,
	}
	var pg parallel.Group
	pg.AddFunc("logs", r.relayLogs)
	pg.AddFunc("artifact events", r.relayArtifactEvents)
	pg.AddFunc("status events", r.relayStatusEvents)
	return pg.RunCancelEarly(ctx)
}

type relayer struct {
	wharfapi wharfapi.Client
	worker   workerclient.Client
}

func (r relayer) relayLogs(ctx context.Context) error {
	reader, err := r.worker.StreamLogs(ctx, &workerclient.LogsRequest{})
	if err != nil {
		return fmt.Errorf("open logs stream from wharf-cmd-worker: %w", err)
	}
	defer reader.CloseSend()

	writer, err := r.wharfapi.CreateBuildLogStream(ctx)
	if err != nil {
		return fmt.Errorf("open logs stream to wharf-api: %w", err)
	}
	defer writer.CloseAndRecv()

	for {
		logLine, err := reader.Recv()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return err
			}
			break
		}
		log.Debug().Message(logLine.Message)
		writer.Send(request.Log{
			BuildID:      uint(logLine.BuildID),
			WorkerLogID:  uint(logLine.LogID),
			WorkerStepID: uint(logLine.StepID),
			Timestamp:    logLine.Timestamp.AsTime(),
			Message:      logLine.Message,
		})
	}
	return nil
}

func (r relayer) relayArtifactEvents(ctx context.Context) error {
	stream, err := r.worker.StreamArtifactEvents(ctx, &workerclient.ArtifactEventsRequest{})
	if err != nil {
		return fmt.Errorf("open artifact events stream from wharf-cmd-worker: %w", err)
	}
	defer stream.CloseSend()

	for {
		artifactEvent, err := stream.Recv()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return err
			}
			break
		}
		// No way to send to wharf DB through stream currently
		// so we're just logging it here.
		log.Debug().
			WithString("name", artifactEvent.Name).
			WithUint64("id", artifactEvent.ArtifactID).
			Message("Received artifact event.")
	}
	return nil
}

func (r relayer) relayStatusEvents(ctx context.Context) error {
	stream, err := r.worker.StreamStatusEvents(ctx, &workerclient.StatusEventsRequest{})
	if err != nil {
		return fmt.Errorf("open status events stream from wharf-cmd-worker: %w", err)
	}
	defer stream.CloseSend()

	// TODO: Update build status based on statuses
	for {
		statusEvent, err := stream.Recv()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return err
			}
			break
		}
		log.Debug().
			WithStringer("status", statusEvent.Status).
			WithUint64("step", statusEvent.StepID).
			Message("Received status event.")
	}
	return nil
}
