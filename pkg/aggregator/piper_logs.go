package aggregator

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/model/request"
	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/wharfapi"
	v1 "github.com/iver-wharf/wharf-cmd/api/workerapi/v1"
	"github.com/iver-wharf/wharf-cmd/pkg/workerapi/workerclient"
)

type logsPiper struct {
	wharfapi wharfapi.Client
	worker   workerclient.Client
	in       v1.Worker_StreamLogsClient
	out      logsWriter
}

func newLogsPiper(ctx context.Context, wharfapi wharfapi.Client, worker workerclient.Client, buildID uint) (logsPiper, error) {
	out, err := newLogsWriter(ctx, wharfapi, buildID)
	if err != nil {
		return logsPiper{}, err
	}
	in, err := worker.StreamLogs(ctx, &workerclient.LogsRequest{})
	if err != nil {
		return logsPiper{}, fmt.Errorf("open logs stream from wharf-cmd-worker: %w", err)
	}
	return logsPiper{
		wharfapi: wharfapi,
		worker:   worker,
		in:       in,
		out:      out,
	}, nil
}

func (p logsPiper) PipeMessage() error {
	msg, err := p.readLog()
	if err != nil {
		return fmt.Errorf("read log: %w", err)
	}
	if err := p.out.write(msg); err != nil {
		return fmt.Errorf("write log: %w", err)
	}
	return nil
}

// Close closes all active streams.
func (p logsPiper) Close() error {
	if p.in != nil {
		if err := p.in.CloseSend(); err != nil {
			log.Error().WithError(err).Message("Failed closing log stream from worker.")
		}
	}
	return p.out.Close()
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

type logsWriter struct {
	s       wharfapi.CreateBuildLogStream
	buildID uint

	io.Closer
}

func newLogsWriter(ctx context.Context, wharfapi wharfapi.Client, buildID uint) (logsWriter, error) {
	s, err := wharfapi.CreateBuildLogStream(ctx)
	if err != nil {
		return logsWriter{}, fmt.Errorf("open logs stream to wharf-api: %w", err)
	}
	return logsWriter{
		s:       s,
		buildID: buildID,
	}, nil
}

func (w logsWriter) transform(msg any) (request.Log, error) {
	switch m := msg.(type) {
	case string:
		return request.Log{
			Message:   m,
			BuildID:   w.buildID,
			Timestamp: time.Now(),
		}, nil
	case *workerclient.LogLine:
		return request.Log{
			BuildID:      w.buildID,
			WorkerLogID:  uint(m.LogID),
			WorkerStepID: uint(m.StepID),
			Timestamp:    m.Timestamp.AsTime(),
			Message:      m.Message,
		}, nil
	default:
		return request.Log{}, errors.New("unsupported log type")
	}
}

func (w logsWriter) write(msg any) error {
	if w.s == nil {
		return errors.New("output stream is nil")
	}
	m, err := w.transform(msg)
	if err != nil {
		return fmt.Errorf("transform log: %w", err)
	}
	return w.s.Send(m)
}

func (w logsWriter) pipeAndCloseReader(reader io.ReadCloser) error {
	defer reader.Close()

	empty := true
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		txt := scanner.Text()
		idx := strings.LastIndexByte(txt, '\r')
		if idx != -1 {
			txt = txt[idx+1:]
		}
		if err := w.write(txt); err != nil {
			return err
		}
		if err := w.write("\n"); err != nil {
			return err
		}
		empty = false
	}
	if scanner.Err() != io.EOF {
		return scanner.Err()
	}
	if empty {
		return w.write("<none>")
	}
	return nil
}

func (w logsWriter) Close() error {
	summary, err := w.s.CloseAndRecv()
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
