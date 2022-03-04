// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package v1

import (
	context "context"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// WorkerClient is the client API for Worker service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type WorkerClient interface {
	StreamLogs(ctx context.Context, in *StreamLogsRequest, opts ...grpc.CallOption) (Worker_StreamLogsClient, error)
	StreamStatusEvents(ctx context.Context, in *StreamStatusEventsRequest, opts ...grpc.CallOption) (Worker_StreamStatusEventsClient, error)
	StreamArtifactEvents(ctx context.Context, in *StreamArtifactEventsRequest, opts ...grpc.CallOption) (Worker_StreamArtifactEventsClient, error)
}

type workerClient struct {
	cc grpc.ClientConnInterface
}

func NewWorkerClient(cc grpc.ClientConnInterface) WorkerClient {
	return &workerClient{cc}
}

func (c *workerClient) StreamLogs(ctx context.Context, in *StreamLogsRequest, opts ...grpc.CallOption) (Worker_StreamLogsClient, error) {
	stream, err := c.cc.NewStream(ctx, &Worker_ServiceDesc.Streams[0], "/wharf.worker.v1.Worker/StreamLogs", opts...)
	if err != nil {
		return nil, err
	}
	x := &workerStreamLogsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Worker_StreamLogsClient interface {
	Recv() (*StreamLogsResponse, error)
	grpc.ClientStream
}

type workerStreamLogsClient struct {
	grpc.ClientStream
}

func (x *workerStreamLogsClient) Recv() (*StreamLogsResponse, error) {
	m := new(StreamLogsResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *workerClient) StreamStatusEvents(ctx context.Context, in *StreamStatusEventsRequest, opts ...grpc.CallOption) (Worker_StreamStatusEventsClient, error) {
	stream, err := c.cc.NewStream(ctx, &Worker_ServiceDesc.Streams[1], "/wharf.worker.v1.Worker/StreamStatusEvents", opts...)
	if err != nil {
		return nil, err
	}
	x := &workerStreamStatusEventsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Worker_StreamStatusEventsClient interface {
	Recv() (*StreamStatusEventsResponse, error)
	grpc.ClientStream
}

type workerStreamStatusEventsClient struct {
	grpc.ClientStream
}

func (x *workerStreamStatusEventsClient) Recv() (*StreamStatusEventsResponse, error) {
	m := new(StreamStatusEventsResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *workerClient) StreamArtifactEvents(ctx context.Context, in *StreamArtifactEventsRequest, opts ...grpc.CallOption) (Worker_StreamArtifactEventsClient, error) {
	stream, err := c.cc.NewStream(ctx, &Worker_ServiceDesc.Streams[2], "/wharf.worker.v1.Worker/StreamArtifactEvents", opts...)
	if err != nil {
		return nil, err
	}
	x := &workerStreamArtifactEventsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Worker_StreamArtifactEventsClient interface {
	Recv() (*StreamArtifactEventsResponse, error)
	grpc.ClientStream
}

type workerStreamArtifactEventsClient struct {
	grpc.ClientStream
}

func (x *workerStreamArtifactEventsClient) Recv() (*StreamArtifactEventsResponse, error) {
	m := new(StreamArtifactEventsResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// WorkerServer is the server API for Worker service.
// All implementations must embed UnimplementedWorkerServer
// for forward compatibility
type WorkerServer interface {
	StreamLogs(*StreamLogsRequest, Worker_StreamLogsServer) error
	StreamStatusEvents(*StreamStatusEventsRequest, Worker_StreamStatusEventsServer) error
	StreamArtifactEvents(*StreamArtifactEventsRequest, Worker_StreamArtifactEventsServer) error
	mustEmbedUnimplementedWorkerServer()
}

// UnimplementedWorkerServer must be embedded to have forward compatible implementations.
type UnimplementedWorkerServer struct {
}

func (UnimplementedWorkerServer) StreamLogs(*StreamLogsRequest, Worker_StreamLogsServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamLogs not implemented")
}
func (UnimplementedWorkerServer) StreamStatusEvents(*StreamStatusEventsRequest, Worker_StreamStatusEventsServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamStatusEvents not implemented")
}
func (UnimplementedWorkerServer) StreamArtifactEvents(*StreamArtifactEventsRequest, Worker_StreamArtifactEventsServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamArtifactEvents not implemented")
}
func (UnimplementedWorkerServer) mustEmbedUnimplementedWorkerServer() {}

// UnsafeWorkerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to WorkerServer will
// result in compilation errors.
type UnsafeWorkerServer interface {
	mustEmbedUnimplementedWorkerServer()
}

func RegisterWorkerServer(s grpc.ServiceRegistrar, srv WorkerServer) {
	s.RegisterService(&Worker_ServiceDesc, srv)
}

func _Worker_StreamLogs_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(StreamLogsRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(WorkerServer).StreamLogs(m, &workerStreamLogsServer{stream})
}

type Worker_StreamLogsServer interface {
	Send(*StreamLogsResponse) error
	grpc.ServerStream
}

type workerStreamLogsServer struct {
	grpc.ServerStream
}

func (x *workerStreamLogsServer) Send(m *StreamLogsResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _Worker_StreamStatusEvents_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(StreamStatusEventsRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(WorkerServer).StreamStatusEvents(m, &workerStreamStatusEventsServer{stream})
}

type Worker_StreamStatusEventsServer interface {
	Send(*StreamStatusEventsResponse) error
	grpc.ServerStream
}

type workerStreamStatusEventsServer struct {
	grpc.ServerStream
}

func (x *workerStreamStatusEventsServer) Send(m *StreamStatusEventsResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _Worker_StreamArtifactEvents_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(StreamArtifactEventsRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(WorkerServer).StreamArtifactEvents(m, &workerStreamArtifactEventsServer{stream})
}

type Worker_StreamArtifactEventsServer interface {
	Send(*StreamArtifactEventsResponse) error
	grpc.ServerStream
}

type workerStreamArtifactEventsServer struct {
	grpc.ServerStream
}

func (x *workerStreamArtifactEventsServer) Send(m *StreamArtifactEventsResponse) error {
	return x.ServerStream.SendMsg(m)
}

// Worker_ServiceDesc is the grpc.ServiceDesc for Worker service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Worker_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "wharf.worker.v1.Worker",
	HandlerType: (*WorkerServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamLogs",
			Handler:       _Worker_StreamLogs_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "StreamStatusEvents",
			Handler:       _Worker_StreamStatusEvents_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "StreamArtifactEvents",
			Handler:       _Worker_StreamArtifactEvents_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "api/workerapi/v1/worker.proto",
}