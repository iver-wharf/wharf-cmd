package workerserver

import (
	"github.com/golang/protobuf/ptypes/timestamp"
	v1 "github.com/iver-wharf/wharf-cmd/api/workerapi/v1"
)

type workerServer struct {
	v1.UnimplementedWorkerServer
}

func (s *workerServer) StreamLogs(_ *v1.StreamLogsRequest, stream v1.Worker_StreamLogsServer) error {
	streamLogResponses := []*v1.StreamLogsResponse{
		{
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
		},
		{
			Logs: []*v1.LogLine{
				{
					LogId:     3,
					StepId:    2,
					Timestamp: &timestamp.Timestamp{},
					Line:      "Chunk 2 first text here",
				},
				{
					LogId:     4,
					StepId:    3,
					Timestamp: &timestamp.Timestamp{},
					Line:      "Chunk 2 second text here",
				},
			},
		},
	}

	for _, streamLogResponse := range streamLogResponses {
		if err := stream.Send(streamLogResponse); err != nil {
			return err
		}
	}

	return nil
}

func (s *workerServer) Log(_ *v1.LogRequest, stream v1.Worker_LogServer) error {
	logs := []*v1.LogResponse{
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

func (s *workerServer) StatusEvent(_ *v1.StatusEventRequest, stream v1.Worker_StatusEventServer) error {
	statuses := []*v1.StatusEventResponse{
		{
			EventId: 1,
			StepId:  2,
			Status:  v1.StatusEventResponse_RUNNING,
		},
		{
			EventId: 2,
			StepId:  2,
			Status:  v1.StatusEventResponse_COMPLETED,
		},
	}

	for _, status := range statuses {
		if err := stream.Send(status); err != nil {
			return err
		}
	}

	return nil
}

func (s *workerServer) ArtifactEvent(_ *v1.ArtifactEventRequest, stream v1.Worker_ArtifactEventServer) error {
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

func newWorkerServer() *workerServer {
	return &workerServer{}
}
