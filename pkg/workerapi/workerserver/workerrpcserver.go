package workerserver

import (
	"strconv"

	v1 "github.com/iver-wharf/wharf-cmd/api/workerapi/v1"
	"github.com/iver-wharf/wharf-cmd/pkg/resultstore"
	"github.com/iver-wharf/wharf-cmd/pkg/worker"
	"github.com/iver-wharf/wharf-core/pkg/logger"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type workerRPCServer struct {
	v1.UnimplementedWorkerServer
	store resultstore.Store
	log   logger.Logger
}

func newWorkerRPCServer(store resultstore.Store) *workerRPCServer {
	return &workerRPCServer{
		store: store,
		log:   logger.NewScoped("WORKER-RPC-SERVER"),
	}
}

func (s *workerRPCServer) StreamLogs(req *v1.StreamLogLineRequest, stream v1.Worker_StreamLogsServer) error {
	bufferSize := 100
	ch, err := s.store.SubAllLogLines(bufferSize)
	if err != nil {
		return err
	}
	defer func() {
		if !s.store.UnsubAllLogLines(ch) {
			s.log.Warn().Message("Attempted to unsubscribe a non-subscribed channel.")
		}
	}()

	for {
		select {
		case logLine, ok := <-ch:
			if !ok {
				return nil
			}
			if err := stream.Send(convertToResultStoreLogLine(logLine)); err != nil {
				return err
			}
		default:
			continue
		}
	}
}

func (s *workerRPCServer) StreamStatusEvents(_ *v1.StreamStatusEventRequest, stream v1.Worker_StreamStatusEventsServer) error {
	const bufferSize = 100
	ch, err := s.store.SubAllStatusUpdates(bufferSize)
	if err != nil {
		return err
	}
	defer func() {
		if !s.store.UnsubAllStatusUpdates(ch) {
			s.log.Warn().Message("Attempted to unsubscribe a non-subscribed channel.")
		}
	}()

	for {
		select {
		case statusEvent, ok := <-ch:
			if !ok {
				return nil
			}
			if err := stream.Send(convertToResultStoreStatusEvent(statusEvent)); err != nil {
				return err
			}
		default:
			continue
		}
	}
}

func (s *workerRPCServer) StreamArtifactEvents(_ *v1.StreamArtifactEventRequest, stream v1.Worker_StreamArtifactEventsServer) error {
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

	for {
		select {
		case artifactEvent, ok := <-ch:
			if !ok {
				return nil
			}
			if err := stream.Send(artifactEvent); err != nil {
				return err
			}
		default:
			continue
		}
	}
}

func convertToResultStoreLogLine(line resultstore.LogLine) *v1.LogLine {
	return &v1.LogLine{
		LogID:     line.LogID,
		StepID:    line.StepID,
		Timestamp: timestamppb.New(line.Timestamp),
		Message:   line.Message,
	}
}

func convertToResultStoreStatusEvent(update resultstore.StatusUpdate) *v1.StatusEvent {
	return &v1.StatusEvent{
		EventID: update.UpdateID,
		StepID:  update.StepID,
		Status:  convertToWorkerStatus(update.Status),
	}
}

func convertToWorkerStatus(status worker.Status) v1.StatusEventStatus {
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

// func convertArtifactEvent(ev resultstore.ArtifactEvent) v1.ArtifactEvent {
// 	return v1.ArtifactEvent{
// 		ArtifactID: ev.ArtifactID,
// 		StepID:     ev.StepID,
// 		Name:       ev.Name,
// 	}
// }
