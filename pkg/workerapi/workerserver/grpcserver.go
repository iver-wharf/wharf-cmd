package workerserver

import (
	"net"
	"time"

	v1 "github.com/iver-wharf/wharf-cmd/api/workerapi/v1"
	"github.com/iver-wharf/wharf-cmd/pkg/resultstore"
	"github.com/iver-wharf/wharf-cmd/pkg/worker/workermodel"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func serveGRPC(grpcWorkerServer *grpcWorkerServer, listener net.Listener) error {
	grpcServer := grpc.NewServer()
	v1.RegisterWorkerServer(grpcServer, grpcWorkerServer)
	grpcWorkerServer.grpc = grpcServer
	return grpcServer.Serve(listener)
}

type grpcWorkerServer struct {
	v1.UnimplementedWorkerServer
	store resultstore.Store

	grpc *grpc.Server
}

func newGRPCServer(store resultstore.Store) *grpcWorkerServer {
	return &grpcWorkerServer{
		store: store,
	}
}

func (s *grpcWorkerServer) StreamLogs(_ *v1.StreamLogsRequest, stream v1.Worker_StreamLogsServer) error {
	const bufferSize = 100
	ch, err := s.store.SubAllLogLines(bufferSize)
	if err != nil {
		return err
	}
	defer unsubWithErrorHandle(ch, s.store.UnsubAllLogLines)
	lastMsg := time.Now()
	timeout := make(chan bool)
	go func() {
		for {
			time.Sleep(10 * time.Millisecond)
			if time.Since(lastMsg) > 5*time.Second {
				timeout <- true
			}
		}
	}()
	for {
		select {
		case logLine, ok := <-ch:
			log.Info().WithBool("OK", ok).WithStringer("line", logLine).Message("Received from channel")
			if !ok {
				// Never getting reached because channel doesn't close when resultstore is "done".
				log.Info().Message("EOF reached - StreamLogs")
				return nil
			}
			if err := stream.Send(ConvertToStreamLogsResponse(logLine)); err != nil {
				log.Error().WithError(err).Message("Error - StreamLogs")
				return err
			}
			lastMsg = time.Now()
		case <-timeout:
			return nil
		default:
			continue
		}
	}
}

func (s *grpcWorkerServer) StreamStatusEvents(_ *v1.StreamStatusEventsRequest, stream v1.Worker_StreamStatusEventsServer) error {
	const bufferSize = 100
	ch, err := s.store.SubAllStatusUpdates(bufferSize)
	if err != nil {
		return err
	}
	defer unsubWithErrorHandle(ch, s.store.UnsubAllStatusUpdates)

	for {
		select {
		case artifactEvent, ok := <-ch:
			if !ok {
				log.Info().Message("EOF reached - StreamStatusEvents")
				return nil
			}
			if err := stream.Send(ConvertToStreamStatusEventsResponse(artifactEvent)); err != nil {
				log.Error().WithError(err).Message("Error - StreamStatusEvents")
				return err
			}
		}
	}
}

func (s *grpcWorkerServer) StreamArtifactEvents(_ *v1.StreamArtifactEventsRequest, stream v1.Worker_StreamArtifactEventsServer) error {
	const bufferSize = 100
	ch, err := s.store.SubAllArtifactEvents(bufferSize)
	if err != nil {
		return err
	}
	defer unsubWithErrorHandle(ch, s.store.UnsubAllArtifactEvents)

	for {
		select {
		case artifactEvent, ok := <-ch:
			if !ok {
				log.Info().Message("EOF reached - StreamArtifactEvents")
				return nil
			}
			if err := stream.Send(ConvertToStreamArtifactEventsResponse(artifactEvent)); err != nil {
				log.Error().WithError(err).Message("Error - StreamArtifactEvents")
				return err
			}
		}
	}
}

func unsubWithErrorHandle[E any](ch <-chan E, unsub func(ch <-chan E) error) {
	if err := unsub(ch); err != nil {
		log.Warn().WithError(err).Message("Failed to unsubscribe channel.")
	}
}

// ConvertToStreamLogsResponse converts a resultstore log line to the equivalent response type.
func ConvertToStreamLogsResponse(line resultstore.LogLine) *v1.StreamLogsResponse {
	return &v1.StreamLogsResponse{
		LogID:     line.LogID,
		StepID:    line.StepID,
		Timestamp: timestamppb.New(line.Timestamp),
		Message:   line.Message,
	}
}

// ConvertToStreamStatusEventsResponse converts a resultstore status update to the equivalent response type.
func ConvertToStreamStatusEventsResponse(update resultstore.StatusUpdate) *v1.StreamStatusEventsResponse {
	return &v1.StreamStatusEventsResponse{
		EventID: update.UpdateID,
		StepID:  update.StepID,
		Status:  convertToStreamStatusEventsResponseStatus(update.Status),
	}
}

func convertToStreamStatusEventsResponseStatus(status workermodel.Status) v1.StreamStatusEventsResponseStatus {
	switch status {
	case workermodel.StatusNone:
		return v1.StreamStatusEventsResponsePending
	case workermodel.StatusScheduling:
		return v1.StreamStatusEventsResponseScheduling
	case workermodel.StatusInitializing:
		return v1.StreamStatusEventsResponseInitializing
	case workermodel.StatusRunning:
		return v1.StreamStatusEventsResponseRunning
	case workermodel.StatusSuccess:
		return v1.StreamStatusEventsResponseSuccess
	case workermodel.StatusFailed:
		return v1.StreamStatusEventsResponseFailed
	case workermodel.StatusCancelled:
		return v1.StreamStatusEventsResponseCancelled
	default:
		return v1.StreamStatusEventsResponseUnspecified
	}
}

// ConvertToStreamArtifactEventsResponse converts a resultstore artifact event
// to the equivalent response type.
func ConvertToStreamArtifactEventsResponse(event resultstore.ArtifactEvent) *v1.StreamArtifactEventsResponse {
	return &v1.StreamArtifactEventsResponse{
		ArtifactID: event.ArtifactID,
		StepID:     event.StepID,
		Name:       event.Name,
	}
}
