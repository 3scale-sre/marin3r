// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v5.29.3
// source: envoy/extensions/matching/common_inputs/ssl/v3/ssl_inputs.proto

package sslv3

import (
	_ "github.com/cncf/xds/go/udpa/annotations"
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

// List of comma-delimited URIs in the SAN field of the peer certificate for a downstream.
// [#extension: envoy.matching.inputs.uri_san]
type UriSanInput struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *UriSanInput) Reset() {
	*x = UriSanInput{}
	if protoimpl.UnsafeEnabled {
		mi := &file_envoy_extensions_matching_common_inputs_ssl_v3_ssl_inputs_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UriSanInput) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UriSanInput) ProtoMessage() {}

func (x *UriSanInput) ProtoReflect() protoreflect.Message {
	mi := &file_envoy_extensions_matching_common_inputs_ssl_v3_ssl_inputs_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UriSanInput.ProtoReflect.Descriptor instead.
func (*UriSanInput) Descriptor() ([]byte, []int) {
	return file_envoy_extensions_matching_common_inputs_ssl_v3_ssl_inputs_proto_rawDescGZIP(), []int{0}
}

// List of comma-delimited DNS entries in the SAN field of the peer certificate for a downstream.
// [#extension: envoy.matching.inputs.dns_san]
type DnsSanInput struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *DnsSanInput) Reset() {
	*x = DnsSanInput{}
	if protoimpl.UnsafeEnabled {
		mi := &file_envoy_extensions_matching_common_inputs_ssl_v3_ssl_inputs_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DnsSanInput) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DnsSanInput) ProtoMessage() {}

func (x *DnsSanInput) ProtoReflect() protoreflect.Message {
	mi := &file_envoy_extensions_matching_common_inputs_ssl_v3_ssl_inputs_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DnsSanInput.ProtoReflect.Descriptor instead.
func (*DnsSanInput) Descriptor() ([]byte, []int) {
	return file_envoy_extensions_matching_common_inputs_ssl_v3_ssl_inputs_proto_rawDescGZIP(), []int{1}
}

// Input that matches the subject field of the peer certificate in RFC 2253 format for a
// downstream.
// [#extension: envoy.matching.inputs.subject]
type SubjectInput struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *SubjectInput) Reset() {
	*x = SubjectInput{}
	if protoimpl.UnsafeEnabled {
		mi := &file_envoy_extensions_matching_common_inputs_ssl_v3_ssl_inputs_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SubjectInput) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubjectInput) ProtoMessage() {}

func (x *SubjectInput) ProtoReflect() protoreflect.Message {
	mi := &file_envoy_extensions_matching_common_inputs_ssl_v3_ssl_inputs_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SubjectInput.ProtoReflect.Descriptor instead.
func (*SubjectInput) Descriptor() ([]byte, []int) {
	return file_envoy_extensions_matching_common_inputs_ssl_v3_ssl_inputs_proto_rawDescGZIP(), []int{2}
}

var File_envoy_extensions_matching_common_inputs_ssl_v3_ssl_inputs_proto protoreflect.FileDescriptor

var file_envoy_extensions_matching_common_inputs_ssl_v3_ssl_inputs_proto_rawDesc = []byte{
	0x0a, 0x3f, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x2f, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f,
	0x6e, 0x73, 0x2f, 0x6d, 0x61, 0x74, 0x63, 0x68, 0x69, 0x6e, 0x67, 0x2f, 0x63, 0x6f, 0x6d, 0x6d,
	0x6f, 0x6e, 0x5f, 0x69, 0x6e, 0x70, 0x75, 0x74, 0x73, 0x2f, 0x73, 0x73, 0x6c, 0x2f, 0x76, 0x33,
	0x2f, 0x73, 0x73, 0x6c, 0x5f, 0x69, 0x6e, 0x70, 0x75, 0x74, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x2e, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x2e, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69,
	0x6f, 0x6e, 0x73, 0x2e, 0x6d, 0x61, 0x74, 0x63, 0x68, 0x69, 0x6e, 0x67, 0x2e, 0x63, 0x6f, 0x6d,
	0x6d, 0x6f, 0x6e, 0x5f, 0x69, 0x6e, 0x70, 0x75, 0x74, 0x73, 0x2e, 0x73, 0x73, 0x6c, 0x2e, 0x76,
	0x33, 0x1a, 0x1d, 0x75, 0x64, 0x70, 0x61, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x2f, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0x0d, 0x0a, 0x0b, 0x55, 0x72, 0x69, 0x53, 0x61, 0x6e, 0x49, 0x6e, 0x70, 0x75, 0x74, 0x22,
	0x0d, 0x0a, 0x0b, 0x44, 0x6e, 0x73, 0x53, 0x61, 0x6e, 0x49, 0x6e, 0x70, 0x75, 0x74, 0x22, 0x0e,
	0x0a, 0x0c, 0x53, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x6e, 0x70, 0x75, 0x74, 0x42, 0xb5,
	0x01, 0xba, 0x80, 0xc8, 0xd1, 0x06, 0x02, 0x10, 0x02, 0x0a, 0x3c, 0x69, 0x6f, 0x2e, 0x65, 0x6e,
	0x76, 0x6f, 0x79, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x2e, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x2e, 0x65,
	0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x6d, 0x61, 0x74, 0x63, 0x68, 0x69,
	0x6e, 0x67, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x5f, 0x69, 0x6e, 0x70, 0x75, 0x74, 0x73,
	0x2e, 0x73, 0x73, 0x6c, 0x2e, 0x76, 0x33, 0x42, 0x0e, 0x53, 0x73, 0x6c, 0x49, 0x6e, 0x70, 0x75,
	0x74, 0x73, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x5b, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x70, 0x72, 0x6f, 0x78, 0x79,
	0x2f, 0x67, 0x6f, 0x2d, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x2d, 0x70, 0x6c, 0x61, 0x6e,
	0x65, 0x2f, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x2f, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f,
	0x6e, 0x73, 0x2f, 0x6d, 0x61, 0x74, 0x63, 0x68, 0x69, 0x6e, 0x67, 0x2f, 0x63, 0x6f, 0x6d, 0x6d,
	0x6f, 0x6e, 0x5f, 0x69, 0x6e, 0x70, 0x75, 0x74, 0x73, 0x2f, 0x73, 0x73, 0x6c, 0x2f, 0x76, 0x33,
	0x3b, 0x73, 0x73, 0x6c, 0x76, 0x33, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_envoy_extensions_matching_common_inputs_ssl_v3_ssl_inputs_proto_rawDescOnce sync.Once
	file_envoy_extensions_matching_common_inputs_ssl_v3_ssl_inputs_proto_rawDescData = file_envoy_extensions_matching_common_inputs_ssl_v3_ssl_inputs_proto_rawDesc
)

func file_envoy_extensions_matching_common_inputs_ssl_v3_ssl_inputs_proto_rawDescGZIP() []byte {
	file_envoy_extensions_matching_common_inputs_ssl_v3_ssl_inputs_proto_rawDescOnce.Do(func() {
		file_envoy_extensions_matching_common_inputs_ssl_v3_ssl_inputs_proto_rawDescData = protoimpl.X.CompressGZIP(file_envoy_extensions_matching_common_inputs_ssl_v3_ssl_inputs_proto_rawDescData)
	})
	return file_envoy_extensions_matching_common_inputs_ssl_v3_ssl_inputs_proto_rawDescData
}

var file_envoy_extensions_matching_common_inputs_ssl_v3_ssl_inputs_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_envoy_extensions_matching_common_inputs_ssl_v3_ssl_inputs_proto_goTypes = []interface{}{
	(*UriSanInput)(nil),  // 0: envoy.extensions.matching.common_inputs.ssl.v3.UriSanInput
	(*DnsSanInput)(nil),  // 1: envoy.extensions.matching.common_inputs.ssl.v3.DnsSanInput
	(*SubjectInput)(nil), // 2: envoy.extensions.matching.common_inputs.ssl.v3.SubjectInput
}
var file_envoy_extensions_matching_common_inputs_ssl_v3_ssl_inputs_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_envoy_extensions_matching_common_inputs_ssl_v3_ssl_inputs_proto_init() }
func file_envoy_extensions_matching_common_inputs_ssl_v3_ssl_inputs_proto_init() {
	if File_envoy_extensions_matching_common_inputs_ssl_v3_ssl_inputs_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_envoy_extensions_matching_common_inputs_ssl_v3_ssl_inputs_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UriSanInput); i {
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
		file_envoy_extensions_matching_common_inputs_ssl_v3_ssl_inputs_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DnsSanInput); i {
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
		file_envoy_extensions_matching_common_inputs_ssl_v3_ssl_inputs_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SubjectInput); i {
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
			RawDescriptor: file_envoy_extensions_matching_common_inputs_ssl_v3_ssl_inputs_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_envoy_extensions_matching_common_inputs_ssl_v3_ssl_inputs_proto_goTypes,
		DependencyIndexes: file_envoy_extensions_matching_common_inputs_ssl_v3_ssl_inputs_proto_depIdxs,
		MessageInfos:      file_envoy_extensions_matching_common_inputs_ssl_v3_ssl_inputs_proto_msgTypes,
	}.Build()
	File_envoy_extensions_matching_common_inputs_ssl_v3_ssl_inputs_proto = out.File
	file_envoy_extensions_matching_common_inputs_ssl_v3_ssl_inputs_proto_rawDesc = nil
	file_envoy_extensions_matching_common_inputs_ssl_v3_ssl_inputs_proto_goTypes = nil
	file_envoy_extensions_matching_common_inputs_ssl_v3_ssl_inputs_proto_depIdxs = nil
}
