package workerserver

import (
	"errors"
	"net"

	v1 "github.com/iver-wharf/wharf-cmd/api/workerapi/v1"
	"github.com/iver-wharf/wharf-cmd/pkg/resultstore"
	"github.com/iver-wharf/wharf-cmd/pkg/worker/workermodel"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/typ.v3/pkg/chans"
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
		log.Error().WithError(err).Message("sub log lines error")
		return err
	}
	for {
		select {
		case logLine, ok := <-ch:
			log.Debug().WithStringf("line", "%v", logLine).WithBool("OK", ok).Message("channel receive")
			if !ok {
				return nil
			}
			if err := stream.Send(ConvertToStreamLogsResponse(logLine)); err != nil {
				log.Error().WithError(err).Message("Failed sending logs to client.")
				if err := s.store.UnsubAllLogLines(ch); err != nil && !errors.Is(err, chans.ErrAlreadyUnsubscribed) {
					log.Warn().WithError(err).Message("Failed to unsubscribe channel.")
				}
				return err
			}
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

	for {
		select {
		case statusEvent, ok := <-ch:
			if !ok {
				return nil
			}
			if err := stream.Send(ConvertToStreamStatusEventsResponse(statusEvent)); err != nil {
				log.Error().WithError(err).Message("Failed sending status events to client.")
				if err := s.store.UnsubAllStatusUpdates(ch); err != nil && !errors.Is(err, chans.ErrAlreadyUnsubscribed) {
					log.Warn().WithError(err).Message("Failed to unsubscribe channel.")
				}
				return err
			}
		default:
			continue
		}
	}
}

func (s *grpcWorkerServer) StreamArtifactEvents(_ *v1.StreamArtifactEventsRequest, stream v1.Worker_StreamArtifactEventsServer) error {
	const bufferSize = 100
	ch, err := s.store.SubAllArtifactEvents(bufferSize)
	if err != nil {
		return err
	}
	for {
		select {
		case artifactEvent, ok := <-ch:
			if !ok {
				return nil
			}
			if err := stream.Send(ConvertToStreamArtifactEventsResponse(artifactEvent)); err != nil {
				log.Error().WithError(err).Message("Failed sending artifact events to client.")
				if err := s.store.UnsubAllArtifactEvents(ch); err != nil && !errors.Is(err, chans.ErrAlreadyUnsubscribed) {
					log.Warn().WithError(err).Message("Failed to unsubscribe channel.")
				}
				return err
			}
		default:
			continue
		}
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
