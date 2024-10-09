// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        (unknown)
// source: raft/v1/heartbeat.proto

package raftv1

import (
	_ "github.com/envoyproxy/protoc-gen-validate/validate"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// HeartbeatRequest represents a request to send a heartbeat signal.
type HeartbeatRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The ID of the sender node.
	// Changed from bool to string for proper validation.
	IsAlive bool `protobuf:"varint,1,opt,name=isAlive,proto3" json:"isAlive,omitempty"` // Removed validation for bool
	// The term of the sender node.
	Addr string   `protobuf:"bytes,2,opt,name=addr,proto3" json:"addr,omitempty"` // Unchanged
	Node []string `protobuf:"bytes,3,rep,name=node,proto3" json:"node,omitempty"`
}

func (x *HeartbeatRequest) Reset() {
	*x = HeartbeatRequest{}
	mi := &file_raft_v1_heartbeat_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HeartbeatRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HeartbeatRequest) ProtoMessage() {}

func (x *HeartbeatRequest) ProtoReflect() protoreflect.Message {
	mi := &file_raft_v1_heartbeat_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HeartbeatRequest.ProtoReflect.Descriptor instead.
func (*HeartbeatRequest) Descriptor() ([]byte, []int) {
	return file_raft_v1_heartbeat_proto_rawDescGZIP(), []int{0}
}

func (x *HeartbeatRequest) GetIsAlive() bool {
	if x != nil {
		return x.IsAlive
	}
	return false
}

func (x *HeartbeatRequest) GetAddr() string {
	if x != nil {
		return x.Addr
	}
	return ""
}

func (x *HeartbeatRequest) GetNode() []string {
	if x != nil {
		return x.Node
	}
	return nil
}

// HeartbeatResponse represents the response to a heartbeat request.
type HeartbeatResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The ID of the sender node.
	// Changed from bool to string for proper validation.
	IsAlive bool `protobuf:"varint,1,opt,name=isAlive,proto3" json:"isAlive,omitempty"` // Removed validation for bool
	// The term of the sender node.
	Addr string `protobuf:"bytes,2,opt,name=addr,proto3" json:"addr,omitempty"` // Unchanged
}

func (x *HeartbeatResponse) Reset() {
	*x = HeartbeatResponse{}
	mi := &file_raft_v1_heartbeat_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HeartbeatResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HeartbeatResponse) ProtoMessage() {}

func (x *HeartbeatResponse) ProtoReflect() protoreflect.Message {
	mi := &file_raft_v1_heartbeat_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HeartbeatResponse.ProtoReflect.Descriptor instead.
func (*HeartbeatResponse) Descriptor() ([]byte, []int) {
	return file_raft_v1_heartbeat_proto_rawDescGZIP(), []int{1}
}

func (x *HeartbeatResponse) GetIsAlive() bool {
	if x != nil {
		return x.IsAlive
	}
	return false
}

func (x *HeartbeatResponse) GetAddr() string {
	if x != nil {
		return x.Addr
	}
	return ""
}

var File_raft_v1_heartbeat_proto protoreflect.FileDescriptor

var file_raft_v1_heartbeat_proto_rawDesc = []byte{
	0x0a, 0x17, 0x72, 0x61, 0x66, 0x74, 0x2f, 0x76, 0x31, 0x2f, 0x68, 0x65, 0x61, 0x72, 0x74, 0x62,
	0x65, 0x61, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x72, 0x61, 0x66, 0x74, 0x2e,
	0x76, 0x31, 0x1a, 0x17, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76, 0x61, 0x6c,
	0x69, 0x64, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x69, 0x0a, 0x10, 0x48,
	0x65, 0x61, 0x72, 0x74, 0x62, 0x65, 0x61, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x18, 0x0a, 0x07, 0x69, 0x73, 0x41, 0x6c, 0x69, 0x76, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x07, 0x69, 0x73, 0x41, 0x6c, 0x69, 0x76, 0x65, 0x12, 0x1d, 0x0a, 0x04, 0x61, 0x64, 0x64,
	0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x09, 0xfa, 0x42, 0x06, 0x72, 0x04, 0x10, 0x01,
	0x18, 0x40, 0x52, 0x04, 0x61, 0x64, 0x64, 0x72, 0x12, 0x1c, 0x0a, 0x04, 0x6e, 0x6f, 0x64, 0x65,
	0x18, 0x03, 0x20, 0x03, 0x28, 0x09, 0x42, 0x08, 0xfa, 0x42, 0x05, 0x92, 0x01, 0x02, 0x08, 0x00,
	0x52, 0x04, 0x6e, 0x6f, 0x64, 0x65, 0x22, 0x4c, 0x0a, 0x11, 0x48, 0x65, 0x61, 0x72, 0x74, 0x62,
	0x65, 0x61, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x69,
	0x73, 0x41, 0x6c, 0x69, 0x76, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x69, 0x73,
	0x41, 0x6c, 0x69, 0x76, 0x65, 0x12, 0x1d, 0x0a, 0x04, 0x61, 0x64, 0x64, 0x72, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x42, 0x09, 0xfa, 0x42, 0x06, 0x72, 0x04, 0x10, 0x01, 0x18, 0x40, 0x52, 0x04,
	0x61, 0x64, 0x64, 0x72, 0x32, 0x5c, 0x0a, 0x10, 0x48, 0x65, 0x61, 0x72, 0x74, 0x62, 0x65, 0x61,
	0x74, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x48, 0x0a, 0x0d, 0x53, 0x65, 0x6e, 0x64,
	0x48, 0x65, 0x61, 0x72, 0x74, 0x62, 0x65, 0x61, 0x74, 0x12, 0x19, 0x2e, 0x72, 0x61, 0x66, 0x74,
	0x2e, 0x76, 0x31, 0x2e, 0x48, 0x65, 0x61, 0x72, 0x74, 0x62, 0x65, 0x61, 0x74, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x1a, 0x2e, 0x72, 0x61, 0x66, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x48,
	0x65, 0x61, 0x72, 0x74, 0x62, 0x65, 0x61, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x00, 0x42, 0x7c, 0x0a, 0x0b, 0x63, 0x6f, 0x6d, 0x2e, 0x72, 0x61, 0x66, 0x74, 0x2e, 0x76,
	0x31, 0x42, 0x0e, 0x48, 0x65, 0x61, 0x72, 0x74, 0x62, 0x65, 0x61, 0x74, 0x50, 0x72, 0x6f, 0x74,
	0x6f, 0x50, 0x01, 0x5a, 0x20, 0x72, 0x61, 0x66, 0x74, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e,
	0x61, 0x6c, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x72, 0x61, 0x66, 0x74, 0x2f, 0x76, 0x31, 0x3b, 0x72,
	0x61, 0x66, 0x74, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x52, 0x58, 0x58, 0xaa, 0x02, 0x07, 0x52, 0x61,
	0x66, 0x74, 0x2e, 0x56, 0x31, 0xca, 0x02, 0x07, 0x52, 0x61, 0x66, 0x74, 0x5c, 0x56, 0x31, 0xe2,
	0x02, 0x13, 0x52, 0x61, 0x66, 0x74, 0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74,
	0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x08, 0x52, 0x61, 0x66, 0x74, 0x3a, 0x3a, 0x56, 0x31,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_raft_v1_heartbeat_proto_rawDescOnce sync.Once
	file_raft_v1_heartbeat_proto_rawDescData = file_raft_v1_heartbeat_proto_rawDesc
)

func file_raft_v1_heartbeat_proto_rawDescGZIP() []byte {
	file_raft_v1_heartbeat_proto_rawDescOnce.Do(func() {
		file_raft_v1_heartbeat_proto_rawDescData = protoimpl.X.CompressGZIP(file_raft_v1_heartbeat_proto_rawDescData)
	})
	return file_raft_v1_heartbeat_proto_rawDescData
}

var file_raft_v1_heartbeat_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_raft_v1_heartbeat_proto_goTypes = []any{
	(*HeartbeatRequest)(nil),  // 0: raft.v1.HeartbeatRequest
	(*HeartbeatResponse)(nil), // 1: raft.v1.HeartbeatResponse
}
var file_raft_v1_heartbeat_proto_depIdxs = []int32{
	0, // 0: raft.v1.HeartbeatService.SendHeartbeat:input_type -> raft.v1.HeartbeatRequest
	1, // 1: raft.v1.HeartbeatService.SendHeartbeat:output_type -> raft.v1.HeartbeatResponse
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_raft_v1_heartbeat_proto_init() }
func file_raft_v1_heartbeat_proto_init() {
	if File_raft_v1_heartbeat_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_raft_v1_heartbeat_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_raft_v1_heartbeat_proto_goTypes,
		DependencyIndexes: file_raft_v1_heartbeat_proto_depIdxs,
		MessageInfos:      file_raft_v1_heartbeat_proto_msgTypes,
	}.Build()
	File_raft_v1_heartbeat_proto = out.File
	file_raft_v1_heartbeat_proto_rawDesc = nil
	file_raft_v1_heartbeat_proto_goTypes = nil
	file_raft_v1_heartbeat_proto_depIdxs = nil
}
