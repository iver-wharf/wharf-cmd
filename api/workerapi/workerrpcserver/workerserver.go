package workerrpcserver

import (
	"strconv"

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

func (s *workerServer) StreamLogs(req *v1.StreamLogLineRequest, stream v1.Worker_StreamLogsServer) error {
	bufferSize := 100
	ch, err := s.store.SubAllLogLines(bufferSize)
	if err != nil {
		return err
	}
	defer func() {
		if !s.store.UnsubAllLogLines(ch) {
			log.Warn().Message("Attempted to unsubscribe a non-subscribed channel.")
		}
	}()

	var resp *v1.LogLine
	for {
		select {
		case logLine, ok := <-ch:
			if !ok {
				return nil
			}
			resp = convertToLogLine(logLine)
		default:
			continue
		}

		if err := stream.Send(resp); err != nil {
			return err
		}
	}
}

func (s *workerServer) StreamStatusEvents(_ *v1.StreamStatusEventRequest, stream v1.Worker_StreamStatusEventsServer) error {
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

	var resp *v1.StatusEvent
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

func (s *workerServer) StreamArtifactEvents(_ *v1.StreamArtifactEventRequest, stream v1.Worker_StreamArtifactEventsServer) error {
	// Doesn't exist in resultstore currently so mocking it directly here.
	var ch <-chan *v1.ArtifactEvent
	go func() {
		sendCh := make(chan *v1.ArtifactEvent)
		ch = sendCh
		for i := 1; i <= 10; i++ {
			sendCh <- &v1.ArtifactEvent{
				ArtifactID: uint32(i),
				StepID:     uint32(i/3) + 1,
				Name:       "Artifact " + strconv.Itoa(i),
			}
		}
		close(sendCh)
	}()

	var resp *v1.ArtifactEvent
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

func convertToLogLine(line resultstore.LogLine) *v1.LogLine {
	return &v1.LogLine{
		LogID:     line.LogID,
		StepID:    line.StepID,
		Timestamp: timestamppb.New(line.Timestamp),
		Message:   line.Message,
	}
}

func convertToStatus(status worker.Status) v1.StatusEventStatus {
	switch status {
	case worker.StatusNone:
		return v1.StatusEventNone
	case worker.StatusScheduling:
		return v1.StatusEventScheduling
	case worker.StatusInitializing:
		return v1.StatusEventInitializing
	case worker.StatusRunning:
		return v1.StatusEventRunning
	case worker.StatusSuccess:
		return v1.StatusEventSuccess
	case worker.StatusFailed:
		return v1.StatusEventFailed
	case worker.StatusCancelled:
		return v1.StatusEventCancelled
	default:
		return v1.StatusEventUnknown
	}
}

func convertToStatusEvent(update resultstore.StatusUpdate) *v1.StatusEvent {
	return &v1.StatusEvent{
		EventID: update.UpdateID,
		StepID:  update.StepID,
		Status:  convertToStatus(update.Status),
	}
}

// func convertArtifactEvent(ev resultstore.ArtifactEvent) v1.ArtifactEvent {
// 	return v1.ArtifactEvent{
// 		ArtifactID: ev.ArtifactID,
// 		StepID:     ev.StepID,
// 		Name:       ev.Name,
// 	}
// }
