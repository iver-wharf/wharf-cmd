package workerserver

import (
	"strconv"
	"time"

	v1 "github.com/iver-wharf/wharf-cmd/api/workerapi/v1"
	"github.com/iver-wharf/wharf-cmd/pkg/resultstore"
	"github.com/iver-wharf/wharf-cmd/pkg/worker"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type workerServer struct {
	v1.UnimplementedWorkerServer
	store resultstore.Store
}

func newWorkerServer(store resultstore.Store) *workerServer {
	return &workerServer{store: store}
}

func (s *workerServer) StreamLogs(req *v1.StreamLogsRequest, stream v1.Worker_StreamLogsServer) error {
	bufferSize := 100
	if req.ChunkSize > 0 {
		bufferSize = int(req.ChunkSize)
	}
	ch, err := s.store.SubAllLogLines(bufferSize)
	if err != nil {
		return err
	}
	defer func() {
		if !s.store.UnsubAllLogLines(ch) {
			log.Warn().Message("Attempted to unsubscribe a non-subscribed channel.")
		}
	}()

	logs := make([]*v1.LogLine, bufferSize, bufferSize)
	resp := v1.StreamLogsResponse{}

	i := 0
	send := func() error {
		// Slice slice to avoid sending old objects when we don't have a full chunk.
		resp.Logs = logs[:i]
		if len(resp.Logs) > 0 {
			if err := stream.Send(&resp); err != nil {
				return err
			}
			i = 0
		}
		return nil
	}
outer:
	for {
		for i < bufferSize {
			select {
			case line, ok := <-ch:
				if !ok {
					break outer
				}
				logs[i] = convertToLogLine(line)
				i++
			default:
				break
			}
		}
		if err := send(); err != nil {
			return err
		}
	}
	if err := send(); err != nil {
		return err
	}
	return nil
}

func (s *workerServer) Log(_ *v1.LogRequest, stream v1.Worker_LogServer) error {
	const bufferSize = 100
	ch, err := s.store.SubAllLogLines(bufferSize)
	if err != nil {
		return err
	}
	defer func() {
		if !s.store.UnsubAllLogLines(ch) {
			log.Warn().Message("Attempted to unsubscribe a non-subscribed channel.")
		}
	}()

	var resp *v1.LogResponse
	for {
		select {
		case line, ok := <-ch:
			if !ok {
				return nil
			}
			resp = convertToLogResponse(line)
		default:
			continue
		}

		if err := stream.Send(resp); err != nil {
			return err
		}
	}
}

func (s *workerServer) StatusEvent(_ *v1.StatusEventRequest, stream v1.Worker_StatusEventServer) error {
	const bufferSize = 100
	ch, err := s.store.SubAllStatusUpdates(bufferSize)
	if err != nil {
		return err
	}
	defer func() {
		if !s.store.UnsubAllStatusUpdates(ch) {
			log.Warn().Message("Attempted to unsubscribe a non-subscribed channel.")
		}
	}()

	var resp *v1.StatusEventResponse
	for {
		select {
		case statusEvent, ok := <-ch:
			if !ok {
				return nil
			}
			resp = convertToStatusEvent(statusEvent)
		default:
			continue
		}

		if err := stream.Send(resp); err != nil {
			return err
		}
	}
}

func (s *workerServer) ArtifactEvent(_ *v1.ArtifactEventRequest, stream v1.Worker_ArtifactEventServer) error {
	// Doesn't exist in resultstore currently so mocking it directly here.
	var ch <-chan *v1.ArtifactEventResponse
	go func() {
		sendCh := make(chan *v1.ArtifactEventResponse)
		ch = sendCh
		for i := 1; i <= 10; i++ {
			sendCh <- &v1.ArtifactEventResponse{
				ArtifactID: uint32(i),
				StepID:     uint32(i/3) + 1,
				Name:       "Artifact " + strconv.Itoa(i),
			}
		}
		close(sendCh)
	}()

	var resp *v1.ArtifactEventResponse
	for {
		select {
		case artifactEvent, ok := <-ch:
			if !ok {
				return nil
			}
			resp = artifactEvent
		default:
			continue
		}

		if err := stream.Send(resp); err != nil {
			return err
		}
	}
}

func convertToTimestamppb(ts time.Time) *timestamppb.Timestamp {
	return &timestamppb.Timestamp{
		Seconds: ts.Unix(),
		Nanos:   int32(ts.Nanosecond()),
	}
}

func convertToLogLine(line resultstore.LogLine) *v1.LogLine {
	return &v1.LogLine{
		LogID:     line.LogID,
		StepID:    line.StepID,
		Timestamp: convertToTimestamppb(line.Timestamp),
		Line:      line.Message,
	}
}

func convertToLogResponse(line resultstore.LogLine) *v1.LogResponse {
	return &v1.LogResponse{
		LogID:     line.LogID,
		StepID:    line.StepID,
		Timestamp: convertToTimestamppb(line.Timestamp),
		Line:      line.Message,
	}
}

func convertToStatus(status worker.Status) v1.StatusEventResponse_Status {
	switch status {
	case worker.StatusNone:
		return v1.StatusEventResponse_NONE
	case worker.StatusScheduling:
		return v1.StatusEventResponse_SCHEDULING
	case worker.StatusInitializing:
		return v1.StatusEventResponse_INITIALIZING
	case worker.StatusRunning:
		return v1.StatusEventResponse_RUNNING
	case worker.StatusSuccess:
		return v1.StatusEventResponse_SUCCESS
	case worker.StatusFailed:
		return v1.StatusEventResponse_FAILED
	case worker.StatusCancelled:
		return v1.StatusEventResponse_CANCELLED
	default:
		return v1.StatusEventResponse_UNKNOWN
	}
}

func convertToStatusEvent(update resultstore.StatusUpdate) *v1.StatusEventResponse {
	return &v1.StatusEventResponse{
		EventID: update.UpdateID,
		StepID:  update.StepID,
		Status:  convertToStatus(update.Status),
	}
}

// func convertArtifactEvent(ev resultstore.ArtifactEvent) v1.ArtifactEventResponse {
// 	return v1.ArtifactEventResponse{
// 		ArtifactID: ev.ArtifactID,
// 		StepID:     ev.StepID,
// 		Name:       ev.Name,
// 	}
// }
