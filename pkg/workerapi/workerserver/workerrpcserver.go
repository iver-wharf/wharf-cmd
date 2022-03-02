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

func (s *workerRPCServer) StreamLogs(_ *v1.StreamLogsRequest, stream v1.Worker_StreamLogsServer) error {
	const bufferSize = 100
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

func (s *workerRPCServer) StreamStatusEvents(_ *v1.StreamStatusEventsRequest, stream v1.Worker_StreamStatusEventsServer) error {
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

func (s *workerRPCServer) StreamArtifactEvents(_ *v1.StreamArtifactEventsRequest, stream v1.Worker_StreamArtifactEventsServer) error {
	// Doesn't exist in resultstore currently so mocking it directly here.
	var ch <-chan *v1.StreamArtifactEventsResponse
	go func() {
		sendCh := make(chan *v1.StreamArtifactEventsResponse)
		ch = sendCh
		for i := 1; i <= 10; i++ {
			sendCh <- &v1.StreamArtifactEventsResponse{
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

func convertToResultStoreLogLine(line resultstore.LogLine) *v1.StreamLogsResponse {
	return &v1.StreamLogsResponse{
		LogID:     line.LogID,
		StepID:    line.StepID,
		Timestamp: timestamppb.New(line.Timestamp),
		Message:   line.Message,
	}
}

func convertToResultStoreStatusEvent(update resultstore.StatusUpdate) *v1.StreamStatusEventsResponse {
	return &v1.StreamStatusEventsResponse{
		EventID: update.UpdateID,
		StepID:  update.StepID,
		Status:  convertToWorkerStatus(update.Status),
	}
}

func convertToWorkerStatus(status worker.Status) v1.StreamStatusEventsResponseStatus {
	switch status {
	case worker.StatusNone:
		return v1.StreamStatusEventsResponseNone
	case worker.StatusScheduling:
		return v1.StreamStatusEventsResponseScheduling
	case worker.StatusInitializing:
		return v1.StreamStatusEventsResponseInitializing
	case worker.StatusRunning:
		return v1.StreamStatusEventsResponseRunning
	case worker.StatusSuccess:
		return v1.StreamStatusEventsResponseSuccess
	case worker.StatusFailed:
		return v1.StreamStatusEventsResponseFailed
	case worker.StatusCancelled:
		return v1.StreamStatusEventsResponseCancelled
	default:
		return v1.StreamStatusEventsResponseUnknown
	}
}

// func convertArtifactEvent(ev resultstore.ArtifactEvent) v1.StreamArtifactEventsResponse {
// 	return v1.StreamArtifactEventsResponse{
// 		ArtifactID: ev.ArtifactID,
// 		StepID:     ev.StepID,
// 		Name:       ev.Name,
// 	}
// }
