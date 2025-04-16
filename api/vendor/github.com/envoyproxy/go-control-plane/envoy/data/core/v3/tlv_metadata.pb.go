// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v5.29.3
// source: envoy/data/core/v3/tlv_metadata.proto

package corev3

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

type TlvsMetadata struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Typed metadata for :ref:`Proxy protocol filter <envoy_v3_api_msg_extensions.filters.listener.proxy_protocol.v3.ProxyProtocol>`, that represents a map of TLVs.
	// Each entry in the map consists of a key which corresponds to a configured
	// :ref:`rule key <envoy_v3_api_field_extensions.filters.listener.proxy_protocol.v3.ProxyProtocol.KeyValuePair.key>` and a value (TLV value in bytes).
	// When runtime flag “envoy.reloadable_features.use_typed_metadata_in_proxy_protocol_listener“ is enabled,
	// :ref:`Proxy protocol filter <envoy_v3_api_msg_extensions.filters.listener.proxy_protocol.v3.ProxyProtocol>`
	// will populate typed metadata and regular metadata. By default filter will populate typed and untyped metadata.
	TypedMetadata map[string][]byte `protobuf:"bytes,1,rep,name=typed_metadata,json=typedMetadata,proto3" json:"typed_metadata,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *TlvsMetadata) Reset() {
	*x = TlvsMetadata{}
	if protoimpl.UnsafeEnabled {
		mi := &file_envoy_data_core_v3_tlv_metadata_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TlvsMetadata) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TlvsMetadata) ProtoMessage() {}

func (x *TlvsMetadata) ProtoReflect() protoreflect.Message {
	mi := &file_envoy_data_core_v3_tlv_metadata_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TlvsMetadata.ProtoReflect.Descriptor instead.
func (*TlvsMetadata) Descriptor() ([]byte, []int) {
	return file_envoy_data_core_v3_tlv_metadata_proto_rawDescGZIP(), []int{0}
}

func (x *TlvsMetadata) GetTypedMetadata() map[string][]byte {
	if x != nil {
		return x.TypedMetadata
	}
	return nil
}

var File_envoy_data_core_v3_tlv_metadata_proto protoreflect.FileDescriptor

var file_envoy_data_core_v3_tlv_metadata_proto_rawDesc = []byte{
	0x0a, 0x25, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x2f, 0x64, 0x61, 0x74, 0x61, 0x2f, 0x63, 0x6f, 0x72,
	0x65, 0x2f, 0x76, 0x33, 0x2f, 0x74, 0x6c, 0x76, 0x5f, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74,
	0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x12, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x2e, 0x64,
	0x61, 0x74, 0x61, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x76, 0x33, 0x1a, 0x1d, 0x75, 0x64, 0x70,
	0x61, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x73, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xac, 0x01, 0x0a, 0x0c, 0x54,
	0x6c, 0x76, 0x73, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x12, 0x5a, 0x0a, 0x0e, 0x74,
	0x79, 0x70, 0x65, 0x64, 0x5f, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x33, 0x2e, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x2e, 0x64, 0x61, 0x74, 0x61,
	0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x76, 0x33, 0x2e, 0x54, 0x6c, 0x76, 0x73, 0x4d, 0x65, 0x74,
	0x61, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x54, 0x79, 0x70, 0x65, 0x64, 0x4d, 0x65, 0x74, 0x61, 0x64,
	0x61, 0x74, 0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x0d, 0x74, 0x79, 0x70, 0x65, 0x64, 0x4d,
	0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x1a, 0x40, 0x0a, 0x12, 0x54, 0x79, 0x70, 0x65, 0x64,
	0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a,
	0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12,
	0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x42, 0x80, 0x01, 0xba, 0x80, 0xc8, 0xd1,
	0x06, 0x02, 0x10, 0x02, 0x0a, 0x20, 0x69, 0x6f, 0x2e, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x70, 0x72,
	0x6f, 0x78, 0x79, 0x2e, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x2e, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x63,
	0x6f, 0x72, 0x65, 0x2e, 0x76, 0x33, 0x42, 0x10, 0x54, 0x6c, 0x76, 0x4d, 0x65, 0x74, 0x61, 0x64,
	0x61, 0x74, 0x61, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x40, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x70, 0x72, 0x6f, 0x78,
	0x79, 0x2f, 0x67, 0x6f, 0x2d, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x2d, 0x70, 0x6c, 0x61,
	0x6e, 0x65, 0x2f, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x2f, 0x64, 0x61, 0x74, 0x61, 0x2f, 0x63, 0x6f,
	0x72, 0x65, 0x2f, 0x76, 0x33, 0x3b, 0x63, 0x6f, 0x72, 0x65, 0x76, 0x33, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_envoy_data_core_v3_tlv_metadata_proto_rawDescOnce sync.Once
	file_envoy_data_core_v3_tlv_metadata_proto_rawDescData = file_envoy_data_core_v3_tlv_metadata_proto_rawDesc
)

func file_envoy_data_core_v3_tlv_metadata_proto_rawDescGZIP() []byte {
	file_envoy_data_core_v3_tlv_metadata_proto_rawDescOnce.Do(func() {
		file_envoy_data_core_v3_tlv_metadata_proto_rawDescData = protoimpl.X.CompressGZIP(file_envoy_data_core_v3_tlv_metadata_proto_rawDescData)
	})
	return file_envoy_data_core_v3_tlv_metadata_proto_rawDescData
}

var file_envoy_data_core_v3_tlv_metadata_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_envoy_data_core_v3_tlv_metadata_proto_goTypes = []interface{}{
	(*TlvsMetadata)(nil), // 0: envoy.data.core.v3.TlvsMetadata
	nil,                  // 1: envoy.data.core.v3.TlvsMetadata.TypedMetadataEntry
}
var file_envoy_data_core_v3_tlv_metadata_proto_depIdxs = []int32{
	1, // 0: envoy.data.core.v3.TlvsMetadata.typed_metadata:type_name -> envoy.data.core.v3.TlvsMetadata.TypedMetadataEntry
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_envoy_data_core_v3_tlv_metadata_proto_init() }
func file_envoy_data_core_v3_tlv_metadata_proto_init() {
	if File_envoy_data_core_v3_tlv_metadata_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_envoy_data_core_v3_tlv_metadata_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TlvsMetadata); i {
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
			RawDescriptor: file_envoy_data_core_v3_tlv_metadata_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_envoy_data_core_v3_tlv_metadata_proto_goTypes,
		DependencyIndexes: file_envoy_data_core_v3_tlv_metadata_proto_depIdxs,
		MessageInfos:      file_envoy_data_core_v3_tlv_metadata_proto_msgTypes,
	}.Build()
	File_envoy_data_core_v3_tlv_metadata_proto = out.File
	file_envoy_data_core_v3_tlv_metadata_proto_rawDesc = nil
	file_envoy_data_core_v3_tlv_metadata_proto_goTypes = nil
	file_envoy_data_core_v3_tlv_metadata_proto_depIdxs = nil
}
