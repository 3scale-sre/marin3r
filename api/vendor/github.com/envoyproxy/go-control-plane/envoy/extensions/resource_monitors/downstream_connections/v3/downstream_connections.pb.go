// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v5.29.3
// source: envoy/extensions/resource_monitors/downstream_connections/v3/downstream_connections.proto

package downstream_connectionsv3

import (
	_ "github.com/cncf/xds/go/udpa/annotations"
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

// The downstream connections resource monitor tracks the global number of open downstream connections.
type DownstreamConnectionsConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Maximum threshold for global open downstream connections, defaults to 0.
	// If monitor is enabled in Overload manager api, this field should be explicitly configured with value greater than 0.
	MaxActiveDownstreamConnections int64 `protobuf:"varint,1,opt,name=max_active_downstream_connections,json=maxActiveDownstreamConnections,proto3" json:"max_active_downstream_connections,omitempty"`
}

func (x *DownstreamConnectionsConfig) Reset() {
	*x = DownstreamConnectionsConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_envoy_extensions_resource_monitors_downstream_connections_v3_downstream_connections_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DownstreamConnectionsConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DownstreamConnectionsConfig) ProtoMessage() {}

func (x *DownstreamConnectionsConfig) ProtoReflect() protoreflect.Message {
	mi := &file_envoy_extensions_resource_monitors_downstream_connections_v3_downstream_connections_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DownstreamConnectionsConfig.ProtoReflect.Descriptor instead.
func (*DownstreamConnectionsConfig) Descriptor() ([]byte, []int) {
	return file_envoy_extensions_resource_monitors_downstream_connections_v3_downstream_connections_proto_rawDescGZIP(), []int{0}
}

func (x *DownstreamConnectionsConfig) GetMaxActiveDownstreamConnections() int64 {
	if x != nil {
		return x.MaxActiveDownstreamConnections
	}
	return 0
}

var File_envoy_extensions_resource_monitors_downstream_connections_v3_downstream_connections_proto protoreflect.FileDescriptor

var file_envoy_extensions_resource_monitors_downstream_connections_v3_downstream_connections_proto_rawDesc = []byte{
	0x0a, 0x59, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x2f, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f,
	0x6e, 0x73, 0x2f, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x6d, 0x6f, 0x6e, 0x69,
	0x74, 0x6f, 0x72, 0x73, 0x2f, 0x64, 0x6f, 0x77, 0x6e, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x5f,
	0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x76, 0x33, 0x2f, 0x64,
	0x6f, 0x77, 0x6e, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x5f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x3c, 0x65, 0x6e, 0x76,
	0x6f, 0x79, 0x2e, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x72, 0x65,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x73, 0x2e,
	0x64, 0x6f, 0x77, 0x6e, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x5f, 0x63, 0x6f, 0x6e, 0x6e, 0x65,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x76, 0x33, 0x1a, 0x1d, 0x75, 0x64, 0x70, 0x61, 0x2f,
	0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x73, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x17, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61,
	0x74, 0x65, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0x71, 0x0a, 0x1b, 0x44, 0x6f, 0x77, 0x6e, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x43,
	0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x12, 0x52, 0x0a, 0x21, 0x6d, 0x61, 0x78, 0x5f, 0x61, 0x63, 0x74, 0x69, 0x76, 0x65, 0x5f, 0x64,
	0x6f, 0x77, 0x6e, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x5f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x42, 0x07, 0xfa, 0x42, 0x04,
	0x22, 0x02, 0x20, 0x00, 0x52, 0x1e, 0x6d, 0x61, 0x78, 0x41, 0x63, 0x74, 0x69, 0x76, 0x65, 0x44,
	0x6f, 0x77, 0x6e, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x42, 0xf0, 0x01, 0xba, 0x80, 0xc8, 0xd1, 0x06, 0x02, 0x10, 0x02, 0x0a,
	0x4a, 0x69, 0x6f, 0x2e, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x2e, 0x65,
	0x6e, 0x76, 0x6f, 0x79, 0x2e, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x2e,
	0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72,
	0x73, 0x2e, 0x64, 0x6f, 0x77, 0x6e, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x5f, 0x63, 0x6f, 0x6e,
	0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x76, 0x33, 0x42, 0x1a, 0x44, 0x6f, 0x77,
	0x6e, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x7c, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x70, 0x72, 0x6f, 0x78, 0x79,
	0x2f, 0x67, 0x6f, 0x2d, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x2d, 0x70, 0x6c, 0x61, 0x6e,
	0x65, 0x2f, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x2f, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f,
	0x6e, 0x73, 0x2f, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x6d, 0x6f, 0x6e, 0x69,
	0x74, 0x6f, 0x72, 0x73, 0x2f, 0x64, 0x6f, 0x77, 0x6e, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x5f,
	0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x76, 0x33, 0x3b, 0x64,
	0x6f, 0x77, 0x6e, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x5f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x76, 0x33, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_envoy_extensions_resource_monitors_downstream_connections_v3_downstream_connections_proto_rawDescOnce sync.Once
	file_envoy_extensions_resource_monitors_downstream_connections_v3_downstream_connections_proto_rawDescData = file_envoy_extensions_resource_monitors_downstream_connections_v3_downstream_connections_proto_rawDesc
)

func file_envoy_extensions_resource_monitors_downstream_connections_v3_downstream_connections_proto_rawDescGZIP() []byte {
	file_envoy_extensions_resource_monitors_downstream_connections_v3_downstream_connections_proto_rawDescOnce.Do(func() {
		file_envoy_extensions_resource_monitors_downstream_connections_v3_downstream_connections_proto_rawDescData = protoimpl.X.CompressGZIP(file_envoy_extensions_resource_monitors_downstream_connections_v3_downstream_connections_proto_rawDescData)
	})
	return file_envoy_extensions_resource_monitors_downstream_connections_v3_downstream_connections_proto_rawDescData
}

var file_envoy_extensions_resource_monitors_downstream_connections_v3_downstream_connections_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_envoy_extensions_resource_monitors_downstream_connections_v3_downstream_connections_proto_goTypes = []interface{}{
	(*DownstreamConnectionsConfig)(nil), // 0: envoy.extensions.resource_monitors.downstream_connections.v3.DownstreamConnectionsConfig
}
var file_envoy_extensions_resource_monitors_downstream_connections_v3_downstream_connections_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() {
	file_envoy_extensions_resource_monitors_downstream_connections_v3_downstream_connections_proto_init()
}
func file_envoy_extensions_resource_monitors_downstream_connections_v3_downstream_connections_proto_init() {
	if File_envoy_extensions_resource_monitors_downstream_connections_v3_downstream_connections_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_envoy_extensions_resource_monitors_downstream_connections_v3_downstream_connections_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DownstreamConnectionsConfig); i {
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
			RawDescriptor: file_envoy_extensions_resource_monitors_downstream_connections_v3_downstream_connections_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_envoy_extensions_resource_monitors_downstream_connections_v3_downstream_connections_proto_goTypes,
		DependencyIndexes: file_envoy_extensions_resource_monitors_downstream_connections_v3_downstream_connections_proto_depIdxs,
		MessageInfos:      file_envoy_extensions_resource_monitors_downstream_connections_v3_downstream_connections_proto_msgTypes,
	}.Build()
	File_envoy_extensions_resource_monitors_downstream_connections_v3_downstream_connections_proto = out.File
	file_envoy_extensions_resource_monitors_downstream_connections_v3_downstream_connections_proto_rawDesc = nil
	file_envoy_extensions_resource_monitors_downstream_connections_v3_downstream_connections_proto_goTypes = nil
	file_envoy_extensions_resource_monitors_downstream_connections_v3_downstream_connections_proto_depIdxs = nil
}
