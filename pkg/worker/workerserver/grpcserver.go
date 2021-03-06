package workerserver

import (
	"net"

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
	defer s.store.UnsubAllLogLines(ch)

	for {
		select {
		case logLine, ok := <-ch:
			if !ok {
				return nil
			}
			if err := stream.Send(ConvertToStreamLogsResponse(logLine)); err != nil {
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
	defer s.store.UnsubAllStatusUpdates(ch)

	for {
		select {
		case statusEvent, ok := <-ch:
			if !ok {
				return nil
			}
			if err := stream.Send(ConvertToStreamStatusEventsResponse(statusEvent)); err != nil {
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
	defer s.store.UnsubAllArtifactEvents(ch)

	for {
		select {
		case artifactEvent, ok := <-ch:
			if !ok {
				return nil
			}
			if err := stream.Send(ConvertToStreamArtifactEventsResponse(artifactEvent)); err != nil {
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

func convertToStreamStatusEventsResponseStatus(status workermodel.Status) v1.Status {
	switch status {
	case workermodel.StatusNone:
		return v1.StatusPending
	case workermodel.StatusScheduling:
		return v1.StatusScheduling
	case workermodel.StatusInitializing:
		return v1.StatusInitializing
	case workermodel.StatusRunning:
		return v1.StatusRunning
	case workermodel.StatusSuccess:
		return v1.StatusSuccess
	case workermodel.StatusFailed:
		return v1.StatusFailed
	case workermodel.StatusCancelled:
		return v1.StatusCancelled
	default:
		return v1.StatusUnspecified
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
