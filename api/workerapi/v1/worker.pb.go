// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.6.1
// source: api/workerapi/v1/worker.proto

// Versioning via package name, for future-proofing

package v1

import (
	reflect "reflect"
	sync "sync"

	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type StatusEventResponse_Status int32

const (
	StatusEventResponse_SCHEDULING StatusEventResponse_Status = 0
	StatusEventResponse_RUNNING    StatusEventResponse_Status = 1
	StatusEventResponse_FAILED     StatusEventResponse_Status = 2
	StatusEventResponse_COMPLETED  StatusEventResponse_Status = 3
)

// Enum value maps for StatusEventResponse_Status.
var (
	StatusEventResponse_Status_name = map[int32]string{
		0: "SCHEDULING",
		1: "RUNNING",
		2: "FAILED",
		3: "COMPLETED",
	}
	StatusEventResponse_Status_value = map[string]int32{
		"SCHEDULING": 0,
		"RUNNING":    1,
		"FAILED":     2,
		"COMPLETED":  3,
	}
)

func (x StatusEventResponse_Status) Enum() *StatusEventResponse_Status {
	p := new(StatusEventResponse_Status)
	*p = x
	return p
}

func (x StatusEventResponse_Status) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (StatusEventResponse_Status) Descriptor() protoreflect.EnumDescriptor {
	return file_api_workerapi_v1_worker_proto_enumTypes[0].Descriptor()
}

func (StatusEventResponse_Status) Type() protoreflect.EnumType {
	return &file_api_workerapi_v1_worker_proto_enumTypes[0]
}

func (x StatusEventResponse_Status) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use StatusEventResponse_Status.Descriptor instead.
func (StatusEventResponse_Status) EnumDescriptor() ([]byte, []int) {
	return file_api_workerapi_v1_worker_proto_rawDescGZIP(), []int{5, 0}
}

// Empty messages, but exists for potential future usages
type StreamLogsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *StreamLogsRequest) Reset() {
	*x = StreamLogsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_workerapi_v1_worker_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StreamLogsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StreamLogsRequest) ProtoMessage() {}

func (x *StreamLogsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_workerapi_v1_worker_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StreamLogsRequest.ProtoReflect.Descriptor instead.
func (*StreamLogsRequest) Descriptor() ([]byte, []int) {
	return file_api_workerapi_v1_worker_proto_rawDescGZIP(), []int{0}
}

type LogRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *LogRequest) Reset() {
	*x = LogRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_workerapi_v1_worker_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LogRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LogRequest) ProtoMessage() {}

func (x *LogRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_workerapi_v1_worker_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LogRequest.ProtoReflect.Descriptor instead.
func (*LogRequest) Descriptor() ([]byte, []int) {
	return file_api_workerapi_v1_worker_proto_rawDescGZIP(), []int{1}
}

type StatusEventRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *StatusEventRequest) Reset() {
	*x = StatusEventRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_workerapi_v1_worker_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StatusEventRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StatusEventRequest) ProtoMessage() {}

func (x *StatusEventRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_workerapi_v1_worker_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StatusEventRequest.ProtoReflect.Descriptor instead.
func (*StatusEventRequest) Descriptor() ([]byte, []int) {
	return file_api_workerapi_v1_worker_proto_rawDescGZIP(), []int{2}
}

type ArtifactEventRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ArtifactEventRequest) Reset() {
	*x = ArtifactEventRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_workerapi_v1_worker_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ArtifactEventRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ArtifactEventRequest) ProtoMessage() {}

func (x *ArtifactEventRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_workerapi_v1_worker_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ArtifactEventRequest.ProtoReflect.Descriptor instead.
func (*ArtifactEventRequest) Descriptor() ([]byte, []int) {
	return file_api_workerapi_v1_worker_proto_rawDescGZIP(), []int{3}
}

type LogResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	LogId     int64                `protobuf:"varint,1,opt,name=log_id,json=logId,proto3" json:"log_id,omitempty"`    // not DB value; worker's own ID of the log line
	StepId    int32                `protobuf:"varint,2,opt,name=step_id,json=stepId,proto3" json:"step_id,omitempty"` // not DB value; worker's own ID of the step ID
	Timestamp *timestamp.Timestamp `protobuf:"bytes,3,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Line      string               `protobuf:"bytes,4,opt,name=line,proto3" json:"line,omitempty"`
}

func (x *LogResponse) Reset() {
	*x = LogResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_workerapi_v1_worker_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LogResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LogResponse) ProtoMessage() {}

func (x *LogResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_workerapi_v1_worker_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LogResponse.ProtoReflect.Descriptor instead.
func (*LogResponse) Descriptor() ([]byte, []int) {
	return file_api_workerapi_v1_worker_proto_rawDescGZIP(), []int{4}
}

func (x *LogResponse) GetLogId() int64 {
	if x != nil {
		return x.LogId
	}
	return 0
}

func (x *LogResponse) GetStepId() int32 {
	if x != nil {
		return x.StepId
	}
	return 0
}

func (x *LogResponse) GetTimestamp() *timestamp.Timestamp {
	if x != nil {
		return x.Timestamp
	}
	return nil
}

func (x *LogResponse) GetLine() string {
	if x != nil {
		return x.Line
	}
	return ""
}

type StatusEventResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	EventId int32                      `protobuf:"varint,1,opt,name=event_id,json=eventId,proto3" json:"event_id,omitempty"` // not DB value; worker's own ID of the event
	StepId  int32                      `protobuf:"varint,2,opt,name=step_id,json=stepId,proto3" json:"step_id,omitempty"`    // not DB value; worker's own ID of the step ID
	Status  StatusEventResponse_Status `protobuf:"varint,3,opt,name=status,proto3,enum=wharf.worker.v1.StatusEventResponse_Status" json:"status,omitempty"`
}

func (x *StatusEventResponse) Reset() {
	*x = StatusEventResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_workerapi_v1_worker_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StatusEventResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StatusEventResponse) ProtoMessage() {}

func (x *StatusEventResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_workerapi_v1_worker_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StatusEventResponse.ProtoReflect.Descriptor instead.
func (*StatusEventResponse) Descriptor() ([]byte, []int) {
	return file_api_workerapi_v1_worker_proto_rawDescGZIP(), []int{5}
}

func (x *StatusEventResponse) GetEventId() int32 {
	if x != nil {
		return x.EventId
	}
	return 0
}

func (x *StatusEventResponse) GetStepId() int32 {
	if x != nil {
		return x.StepId
	}
	return 0
}

func (x *StatusEventResponse) GetStatus() StatusEventResponse_Status {
	if x != nil {
		return x.Status
	}
	return StatusEventResponse_SCHEDULING
}

type ArtifactEventResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ArtifactId int32  `protobuf:"varint,1,opt,name=artifact_id,json=artifactId,proto3" json:"artifact_id,omitempty"` // not DB value; worker's own ID of the artifact
	StepId     int32  `protobuf:"varint,2,opt,name=step_id,json=stepId,proto3" json:"step_id,omitempty"`             // not DB value; worker's own ID of the step ID
	Name       string `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *ArtifactEventResponse) Reset() {
	*x = ArtifactEventResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_workerapi_v1_worker_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ArtifactEventResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ArtifactEventResponse) ProtoMessage() {}

func (x *ArtifactEventResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_workerapi_v1_worker_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ArtifactEventResponse.ProtoReflect.Descriptor instead.
func (*ArtifactEventResponse) Descriptor() ([]byte, []int) {
	return file_api_workerapi_v1_worker_proto_rawDescGZIP(), []int{6}
}

func (x *ArtifactEventResponse) GetArtifactId() int32 {
	if x != nil {
		return x.ArtifactId
	}
	return 0
}

func (x *ArtifactEventResponse) GetStepId() int32 {
	if x != nil {
		return x.StepId
	}
	return 0
}

func (x *ArtifactEventResponse) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type StreamLogsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Logs []*LogLine `protobuf:"bytes,1,rep,name=logs,proto3" json:"logs,omitempty"`
}

func (x *StreamLogsResponse) Reset() {
	*x = StreamLogsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_workerapi_v1_worker_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StreamLogsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StreamLogsResponse) ProtoMessage() {}

func (x *StreamLogsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_workerapi_v1_worker_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StreamLogsResponse.ProtoReflect.Descriptor instead.
func (*StreamLogsResponse) Descriptor() ([]byte, []int) {
	return file_api_workerapi_v1_worker_proto_rawDescGZIP(), []int{7}
}

func (x *StreamLogsResponse) GetLogs() []*LogLine {
	if x != nil {
		return x.Logs
	}
	return nil
}

type LogLine struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	LogId     int64                `protobuf:"varint,1,opt,name=log_id,json=logId,proto3" json:"log_id,omitempty"`    // not DB value; worker's own ID of the log line
	StepId    int32                `protobuf:"varint,2,opt,name=step_id,json=stepId,proto3" json:"step_id,omitempty"` // not DB value; worker's own ID of the step ID
	Timestamp *timestamp.Timestamp `protobuf:"bytes,3,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Line      string               `protobuf:"bytes,4,opt,name=line,proto3" json:"line,omitempty"`
}

func (x *LogLine) Reset() {
	*x = LogLine{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_workerapi_v1_worker_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LogLine) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LogLine) ProtoMessage() {}

func (x *LogLine) ProtoReflect() protoreflect.Message {
	mi := &file_api_workerapi_v1_worker_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LogLine.ProtoReflect.Descriptor instead.
func (*LogLine) Descriptor() ([]byte, []int) {
	return file_api_workerapi_v1_worker_proto_rawDescGZIP(), []int{8}
}

func (x *LogLine) GetLogId() int64 {
	if x != nil {
		return x.LogId
	}
	return 0
}

func (x *LogLine) GetStepId() int32 {
	if x != nil {
		return x.StepId
	}
	return 0
}

func (x *LogLine) GetTimestamp() *timestamp.Timestamp {
	if x != nil {
		return x.Timestamp
	}
	return nil
}

func (x *LogLine) GetLine() string {
	if x != nil {
		return x.Line
	}
	return ""
}

var File_api_workerapi_v1_worker_proto protoreflect.FileDescriptor

var file_api_workerapi_v1_worker_proto_rawDesc = []byte{
	0x0a, 0x1d, 0x61, 0x70, 0x69, 0x2f, 0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x61, 0x70, 0x69, 0x2f,
	0x76, 0x31, 0x2f, 0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x0f, 0x77, 0x68, 0x61, 0x72, 0x66, 0x2e, 0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x2e, 0x76, 0x31,
	0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0x13, 0x0a, 0x11, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x4c, 0x6f, 0x67, 0x73, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x0c, 0x0a, 0x0a, 0x4c, 0x6f, 0x67, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x22, 0x14, 0x0a, 0x12, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x45, 0x76,
	0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x16, 0x0a, 0x14, 0x41, 0x72,
	0x74, 0x69, 0x66, 0x61, 0x63, 0x74, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x22, 0x8b, 0x01, 0x0a, 0x0b, 0x4c, 0x6f, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x15, 0x0a, 0x06, 0x6c, 0x6f, 0x67, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x05, 0x6c, 0x6f, 0x67, 0x49, 0x64, 0x12, 0x17, 0x0a, 0x07, 0x73, 0x74, 0x65,
	0x70, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x73, 0x74, 0x65, 0x70,
	0x49, 0x64, 0x12, 0x38, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d,
	0x70, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x12, 0x12, 0x0a, 0x04,
	0x6c, 0x69, 0x6e, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6c, 0x69, 0x6e, 0x65,
	0x22, 0xd0, 0x01, 0x0a, 0x13, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x45, 0x76, 0x65, 0x6e, 0x74,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x19, 0x0a, 0x08, 0x65, 0x76, 0x65, 0x6e,
	0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x65, 0x76, 0x65, 0x6e,
	0x74, 0x49, 0x64, 0x12, 0x17, 0x0a, 0x07, 0x73, 0x74, 0x65, 0x70, 0x5f, 0x69, 0x64, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x73, 0x74, 0x65, 0x70, 0x49, 0x64, 0x12, 0x43, 0x0a, 0x06,
	0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x2b, 0x2e, 0x77,
	0x68, 0x61, 0x72, 0x66, 0x2e, 0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x53,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x22, 0x40, 0x0a, 0x06, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x0e, 0x0a, 0x0a, 0x53,
	0x43, 0x48, 0x45, 0x44, 0x55, 0x4c, 0x49, 0x4e, 0x47, 0x10, 0x00, 0x12, 0x0b, 0x0a, 0x07, 0x52,
	0x55, 0x4e, 0x4e, 0x49, 0x4e, 0x47, 0x10, 0x01, 0x12, 0x0a, 0x0a, 0x06, 0x46, 0x41, 0x49, 0x4c,
	0x45, 0x44, 0x10, 0x02, 0x12, 0x0d, 0x0a, 0x09, 0x43, 0x4f, 0x4d, 0x50, 0x4c, 0x45, 0x54, 0x45,
	0x44, 0x10, 0x03, 0x22, 0x65, 0x0a, 0x15, 0x41, 0x72, 0x74, 0x69, 0x66, 0x61, 0x63, 0x74, 0x45,
	0x76, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1f, 0x0a, 0x0b,
	0x61, 0x72, 0x74, 0x69, 0x66, 0x61, 0x63, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x0a, 0x61, 0x72, 0x74, 0x69, 0x66, 0x61, 0x63, 0x74, 0x49, 0x64, 0x12, 0x17, 0x0a,
	0x07, 0x73, 0x74, 0x65, 0x70, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06,
	0x73, 0x74, 0x65, 0x70, 0x49, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x42, 0x0a, 0x12, 0x53, 0x74,
	0x72, 0x65, 0x61, 0x6d, 0x4c, 0x6f, 0x67, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x2c, 0x0a, 0x04, 0x6c, 0x6f, 0x67, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x18,
	0x2e, 0x77, 0x68, 0x61, 0x72, 0x66, 0x2e, 0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x2e, 0x76, 0x31,
	0x2e, 0x4c, 0x6f, 0x67, 0x4c, 0x69, 0x6e, 0x65, 0x52, 0x04, 0x6c, 0x6f, 0x67, 0x73, 0x22, 0x87,
	0x01, 0x0a, 0x07, 0x4c, 0x6f, 0x67, 0x4c, 0x69, 0x6e, 0x65, 0x12, 0x15, 0x0a, 0x06, 0x6c, 0x6f,
	0x67, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x6c, 0x6f, 0x67, 0x49,
	0x64, 0x12, 0x17, 0x0a, 0x07, 0x73, 0x74, 0x65, 0x70, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x06, 0x73, 0x74, 0x65, 0x70, 0x49, 0x64, 0x12, 0x38, 0x0a, 0x09, 0x74, 0x69,
	0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e,
	0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73,
	0x74, 0x61, 0x6d, 0x70, 0x12, 0x12, 0x0a, 0x04, 0x6c, 0x69, 0x6e, 0x65, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x6c, 0x69, 0x6e, 0x65, 0x32, 0xe3, 0x02, 0x0a, 0x06, 0x57, 0x6f, 0x72,
	0x6b, 0x65, 0x72, 0x12, 0x57, 0x0a, 0x0a, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x4c, 0x6f, 0x67,
	0x73, 0x12, 0x22, 0x2e, 0x77, 0x68, 0x61, 0x72, 0x66, 0x2e, 0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72,
	0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x4c, 0x6f, 0x67, 0x73, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x23, 0x2e, 0x77, 0x68, 0x61, 0x72, 0x66, 0x2e, 0x77, 0x6f,
	0x72, 0x6b, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x4c, 0x6f,
	0x67, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x30, 0x01, 0x12, 0x42, 0x0a, 0x03,
	0x4c, 0x6f, 0x67, 0x12, 0x1b, 0x2e, 0x77, 0x68, 0x61, 0x72, 0x66, 0x2e, 0x77, 0x6f, 0x72, 0x6b,
	0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x6f, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x1c, 0x2e, 0x77, 0x68, 0x61, 0x72, 0x66, 0x2e, 0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x2e,
	0x76, 0x31, 0x2e, 0x4c, 0x6f, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x30, 0x01,
	0x12, 0x5a, 0x0a, 0x0b, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12,
	0x23, 0x2e, 0x77, 0x68, 0x61, 0x72, 0x66, 0x2e, 0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x2e, 0x76,
	0x31, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x24, 0x2e, 0x77, 0x68, 0x61, 0x72, 0x66, 0x2e, 0x77, 0x6f, 0x72,
	0x6b, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x45, 0x76, 0x65,
	0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x30, 0x01, 0x12, 0x60, 0x0a, 0x0d,
	0x41, 0x72, 0x74, 0x69, 0x66, 0x61, 0x63, 0x74, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x25, 0x2e,
	0x77, 0x68, 0x61, 0x72, 0x66, 0x2e, 0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e,
	0x41, 0x72, 0x74, 0x69, 0x66, 0x61, 0x63, 0x74, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x26, 0x2e, 0x77, 0x68, 0x61, 0x72, 0x66, 0x2e, 0x77, 0x6f, 0x72,
	0x6b, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x72, 0x74, 0x69, 0x66, 0x61, 0x63, 0x74, 0x45,
	0x76, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x30, 0x01, 0x42, 0x32,
	0x5a, 0x30, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x69, 0x76, 0x65,
	0x72, 0x2d, 0x77, 0x68, 0x61, 0x72, 0x66, 0x2f, 0x77, 0x68, 0x61, 0x72, 0x66, 0x2d, 0x63, 0x6d,
	0x64, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x61, 0x70, 0x69, 0x2f,
	0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_workerapi_v1_worker_proto_rawDescOnce sync.Once
	file_api_workerapi_v1_worker_proto_rawDescData = file_api_workerapi_v1_worker_proto_rawDesc
)

func file_api_workerapi_v1_worker_proto_rawDescGZIP() []byte {
	file_api_workerapi_v1_worker_proto_rawDescOnce.Do(func() {
		file_api_workerapi_v1_worker_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_workerapi_v1_worker_proto_rawDescData)
	})
	return file_api_workerapi_v1_worker_proto_rawDescData
}

var file_api_workerapi_v1_worker_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_api_workerapi_v1_worker_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_api_workerapi_v1_worker_proto_goTypes = []interface{}{
	(StatusEventResponse_Status)(0), // 0: wharf.worker.v1.StatusEventResponse.Status
	(*StreamLogsRequest)(nil),       // 1: wharf.worker.v1.StreamLogsRequest
	(*LogRequest)(nil),              // 2: wharf.worker.v1.LogRequest
	(*StatusEventRequest)(nil),      // 3: wharf.worker.v1.StatusEventRequest
	(*ArtifactEventRequest)(nil),    // 4: wharf.worker.v1.ArtifactEventRequest
	(*LogResponse)(nil),             // 5: wharf.worker.v1.LogResponse
	(*StatusEventResponse)(nil),     // 6: wharf.worker.v1.StatusEventResponse
	(*ArtifactEventResponse)(nil),   // 7: wharf.worker.v1.ArtifactEventResponse
	(*StreamLogsResponse)(nil),      // 8: wharf.worker.v1.StreamLogsResponse
	(*LogLine)(nil),                 // 9: wharf.worker.v1.LogLine
	(*timestamp.Timestamp)(nil),     // 10: google.protobuf.Timestamp
}
var file_api_workerapi_v1_worker_proto_depIdxs = []int32{
	10, // 0: wharf.worker.v1.LogResponse.timestamp:type_name -> google.protobuf.Timestamp
	0,  // 1: wharf.worker.v1.StatusEventResponse.status:type_name -> wharf.worker.v1.StatusEventResponse.Status
	9,  // 2: wharf.worker.v1.StreamLogsResponse.logs:type_name -> wharf.worker.v1.LogLine
	10, // 3: wharf.worker.v1.LogLine.timestamp:type_name -> google.protobuf.Timestamp
	1,  // 4: wharf.worker.v1.Worker.StreamLogs:input_type -> wharf.worker.v1.StreamLogsRequest
	2,  // 5: wharf.worker.v1.Worker.Log:input_type -> wharf.worker.v1.LogRequest
	3,  // 6: wharf.worker.v1.Worker.StatusEvent:input_type -> wharf.worker.v1.StatusEventRequest
	4,  // 7: wharf.worker.v1.Worker.ArtifactEvent:input_type -> wharf.worker.v1.ArtifactEventRequest
	8,  // 8: wharf.worker.v1.Worker.StreamLogs:output_type -> wharf.worker.v1.StreamLogsResponse
	5,  // 9: wharf.worker.v1.Worker.Log:output_type -> wharf.worker.v1.LogResponse
	6,  // 10: wharf.worker.v1.Worker.StatusEvent:output_type -> wharf.worker.v1.StatusEventResponse
	7,  // 11: wharf.worker.v1.Worker.ArtifactEvent:output_type -> wharf.worker.v1.ArtifactEventResponse
	8,  // [8:12] is the sub-list for method output_type
	4,  // [4:8] is the sub-list for method input_type
	4,  // [4:4] is the sub-list for extension type_name
	4,  // [4:4] is the sub-list for extension extendee
	0,  // [0:4] is the sub-list for field type_name
}

func init() { file_api_workerapi_v1_worker_proto_init() }
func file_api_workerapi_v1_worker_proto_init() {
	if File_api_workerapi_v1_worker_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_workerapi_v1_worker_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StreamLogsRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_workerapi_v1_worker_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LogRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_workerapi_v1_worker_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StatusEventRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_workerapi_v1_worker_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ArtifactEventRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_workerapi_v1_worker_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LogResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_workerapi_v1_worker_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StatusEventResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_workerapi_v1_worker_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ArtifactEventResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_workerapi_v1_worker_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StreamLogsResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_workerapi_v1_worker_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LogLine); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_api_workerapi_v1_worker_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_api_workerapi_v1_worker_proto_goTypes,
		DependencyIndexes: file_api_workerapi_v1_worker_proto_depIdxs,
		EnumInfos:         file_api_workerapi_v1_worker_proto_enumTypes,
		MessageInfos:      file_api_workerapi_v1_worker_proto_msgTypes,
	}.Build()
	File_api_workerapi_v1_worker_proto = out.File
	file_api_workerapi_v1_worker_proto_rawDesc = nil
	file_api_workerapi_v1_worker_proto_goTypes = nil
	file_api_workerapi_v1_worker_proto_depIdxs = nil
}
