// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v3.19.6
// source: seeder.proto

package v1

import (
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

type ReadAtArgs struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Length int32 `protobuf:"varint,1,opt,name=Length,proto3" json:"Length,omitempty"`
	Off    int64 `protobuf:"varint,2,opt,name=Off,proto3" json:"Off,omitempty"`
}

func (x *ReadAtArgs) Reset() {
	*x = ReadAtArgs{}
	if protoimpl.UnsafeEnabled {
		mi := &file_seeder_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReadAtArgs) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReadAtArgs) ProtoMessage() {}

func (x *ReadAtArgs) ProtoReflect() protoreflect.Message {
	mi := &file_seeder_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReadAtArgs.ProtoReflect.Descriptor instead.
func (*ReadAtArgs) Descriptor() ([]byte, []int) {
	return file_seeder_proto_rawDescGZIP(), []int{0}
}

func (x *ReadAtArgs) GetLength() int32 {
	if x != nil {
		return x.Length
	}
	return 0
}

func (x *ReadAtArgs) GetOff() int64 {
	if x != nil {
		return x.Off
	}
	return 0
}

type ReadAtReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	N int32  `protobuf:"varint,1,opt,name=N,proto3" json:"N,omitempty"`
	P []byte `protobuf:"bytes,2,opt,name=P,proto3" json:"P,omitempty"`
}

func (x *ReadAtReply) Reset() {
	*x = ReadAtReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_seeder_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReadAtReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReadAtReply) ProtoMessage() {}

func (x *ReadAtReply) ProtoReflect() protoreflect.Message {
	mi := &file_seeder_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReadAtReply.ProtoReflect.Descriptor instead.
func (*ReadAtReply) Descriptor() ([]byte, []int) {
	return file_seeder_proto_rawDescGZIP(), []int{1}
}

func (x *ReadAtReply) GetN() int32 {
	if x != nil {
		return x.N
	}
	return 0
}

func (x *ReadAtReply) GetP() []byte {
	if x != nil {
		return x.P
	}
	return nil
}

type TrackArgs struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *TrackArgs) Reset() {
	*x = TrackArgs{}
	if protoimpl.UnsafeEnabled {
		mi := &file_seeder_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TrackArgs) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TrackArgs) ProtoMessage() {}

func (x *TrackArgs) ProtoReflect() protoreflect.Message {
	mi := &file_seeder_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TrackArgs.ProtoReflect.Descriptor instead.
func (*TrackArgs) Descriptor() ([]byte, []int) {
	return file_seeder_proto_rawDescGZIP(), []int{2}
}

type TrackReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *TrackReply) Reset() {
	*x = TrackReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_seeder_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TrackReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TrackReply) ProtoMessage() {}

func (x *TrackReply) ProtoReflect() protoreflect.Message {
	mi := &file_seeder_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TrackReply.ProtoReflect.Descriptor instead.
func (*TrackReply) Descriptor() ([]byte, []int) {
	return file_seeder_proto_rawDescGZIP(), []int{3}
}

type SyncArgs struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *SyncArgs) Reset() {
	*x = SyncArgs{}
	if protoimpl.UnsafeEnabled {
		mi := &file_seeder_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SyncArgs) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SyncArgs) ProtoMessage() {}

func (x *SyncArgs) ProtoReflect() protoreflect.Message {
	mi := &file_seeder_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SyncArgs.ProtoReflect.Descriptor instead.
func (*SyncArgs) Descriptor() ([]byte, []int) {
	return file_seeder_proto_rawDescGZIP(), []int{4}
}

type SyncReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	DirtyOffsets []int64 `protobuf:"varint,1,rep,packed,name=DirtyOffsets,proto3" json:"DirtyOffsets,omitempty"`
}

func (x *SyncReply) Reset() {
	*x = SyncReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_seeder_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SyncReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SyncReply) ProtoMessage() {}

func (x *SyncReply) ProtoReflect() protoreflect.Message {
	mi := &file_seeder_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SyncReply.ProtoReflect.Descriptor instead.
func (*SyncReply) Descriptor() ([]byte, []int) {
	return file_seeder_proto_rawDescGZIP(), []int{5}
}

func (x *SyncReply) GetDirtyOffsets() []int64 {
	if x != nil {
		return x.DirtyOffsets
	}
	return nil
}

type CloseArgs struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *CloseArgs) Reset() {
	*x = CloseArgs{}
	if protoimpl.UnsafeEnabled {
		mi := &file_seeder_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CloseArgs) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CloseArgs) ProtoMessage() {}

func (x *CloseArgs) ProtoReflect() protoreflect.Message {
	mi := &file_seeder_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CloseArgs.ProtoReflect.Descriptor instead.
func (*CloseArgs) Descriptor() ([]byte, []int) {
	return file_seeder_proto_rawDescGZIP(), []int{6}
}

type CloseReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *CloseReply) Reset() {
	*x = CloseReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_seeder_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CloseReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CloseReply) ProtoMessage() {}

func (x *CloseReply) ProtoReflect() protoreflect.Message {
	mi := &file_seeder_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CloseReply.ProtoReflect.Descriptor instead.
func (*CloseReply) Descriptor() ([]byte, []int) {
	return file_seeder_proto_rawDescGZIP(), []int{7}
}

var File_seeder_proto protoreflect.FileDescriptor

var file_seeder_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x73, 0x65, 0x65, 0x64, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x2a,
	0x63, 0x6f, 0x6d, 0x2e, 0x70, 0x6f, 0x6a, 0x74, 0x69, 0x6e, 0x67, 0x65, 0x72, 0x2e, 0x66, 0x65,
	0x6c, 0x69, 0x63, 0x69, 0x74, 0x61, 0x73, 0x2e, 0x72, 0x33, 0x6d, 0x61, 0x70, 0x2e, 0x6d, 0x69,
	0x67, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x22, 0x36, 0x0a, 0x0a, 0x52, 0x65,
	0x61, 0x64, 0x41, 0x74, 0x41, 0x72, 0x67, 0x73, 0x12, 0x16, 0x0a, 0x06, 0x4c, 0x65, 0x6e, 0x67,
	0x74, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x4c, 0x65, 0x6e, 0x67, 0x74, 0x68,
	0x12, 0x10, 0x0a, 0x03, 0x4f, 0x66, 0x66, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x4f,
	0x66, 0x66, 0x22, 0x29, 0x0a, 0x0b, 0x52, 0x65, 0x61, 0x64, 0x41, 0x74, 0x52, 0x65, 0x70, 0x6c,
	0x79, 0x12, 0x0c, 0x0a, 0x01, 0x4e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x01, 0x4e, 0x12,
	0x0c, 0x0a, 0x01, 0x50, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x01, 0x50, 0x22, 0x0b, 0x0a,
	0x09, 0x54, 0x72, 0x61, 0x63, 0x6b, 0x41, 0x72, 0x67, 0x73, 0x22, 0x0c, 0x0a, 0x0a, 0x54, 0x72,
	0x61, 0x63, 0x6b, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x0a, 0x0a, 0x08, 0x53, 0x79, 0x6e, 0x63,
	0x41, 0x72, 0x67, 0x73, 0x22, 0x2f, 0x0a, 0x09, 0x53, 0x79, 0x6e, 0x63, 0x52, 0x65, 0x70, 0x6c,
	0x79, 0x12, 0x22, 0x0a, 0x0c, 0x44, 0x69, 0x72, 0x74, 0x79, 0x4f, 0x66, 0x66, 0x73, 0x65, 0x74,
	0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x03, 0x52, 0x0c, 0x44, 0x69, 0x72, 0x74, 0x79, 0x4f, 0x66,
	0x66, 0x73, 0x65, 0x74, 0x73, 0x22, 0x0b, 0x0a, 0x09, 0x43, 0x6c, 0x6f, 0x73, 0x65, 0x41, 0x72,
	0x67, 0x73, 0x22, 0x0c, 0x0a, 0x0a, 0x43, 0x6c, 0x6f, 0x73, 0x65, 0x52, 0x65, 0x70, 0x6c, 0x79,
	0x32, 0xf0, 0x03, 0x0a, 0x06, 0x53, 0x65, 0x65, 0x64, 0x65, 0x72, 0x12, 0x7b, 0x0a, 0x06, 0x52,
	0x65, 0x61, 0x64, 0x41, 0x74, 0x12, 0x36, 0x2e, 0x63, 0x6f, 0x6d, 0x2e, 0x70, 0x6f, 0x6a, 0x74,
	0x69, 0x6e, 0x67, 0x65, 0x72, 0x2e, 0x66, 0x65, 0x6c, 0x69, 0x63, 0x69, 0x74, 0x61, 0x73, 0x2e,
	0x72, 0x33, 0x6d, 0x61, 0x70, 0x2e, 0x6d, 0x69, 0x67, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e,
	0x76, 0x31, 0x2e, 0x52, 0x65, 0x61, 0x64, 0x41, 0x74, 0x41, 0x72, 0x67, 0x73, 0x1a, 0x37, 0x2e,
	0x63, 0x6f, 0x6d, 0x2e, 0x70, 0x6f, 0x6a, 0x74, 0x69, 0x6e, 0x67, 0x65, 0x72, 0x2e, 0x66, 0x65,
	0x6c, 0x69, 0x63, 0x69, 0x74, 0x61, 0x73, 0x2e, 0x72, 0x33, 0x6d, 0x61, 0x70, 0x2e, 0x6d, 0x69,
	0x67, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x61, 0x64, 0x41,
	0x74, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x12, 0x78, 0x0a, 0x05, 0x54, 0x72, 0x61, 0x63,
	0x6b, 0x12, 0x35, 0x2e, 0x63, 0x6f, 0x6d, 0x2e, 0x70, 0x6f, 0x6a, 0x74, 0x69, 0x6e, 0x67, 0x65,
	0x72, 0x2e, 0x66, 0x65, 0x6c, 0x69, 0x63, 0x69, 0x74, 0x61, 0x73, 0x2e, 0x72, 0x33, 0x6d, 0x61,
	0x70, 0x2e, 0x6d, 0x69, 0x67, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x54,
	0x72, 0x61, 0x63, 0x6b, 0x41, 0x72, 0x67, 0x73, 0x1a, 0x36, 0x2e, 0x63, 0x6f, 0x6d, 0x2e, 0x70,
	0x6f, 0x6a, 0x74, 0x69, 0x6e, 0x67, 0x65, 0x72, 0x2e, 0x66, 0x65, 0x6c, 0x69, 0x63, 0x69, 0x74,
	0x61, 0x73, 0x2e, 0x72, 0x33, 0x6d, 0x61, 0x70, 0x2e, 0x6d, 0x69, 0x67, 0x72, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x72, 0x61, 0x63, 0x6b, 0x52, 0x65, 0x70, 0x6c, 0x79,
	0x22, 0x00, 0x12, 0x75, 0x0a, 0x04, 0x53, 0x79, 0x6e, 0x63, 0x12, 0x34, 0x2e, 0x63, 0x6f, 0x6d,
	0x2e, 0x70, 0x6f, 0x6a, 0x74, 0x69, 0x6e, 0x67, 0x65, 0x72, 0x2e, 0x66, 0x65, 0x6c, 0x69, 0x63,
	0x69, 0x74, 0x61, 0x73, 0x2e, 0x72, 0x33, 0x6d, 0x61, 0x70, 0x2e, 0x6d, 0x69, 0x67, 0x72, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x79, 0x6e, 0x63, 0x41, 0x72, 0x67, 0x73,
	0x1a, 0x35, 0x2e, 0x63, 0x6f, 0x6d, 0x2e, 0x70, 0x6f, 0x6a, 0x74, 0x69, 0x6e, 0x67, 0x65, 0x72,
	0x2e, 0x66, 0x65, 0x6c, 0x69, 0x63, 0x69, 0x74, 0x61, 0x73, 0x2e, 0x72, 0x33, 0x6d, 0x61, 0x70,
	0x2e, 0x6d, 0x69, 0x67, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x79,
	0x6e, 0x63, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x12, 0x78, 0x0a, 0x05, 0x43, 0x6c, 0x6f,
	0x73, 0x65, 0x12, 0x35, 0x2e, 0x63, 0x6f, 0x6d, 0x2e, 0x70, 0x6f, 0x6a, 0x74, 0x69, 0x6e, 0x67,
	0x65, 0x72, 0x2e, 0x66, 0x65, 0x6c, 0x69, 0x63, 0x69, 0x74, 0x61, 0x73, 0x2e, 0x72, 0x33, 0x6d,
	0x61, 0x70, 0x2e, 0x6d, 0x69, 0x67, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e,
	0x43, 0x6c, 0x6f, 0x73, 0x65, 0x41, 0x72, 0x67, 0x73, 0x1a, 0x36, 0x2e, 0x63, 0x6f, 0x6d, 0x2e,
	0x70, 0x6f, 0x6a, 0x74, 0x69, 0x6e, 0x67, 0x65, 0x72, 0x2e, 0x66, 0x65, 0x6c, 0x69, 0x63, 0x69,
	0x74, 0x61, 0x73, 0x2e, 0x72, 0x33, 0x6d, 0x61, 0x70, 0x2e, 0x6d, 0x69, 0x67, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x6c, 0x6f, 0x73, 0x65, 0x52, 0x65, 0x70, 0x6c,
	0x79, 0x22, 0x00, 0x42, 0x35, 0x5a, 0x33, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x70, 0x6f, 0x6a, 0x6e, 0x74, 0x66, 0x78, 0x2f, 0x72, 0x33, 0x6d, 0x61, 0x70, 0x2f,
	0x70, 0x6b, 0x67, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x69,
	0x67, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_seeder_proto_rawDescOnce sync.Once
	file_seeder_proto_rawDescData = file_seeder_proto_rawDesc
)

func file_seeder_proto_rawDescGZIP() []byte {
	file_seeder_proto_rawDescOnce.Do(func() {
		file_seeder_proto_rawDescData = protoimpl.X.CompressGZIP(file_seeder_proto_rawDescData)
	})
	return file_seeder_proto_rawDescData
}

var file_seeder_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_seeder_proto_goTypes = []interface{}{
	(*ReadAtArgs)(nil),  // 0: com.pojtinger.felicitas.r3map.migration.v1.ReadAtArgs
	(*ReadAtReply)(nil), // 1: com.pojtinger.felicitas.r3map.migration.v1.ReadAtReply
	(*TrackArgs)(nil),   // 2: com.pojtinger.felicitas.r3map.migration.v1.TrackArgs
	(*TrackReply)(nil),  // 3: com.pojtinger.felicitas.r3map.migration.v1.TrackReply
	(*SyncArgs)(nil),    // 4: com.pojtinger.felicitas.r3map.migration.v1.SyncArgs
	(*SyncReply)(nil),   // 5: com.pojtinger.felicitas.r3map.migration.v1.SyncReply
	(*CloseArgs)(nil),   // 6: com.pojtinger.felicitas.r3map.migration.v1.CloseArgs
	(*CloseReply)(nil),  // 7: com.pojtinger.felicitas.r3map.migration.v1.CloseReply
}
var file_seeder_proto_depIdxs = []int32{
	0, // 0: com.pojtinger.felicitas.r3map.migration.v1.Seeder.ReadAt:input_type -> com.pojtinger.felicitas.r3map.migration.v1.ReadAtArgs
	2, // 1: com.pojtinger.felicitas.r3map.migration.v1.Seeder.Track:input_type -> com.pojtinger.felicitas.r3map.migration.v1.TrackArgs
	4, // 2: com.pojtinger.felicitas.r3map.migration.v1.Seeder.Sync:input_type -> com.pojtinger.felicitas.r3map.migration.v1.SyncArgs
	6, // 3: com.pojtinger.felicitas.r3map.migration.v1.Seeder.Close:input_type -> com.pojtinger.felicitas.r3map.migration.v1.CloseArgs
	1, // 4: com.pojtinger.felicitas.r3map.migration.v1.Seeder.ReadAt:output_type -> com.pojtinger.felicitas.r3map.migration.v1.ReadAtReply
	3, // 5: com.pojtinger.felicitas.r3map.migration.v1.Seeder.Track:output_type -> com.pojtinger.felicitas.r3map.migration.v1.TrackReply
	5, // 6: com.pojtinger.felicitas.r3map.migration.v1.Seeder.Sync:output_type -> com.pojtinger.felicitas.r3map.migration.v1.SyncReply
	7, // 7: com.pojtinger.felicitas.r3map.migration.v1.Seeder.Close:output_type -> com.pojtinger.felicitas.r3map.migration.v1.CloseReply
	4, // [4:8] is the sub-list for method output_type
	0, // [0:4] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_seeder_proto_init() }
func file_seeder_proto_init() {
	if File_seeder_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_seeder_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReadAtArgs); i {
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
		file_seeder_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReadAtReply); i {
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
		file_seeder_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TrackArgs); i {
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
		file_seeder_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TrackReply); i {
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
		file_seeder_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SyncArgs); i {
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
		file_seeder_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SyncReply); i {
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
		file_seeder_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CloseArgs); i {
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
		file_seeder_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CloseReply); i {
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
			RawDescriptor: file_seeder_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_seeder_proto_goTypes,
		DependencyIndexes: file_seeder_proto_depIdxs,
		MessageInfos:      file_seeder_proto_msgTypes,
	}.Build()
	File_seeder_proto = out.File
	file_seeder_proto_rawDesc = nil
	file_seeder_proto_goTypes = nil
	file_seeder_proto_depIdxs = nil
}
