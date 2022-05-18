package aggregator

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/model/request"
	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/wharfapi"
	v1 "github.com/iver-wharf/wharf-cmd/api/workerapi/v1"
	"github.com/iver-wharf/wharf-cmd/pkg/workerapi/workerclient"
)

func newLogsPiper(ctx context.Context, wharfapi wharfapi.Client, worker workerclient.Client, buildID uint) (logsPiper, error) {
	var in v1.Worker_StreamLogsClient
	var err error
	if worker != nil {
		in, err = worker.StreamLogs(ctx, &workerclient.LogsRequest{})
		if err != nil {
			return logsPiper{}, fmt.Errorf("open logs stream from wharf-cmd-worker: %w", err)
		}
	}
	out, err := wharfapi.CreateBuildLogStream(ctx)
	if err != nil {
		return logsPiper{}, fmt.Errorf("open logs stream to wharf-api: %w", err)
	}
	return logsPiper{
		wharfapi: wharfapi,
		worker:   worker,
		buildID:  buildID,
		in:       in,
		out:      out,
	}, nil
}

type logsPiper struct {
	wharfapi wharfapi.Client
	worker   workerclient.Client
	buildID  uint
	in       v1.Worker_StreamLogsClient
	out      wharfapi.CreateBuildLogStream
}

func (p logsPiper) PipeMessage() error {
	msg, err := p.readLog()
	if err != nil {
		return fmt.Errorf("read log: %w", err)
	}

	transformedMsg, err := p.transformLog(msg)
	if err != nil {
		return fmt.Errorf("transform log: %w", err)
	}

	if err := p.writeLog(transformedMsg); err != nil {
		return fmt.Errorf("write log: %w", err)
	}
	return nil
}

func (p logsPiper) Close() error {
	if p.in != nil {
		if err := p.in.CloseSend(); err != nil {
			log.Error().WithError(err).Message("Failed closing log stream from worker.")
		}
	}
	summary, err := p.out.CloseAndRecv()
	if err != nil {
		log.Warn().
			WithError(err).
			Message("Failed closing log stream to wharf-api.")
		return err
	}
	log.Debug().
		WithUint("inserted", summary.LogsInserted).
		Message("Inserted logs into wharf-api.")
	return nil
}

func (p logsPiper) readLog() (any, error) {
	if p.in == nil {
		return nil, errors.New("input stream is nil")
	}
	logLine, err := p.in.Recv()
	if err != nil {
		return nil, err
	}
	return logLine, nil
}

func (p logsPiper) transformLog(msg any) (request.Log, error) {
	switch m := msg.(type) {
	case string:
		return request.Log{
			Message:   m,
			BuildID:   p.buildID,
			Timestamp: time.Now(),
		}, nil
	case *workerclient.LogLine:
		return request.Log{
			BuildID:      p.buildID,
			WorkerLogID:  uint(m.LogID),
			WorkerStepID: uint(m.StepID),
			Timestamp:    m.Timestamp.AsTime(),
			Message:      m.Message,
		}, nil
	default:
		return request.Log{}, errors.New("unsupported log type")
	}
}

func (p logsPiper) writeLog(msg request.Log) error {
	if p.out == nil {
		return errors.New("output stream is nil")
	}

	return p.out.Send(msg)
}

func (p logsPiper) writeString(s string) error {
	msg, err := p.transformLog(s)
	if err != nil {
		return fmt.Errorf("transform log: %w", err)
	}
	if err := p.writeLog(msg); err != nil {
		return fmt.Errorf("write log: %w", err)
	}
	return nil
}
