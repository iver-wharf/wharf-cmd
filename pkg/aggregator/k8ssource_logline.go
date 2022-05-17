package aggregator

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/model/request"
	"github.com/iver-wharf/wharf-cmd/pkg/workerapi/workerclient"
)

type logLineSource struct {
	ctx    context.Context
	worker workerclient.Client
}

var _ Source[request.Log] = logLineSource{}

func (s logLineSource) PushInto(dst chan<- request.Log) error {
	reader, err := s.worker.StreamLogs(s.ctx, &workerclient.LogsRequest{})
	if err != nil {
		return fmt.Errorf("open logs stream from wharf-cmd-worker: %w", err)
	}
	defer reader.CloseSend()

	buildID := s.worker.BuildID()
	var sentLogs uint
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
		dst <- request.Log{
			BuildID:      buildID,
			WorkerLogID:  uint(logLine.LogID),
			WorkerStepID: uint(logLine.StepID),
			Timestamp:    logLine.Timestamp.AsTime(),
			Message:      logLine.Message,
		}
		sentLogs++
	}

	log.Debug().
		WithUint("sent", sentLogs).
		Message("Sent logs to wharf-api.")

	return nil
}
