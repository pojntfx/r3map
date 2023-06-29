// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v3.19.6
// source: backend.proto

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
		mi := &file_backend_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReadAtArgs) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReadAtArgs) ProtoMessage() {}

func (x *ReadAtArgs) ProtoReflect() protoreflect.Message {
	mi := &file_backend_proto_msgTypes[0]
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
	return file_backend_proto_rawDescGZIP(), []int{0}
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
		mi := &file_backend_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReadAtReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReadAtReply) ProtoMessage() {}

func (x *ReadAtReply) ProtoReflect() protoreflect.Message {
	mi := &file_backend_proto_msgTypes[1]
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
	return file_backend_proto_rawDescGZIP(), []int{1}
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

type WriteAtArgs struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Off int64  `protobuf:"varint,1,opt,name=Off,proto3" json:"Off,omitempty"`
	P   []byte `protobuf:"bytes,2,opt,name=P,proto3" json:"P,omitempty"`
}

func (x *WriteAtArgs) Reset() {
	*x = WriteAtArgs{}
	if protoimpl.UnsafeEnabled {
		mi := &file_backend_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WriteAtArgs) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WriteAtArgs) ProtoMessage() {}

func (x *WriteAtArgs) ProtoReflect() protoreflect.Message {
	mi := &file_backend_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WriteAtArgs.ProtoReflect.Descriptor instead.
func (*WriteAtArgs) Descriptor() ([]byte, []int) {
	return file_backend_proto_rawDescGZIP(), []int{2}
}

func (x *WriteAtArgs) GetOff() int64 {
	if x != nil {
		return x.Off
	}
	return 0
}

func (x *WriteAtArgs) GetP() []byte {
	if x != nil {
		return x.P
	}
	return nil
}

type WriteAtReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Length int32 `protobuf:"varint,1,opt,name=Length,proto3" json:"Length,omitempty"`
}

func (x *WriteAtReply) Reset() {
	*x = WriteAtReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_backend_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WriteAtReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WriteAtReply) ProtoMessage() {}

func (x *WriteAtReply) ProtoReflect() protoreflect.Message {
	mi := &file_backend_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WriteAtReply.ProtoReflect.Descriptor instead.
func (*WriteAtReply) Descriptor() ([]byte, []int) {
	return file_backend_proto_rawDescGZIP(), []int{3}
}

func (x *WriteAtReply) GetLength() int32 {
	if x != nil {
		return x.Length
	}
	return 0
}

type SizeArgs struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *SizeArgs) Reset() {
	*x = SizeArgs{}
	if protoimpl.UnsafeEnabled {
		mi := &file_backend_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SizeArgs) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SizeArgs) ProtoMessage() {}

func (x *SizeArgs) ProtoReflect() protoreflect.Message {
	mi := &file_backend_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SizeArgs.ProtoReflect.Descriptor instead.
func (*SizeArgs) Descriptor() ([]byte, []int) {
	return file_backend_proto_rawDescGZIP(), []int{4}
}

type SizeReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Size int64 `protobuf:"varint,1,opt,name=Size,proto3" json:"Size,omitempty"`
}

func (x *SizeReply) Reset() {
	*x = SizeReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_backend_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SizeReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SizeReply) ProtoMessage() {}

func (x *SizeReply) ProtoReflect() protoreflect.Message {
	mi := &file_backend_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SizeReply.ProtoReflect.Descriptor instead.
func (*SizeReply) Descriptor() ([]byte, []int) {
	return file_backend_proto_rawDescGZIP(), []int{5}
}

func (x *SizeReply) GetSize() int64 {
	if x != nil {
		return x.Size
	}
	return 0
}

type SyncArgs struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *SyncArgs) Reset() {
	*x = SyncArgs{}
	if protoimpl.UnsafeEnabled {
		mi := &file_backend_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SyncArgs) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SyncArgs) ProtoMessage() {}

func (x *SyncArgs) ProtoReflect() protoreflect.Message {
	mi := &file_backend_proto_msgTypes[6]
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
	return file_backend_proto_rawDescGZIP(), []int{6}
}

type SyncReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *SyncReply) Reset() {
	*x = SyncReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_backend_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SyncReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SyncReply) ProtoMessage() {}

func (x *SyncReply) ProtoReflect() protoreflect.Message {
	mi := &file_backend_proto_msgTypes[7]
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
	return file_backend_proto_rawDescGZIP(), []int{7}
}

var File_backend_proto protoreflect.FileDescriptor

var file_backend_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x26, 0x63, 0x6f, 0x6d, 0x2e, 0x70, 0x6f, 0x6a, 0x74, 0x69, 0x6e, 0x67, 0x65, 0x72, 0x2e, 0x66,
	0x65, 0x6c, 0x69, 0x63, 0x69, 0x74, 0x61, 0x73, 0x2e, 0x72, 0x33, 0x6d, 0x61, 0x70, 0x2e, 0x6d,
	0x6f, 0x75, 0x6e, 0x74, 0x2e, 0x76, 0x31, 0x22, 0x36, 0x0a, 0x0a, 0x52, 0x65, 0x61, 0x64, 0x41,
	0x74, 0x41, 0x72, 0x67, 0x73, 0x12, 0x16, 0x0a, 0x06, 0x4c, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x4c, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x12, 0x10, 0x0a,
	0x03, 0x4f, 0x66, 0x66, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x4f, 0x66, 0x66, 0x22,
	0x29, 0x0a, 0x0b, 0x52, 0x65, 0x61, 0x64, 0x41, 0x74, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x0c,
	0x0a, 0x01, 0x4e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x01, 0x4e, 0x12, 0x0c, 0x0a, 0x01,
	0x50, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x01, 0x50, 0x22, 0x2d, 0x0a, 0x0b, 0x57, 0x72,
	0x69, 0x74, 0x65, 0x41, 0x74, 0x41, 0x72, 0x67, 0x73, 0x12, 0x10, 0x0a, 0x03, 0x4f, 0x66, 0x66,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x4f, 0x66, 0x66, 0x12, 0x0c, 0x0a, 0x01, 0x50,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x01, 0x50, 0x22, 0x26, 0x0a, 0x0c, 0x57, 0x72, 0x69,
	0x74, 0x65, 0x41, 0x74, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x16, 0x0a, 0x06, 0x4c, 0x65, 0x6e,
	0x67, 0x74, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x4c, 0x65, 0x6e, 0x67, 0x74,
	0x68, 0x22, 0x0a, 0x0a, 0x08, 0x53, 0x69, 0x7a, 0x65, 0x41, 0x72, 0x67, 0x73, 0x22, 0x1f, 0x0a,
	0x09, 0x53, 0x69, 0x7a, 0x65, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x53, 0x69,
	0x7a, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x04, 0x53, 0x69, 0x7a, 0x65, 0x22, 0x0a,
	0x0a, 0x08, 0x53, 0x79, 0x6e, 0x63, 0x41, 0x72, 0x67, 0x73, 0x22, 0x0b, 0x0a, 0x09, 0x53, 0x79,
	0x6e, 0x63, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x32, 0xd4, 0x03, 0x0a, 0x07, 0x42, 0x61, 0x63, 0x6b,
	0x65, 0x6e, 0x64, 0x12, 0x73, 0x0a, 0x06, 0x52, 0x65, 0x61, 0x64, 0x41, 0x74, 0x12, 0x32, 0x2e,
	0x63, 0x6f, 0x6d, 0x2e, 0x70, 0x6f, 0x6a, 0x74, 0x69, 0x6e, 0x67, 0x65, 0x72, 0x2e, 0x66, 0x65,
	0x6c, 0x69, 0x63, 0x69, 0x74, 0x61, 0x73, 0x2e, 0x72, 0x33, 0x6d, 0x61, 0x70, 0x2e, 0x6d, 0x6f,
	0x75, 0x6e, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x61, 0x64, 0x41, 0x74, 0x41, 0x72, 0x67,
	0x73, 0x1a, 0x33, 0x2e, 0x63, 0x6f, 0x6d, 0x2e, 0x70, 0x6f, 0x6a, 0x74, 0x69, 0x6e, 0x67, 0x65,
	0x72, 0x2e, 0x66, 0x65, 0x6c, 0x69, 0x63, 0x69, 0x74, 0x61, 0x73, 0x2e, 0x72, 0x33, 0x6d, 0x61,
	0x70, 0x2e, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x61, 0x64, 0x41,
	0x74, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x12, 0x76, 0x0a, 0x07, 0x57, 0x72, 0x69, 0x74,
	0x65, 0x41, 0x74, 0x12, 0x33, 0x2e, 0x63, 0x6f, 0x6d, 0x2e, 0x70, 0x6f, 0x6a, 0x74, 0x69, 0x6e,
	0x67, 0x65, 0x72, 0x2e, 0x66, 0x65, 0x6c, 0x69, 0x63, 0x69, 0x74, 0x61, 0x73, 0x2e, 0x72, 0x33,
	0x6d, 0x61, 0x70, 0x2e, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x57, 0x72, 0x69,
	0x74, 0x65, 0x41, 0x74, 0x41, 0x72, 0x67, 0x73, 0x1a, 0x34, 0x2e, 0x63, 0x6f, 0x6d, 0x2e, 0x70,
	0x6f, 0x6a, 0x74, 0x69, 0x6e, 0x67, 0x65, 0x72, 0x2e, 0x66, 0x65, 0x6c, 0x69, 0x63, 0x69, 0x74,
	0x61, 0x73, 0x2e, 0x72, 0x33, 0x6d, 0x61, 0x70, 0x2e, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x2e, 0x76,
	0x31, 0x2e, 0x57, 0x72, 0x69, 0x74, 0x65, 0x41, 0x74, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00,
	0x12, 0x6d, 0x0a, 0x04, 0x53, 0x69, 0x7a, 0x65, 0x12, 0x30, 0x2e, 0x63, 0x6f, 0x6d, 0x2e, 0x70,
	0x6f, 0x6a, 0x74, 0x69, 0x6e, 0x67, 0x65, 0x72, 0x2e, 0x66, 0x65, 0x6c, 0x69, 0x63, 0x69, 0x74,
	0x61, 0x73, 0x2e, 0x72, 0x33, 0x6d, 0x61, 0x70, 0x2e, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x2e, 0x76,
	0x31, 0x2e, 0x53, 0x69, 0x7a, 0x65, 0x41, 0x72, 0x67, 0x73, 0x1a, 0x31, 0x2e, 0x63, 0x6f, 0x6d,
	0x2e, 0x70, 0x6f, 0x6a, 0x74, 0x69, 0x6e, 0x67, 0x65, 0x72, 0x2e, 0x66, 0x65, 0x6c, 0x69, 0x63,
	0x69, 0x74, 0x61, 0x73, 0x2e, 0x72, 0x33, 0x6d, 0x61, 0x70, 0x2e, 0x6d, 0x6f, 0x75, 0x6e, 0x74,
	0x2e, 0x76, 0x31, 0x2e, 0x53, 0x69, 0x7a, 0x65, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x12,
	0x6d, 0x0a, 0x04, 0x53, 0x79, 0x6e, 0x63, 0x12, 0x30, 0x2e, 0x63, 0x6f, 0x6d, 0x2e, 0x70, 0x6f,
	0x6a, 0x74, 0x69, 0x6e, 0x67, 0x65, 0x72, 0x2e, 0x66, 0x65, 0x6c, 0x69, 0x63, 0x69, 0x74, 0x61,
	0x73, 0x2e, 0x72, 0x33, 0x6d, 0x61, 0x70, 0x2e, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x2e, 0x76, 0x31,
	0x2e, 0x53, 0x79, 0x6e, 0x63, 0x41, 0x72, 0x67, 0x73, 0x1a, 0x31, 0x2e, 0x63, 0x6f, 0x6d, 0x2e,
	0x70, 0x6f, 0x6a, 0x74, 0x69, 0x6e, 0x67, 0x65, 0x72, 0x2e, 0x66, 0x65, 0x6c, 0x69, 0x63, 0x69,
	0x74, 0x61, 0x73, 0x2e, 0x72, 0x33, 0x6d, 0x61, 0x70, 0x2e, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x2e,
	0x76, 0x31, 0x2e, 0x53, 0x79, 0x6e, 0x63, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x42, 0x31,
	0x5a, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x6f, 0x6a,
	0x6e, 0x74, 0x66, 0x78, 0x2f, 0x72, 0x33, 0x6d, 0x61, 0x70, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x61,
	0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x2f, 0x76,
	0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_backend_proto_rawDescOnce sync.Once
	file_backend_proto_rawDescData = file_backend_proto_rawDesc
)

func file_backend_proto_rawDescGZIP() []byte {
	file_backend_proto_rawDescOnce.Do(func() {
		file_backend_proto_rawDescData = protoimpl.X.CompressGZIP(file_backend_proto_rawDescData)
	})
	return file_backend_proto_rawDescData
}

var file_backend_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_backend_proto_goTypes = []interface{}{
	(*ReadAtArgs)(nil),   // 0: com.pojtinger.felicitas.r3map.mount.v1.ReadAtArgs
	(*ReadAtReply)(nil),  // 1: com.pojtinger.felicitas.r3map.mount.v1.ReadAtReply
	(*WriteAtArgs)(nil),  // 2: com.pojtinger.felicitas.r3map.mount.v1.WriteAtArgs
	(*WriteAtReply)(nil), // 3: com.pojtinger.felicitas.r3map.mount.v1.WriteAtReply
	(*SizeArgs)(nil),     // 4: com.pojtinger.felicitas.r3map.mount.v1.SizeArgs
	(*SizeReply)(nil),    // 5: com.pojtinger.felicitas.r3map.mount.v1.SizeReply
	(*SyncArgs)(nil),     // 6: com.pojtinger.felicitas.r3map.mount.v1.SyncArgs
	(*SyncReply)(nil),    // 7: com.pojtinger.felicitas.r3map.mount.v1.SyncReply
}
var file_backend_proto_depIdxs = []int32{
	0, // 0: com.pojtinger.felicitas.r3map.mount.v1.Backend.ReadAt:input_type -> com.pojtinger.felicitas.r3map.mount.v1.ReadAtArgs
	2, // 1: com.pojtinger.felicitas.r3map.mount.v1.Backend.WriteAt:input_type -> com.pojtinger.felicitas.r3map.mount.v1.WriteAtArgs
	4, // 2: com.pojtinger.felicitas.r3map.mount.v1.Backend.Size:input_type -> com.pojtinger.felicitas.r3map.mount.v1.SizeArgs
	6, // 3: com.pojtinger.felicitas.r3map.mount.v1.Backend.Sync:input_type -> com.pojtinger.felicitas.r3map.mount.v1.SyncArgs
	1, // 4: com.pojtinger.felicitas.r3map.mount.v1.Backend.ReadAt:output_type -> com.pojtinger.felicitas.r3map.mount.v1.ReadAtReply
	3, // 5: com.pojtinger.felicitas.r3map.mount.v1.Backend.WriteAt:output_type -> com.pojtinger.felicitas.r3map.mount.v1.WriteAtReply
	5, // 6: com.pojtinger.felicitas.r3map.mount.v1.Backend.Size:output_type -> com.pojtinger.felicitas.r3map.mount.v1.SizeReply
	7, // 7: com.pojtinger.felicitas.r3map.mount.v1.Backend.Sync:output_type -> com.pojtinger.felicitas.r3map.mount.v1.SyncReply
	4, // [4:8] is the sub-list for method output_type
	0, // [0:4] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_backend_proto_init() }
func file_backend_proto_init() {
	if File_backend_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_backend_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
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
		file_backend_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
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
		file_backend_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WriteAtArgs); i {
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
		file_backend_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WriteAtReply); i {
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
		file_backend_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SizeArgs); i {
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
		file_backend_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SizeReply); i {
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
		file_backend_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
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
		file_backend_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
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
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_backend_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_backend_proto_goTypes,
		DependencyIndexes: file_backend_proto_depIdxs,
		MessageInfos:      file_backend_proto_msgTypes,
	}.Build()
	File_backend_proto = out.File
	file_backend_proto_rawDesc = nil
	file_backend_proto_goTypes = nil
	file_backend_proto_depIdxs = nil
}
