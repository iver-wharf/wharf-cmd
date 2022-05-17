package aggregator

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/model/request"
	v1 "github.com/iver-wharf/wharf-cmd/api/workerapi/v1"
	"github.com/iver-wharf/wharf-cmd/pkg/workerapi/workerclient"
)

type statusSource struct {
	ctx    context.Context
	worker workerclient.Client
}

func (s statusSource) pushInto(dst chan<- request.BuildStatus) error {
	reader, err := s.worker.StreamStatusEvents(s.ctx, &workerclient.StatusEventsRequest{})
	if err != nil {
		return fmt.Errorf("open status events stream from wharf-cmd-worker: %w", err)
	}
	defer reader.CloseSend()
	for {
		statusEvent, err := reader.Recv()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return err
			}
			break
		}
		newStatus, ok := convBuildStatus(statusEvent.Status)
		if !ok {
			continue
		}
		log.Debug().
			WithUint64("step", statusEvent.StepID).
			WithStringer("status", statusEvent.Status).
			Message("Received status event.")
		dst <- newStatus
	}
	return nil
}

func convBuildStatus(status v1.Status) (request.BuildStatus, bool) {
	switch status {
	case v1.StatusPending, v1.StatusScheduling:
		return request.BuildScheduling, true
	case v1.StatusRunning, v1.StatusInitializing:
		return request.BuildRunning, true
	case v1.StatusSuccess:
		return request.BuildCompleted, true
	case v1.StatusCancelled, v1.StatusFailed:
		return request.BuildFailed, true
	default:
		return "", false
	}
}
