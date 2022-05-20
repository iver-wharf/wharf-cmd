package aggregator

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/model/request"
	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/wharfapi"
	v1 "github.com/iver-wharf/wharf-cmd/api/workerapi/v1"
	"github.com/iver-wharf/wharf-cmd/pkg/worker/workerclient"
)

type statusEventsPiper struct {
	wharfapi   wharfapi.Client
	worker     portForwardedWorker
	in         v1.Worker_StreamStatusEventsClient
	prevStatus request.BuildStatus
}

func newStatusEventsPiper(ctx context.Context, wharfapi wharfapi.Client, worker portForwardedWorker) (statusEventsPiper, error) {
	in, err := worker.StreamStatusEvents(ctx, &workerclient.StatusEventsRequest{})
	if err != nil {
		return statusEventsPiper{}, fmt.Errorf("open status events stream from wharf-cmd-worker: %w", err)
	}
	return statusEventsPiper{
		wharfapi:   wharfapi,
		worker:     worker,
		in:         in,
		prevStatus: request.BuildScheduling,
	}, nil
}

func (p statusEventsPiper) PipeMessage() error {
	status, err := p.readStatus()
	if err != nil {
		if errors.Is(err, io.EOF) {
			p.ensureStatusCompletedOrFailed()
		}
		return fmt.Errorf("read status: %w", err)
	}

	transformedStatus, err := p.transformStatus(status.Status)
	if err != nil {
		return fmt.Errorf("transform status: %w", err)
	}

	if p.shouldSkip(transformedStatus) {
		return nil
	}

	if err := p.writeStatus(transformedStatus); err != nil {
		return fmt.Errorf("write status: %w", err)
	}

	return nil
}

// Close closes all active streams.
func (p statusEventsPiper) Close() error {
	if err := p.in.CloseSend(); err != nil {
		log.Error().
			WithError(err).
			Message("Failed closing status events stream from worker.")
	}
	return nil
}

func (p statusEventsPiper) readStatus() (*workerclient.StatusEvent, error) {
	if p.in == nil {
		return nil, errors.New("input stream is nil")
	}
	status, err := p.in.Recv()
	if err != nil {
		return nil, err
	}
	return status, nil
}

func (p statusEventsPiper) ensureStatusCompletedOrFailed() error {
	if p.prevStatus != request.BuildCompleted && p.prevStatus != request.BuildFailed {
		finalStatus, _ := p.transformStatus(v1.StatusFailed)
		if err := p.writeStatus(finalStatus); err != nil {
			return fmt.Errorf("write status: %w", err)
		}
	}
	return nil
}

func (p statusEventsPiper) shouldSkip(status request.BuildStatus) bool {
	return status == p.prevStatus
}

func (p statusEventsPiper) transformStatus(status v1.Status) (request.BuildStatus, error) {
	switch status {
	case v1.StatusPending, v1.StatusScheduling:
		return request.BuildScheduling, nil
	case v1.StatusRunning, v1.StatusInitializing:
		return request.BuildRunning, nil
	case v1.StatusSuccess:
		return request.BuildCompleted, nil
	case v1.StatusCancelled, v1.StatusFailed:
		return request.BuildFailed, nil
	default:
		return "", fmt.Errorf("unsupported status %q", status)
	}
}

func (p statusEventsPiper) writeStatus(status request.BuildStatus) error {
	statusUpdate := request.LogOrStatusUpdate{Status: status}
	_, err := p.wharfapi.UpdateBuildStatus(p.worker.BuildID(), statusUpdate)
	if err != nil {
		p.prevStatus = status
	}
	return err
}
