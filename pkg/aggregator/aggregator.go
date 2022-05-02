package aggregator

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/model/request"
	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/wharfapi"
	v1 "github.com/iver-wharf/wharf-cmd/api/workerapi/v1"
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

	buildID := r.worker.BuildID()
	var sentLogs uint

	writer, err := r.wharfapi.CreateBuildLogStream(ctx)
	if err != nil {
		return fmt.Errorf("open logs stream to wharf-api: %w", err)
	}
	defer func() {
		resp, err := writer.CloseAndRecv()
		if err != nil {
			log.Warn().
				WithError(err).
				Message("Unexpected error when closing log writer stream to wharf-api.")
			return
		}
		log.Debug().
			WithUint("sent", sentLogs).
			WithUint("inserted", resp.LogsInserted).
			Message("Inserted logs into wharf-api.")
	}()

	for {
		logLine, err := reader.Recv()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return err
			}
			break
		}
		log.Debug().
			WithUint64("step", logLine.StepID).
			Message(logLine.Message)
		writer.Send(request.Log{
			BuildID:      buildID,
			WorkerLogID:  uint(logLine.LogID),
			WorkerStepID: uint(logLine.StepID),
			Timestamp:    logLine.Timestamp.AsTime(),
			Message:      logLine.Message,
		})
		sentLogs++
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
			WithUint64("step", artifactEvent.StepID).
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

	buildID := r.worker.BuildID()

	for {
		statusEvent, err := stream.Recv()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return err
			}
			break
		}
		status, ok := convBuildStatus(statusEvent.Status)
		if !ok {
			continue
		}
		r.wharfapi.UpdateBuildStatus(buildID, request.LogOrStatusUpdate{
			Status: status,
		})
		log.Debug().
			WithUint64("step", statusEvent.StepID).
			WithStringer("status", statusEvent.Status).
			Message("Received status event.")
	}
	return nil
}

func convBuildStatus(status v1.Status) (request.BuildStatus, bool) {
	switch status {
	case v1.StatusPending, v1.StatusScheduling, v1.StatusInitializing:
		return request.BuildScheduling, true
	case v1.StatusRunning:
		return request.BuildRunning, true
	case v1.StatusSuccess:
		return request.BuildCompleted, true
	case v1.StatusCancelled, v1.StatusFailed:
		return request.BuildFailed, true
	default:
		return "", false
	}
}
