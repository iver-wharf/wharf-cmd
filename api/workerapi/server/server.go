package server

import (
	"github.com/golang/protobuf/ptypes/timestamp"
	v1 "github.com/iver-wharf/wharf-cmd/api/workerapi/v1"
)

type workerServer struct {
	v1.UnimplementedWorkerServer
}

func (s *workerServer) StreamLogs(_ *v1.StreamLogsRequest, stream v1.Worker_StreamLogsServer) error {
	streamLogsResponse := &v1.StreamLogsResponse{
		Logs: []*v1.LogLine{
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
		},
	}

	return stream.Send(streamLogsResponse)
}

func (s *workerServer) Logs(_ *v1.LogsRequest, stream v1.Worker_LogsServer) error {
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

func (s *workerServer) StatusEvents(_ *v1.StatusEventsRequest, stream v1.Worker_StatusEventsServer) error {
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

func (s *workerServer) ArtifactEvents(_ *v1.ArtifactEventsRequest, stream v1.Worker_ArtifactEventsServer) error {
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
