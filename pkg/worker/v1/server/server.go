package v1

import (
	"github.com/golang/protobuf/ptypes/timestamp"
	v1 "github.com/iver-wharf/wharf-cmd/pkg/worker/v1/rpc"
)

type workerServer struct {
	v1.UnimplementedWorkerServer
}

func (s *workerServer) Logs(req *v1.LogsRequest, stream v1.Worker_LogsServer) error {
	logs := []*v1.LogsResponse{
		{
			LogId:     1,
			StepId:    2,
			Timestamp: &timestamp.Timestamp{},
			Line:      "Text here",
		},
		{
			LogId:     2,
			StepId:    2,
			Timestamp: &timestamp.Timestamp{},
			Line:      "Second text here",
		},
	}

	for _, log := range logs {
		if err := stream.Send(log); err != nil {
			return err
		}
	}

	return nil
}

func (s *workerServer) StatusEvents(req *v1.StatusEventsRequest, stream v1.Worker_StatusEventsServer) error {
	statuses := []*v1.StatusEventsResponse{
		{
			EventId: 1,
			StepId:  2,
			Status:  v1.StatusEventsResponse_RUNNING,
		},
		{
			EventId: 2,
			StepId:  2,
			Status:  v1.StatusEventsResponse_COMPLETED,
		},
	}

	for _, status := range statuses {
		if err := stream.Send(status); err != nil {
			return err
		}
	}

	return nil
}

func (s *workerServer) ArtifactEvents(req *v1.ArtifactEventsRequest, stream v1.Worker_ArtifactEventsServer) error {
	artifacts := []*v1.ArtifactEventResponse{
		{
			ArtifactId: 1,
			StepId:     2,
			Name:       "An artifact name",
		},
		{
			ArtifactId: 2,
			StepId:     2,
			Name:       "A second artifact name",
		},
	}

	for _, artifact := range artifacts {
		if err := stream.Send(artifact); err != nil {
			return err
		}
	}

	return nil
}

func newServer() *workerServer {
	return &workerServer{}
}
