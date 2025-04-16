// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v5.29.3
// source: envoy/extensions/filters/network/generic_proxy/action/v3/action.proto

package actionv3

import (
	_ "github.com/cncf/xds/go/udpa/annotations"
	_ "github.com/cncf/xds/go/xds/annotations/v3"
	v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	v31 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	_ "github.com/envoyproxy/protoc-gen-validate/validate"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	anypb "google.golang.org/protobuf/types/known/anypb"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Configuration for the route match action.
// [#next-free-field: 8]
type RouteAction struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The name of the route action. This should be unique across all route actions.
	Name string `protobuf:"bytes,5,opt,name=name,proto3" json:"name,omitempty"`
	// Types that are assignable to ClusterSpecifier:
	//
	//	*RouteAction_Cluster
	//	*RouteAction_WeightedClusters
	ClusterSpecifier isRouteAction_ClusterSpecifier `protobuf_oneof:"cluster_specifier"`
	// Route metadata.
	Metadata *v3.Metadata `protobuf:"bytes,3,opt,name=metadata,proto3" json:"metadata,omitempty"`
	// Route level config for L7 generic filters. The key should be the related :ref:`extension name
	// <envoy_v3_api_field_config.core.v3.TypedExtensionConfig.name>` in the :ref:`generic filters
	// <envoy_v3_api_field_extensions.filters.network.generic_proxy.v3.GenericProxy.filters>`.
	PerFilterConfig map[string]*anypb.Any `protobuf:"bytes,4,rep,name=per_filter_config,json=perFilterConfig,proto3" json:"per_filter_config,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Specifies the upstream timeout for the route. If not specified, the default is 15s. This
	// spans between the point at which the entire downstream request (i.e. end-of-stream) has been
	// processed and when the upstream response has been completely processed. A value of 0 will
	// disable the route's timeout.
	Timeout *durationpb.Duration `protobuf:"bytes,6,opt,name=timeout,proto3" json:"timeout,omitempty"`
	// Specifies the retry policy for the route. If not specified, then no retries will be performed.
	//
	// .. note::
	//
	//	Only simplest retry policy is supported and only ``num_retries`` field is used for generic
	//	proxy. The default value for ``num_retries`` is 1 means that the request will be tried once
	//	and no additional retries will be performed.
	RetryPolicy *v3.RetryPolicy `protobuf:"bytes,7,opt,name=retry_policy,json=retryPolicy,proto3" json:"retry_policy,omitempty"`
}

func (x *RouteAction) Reset() {
	*x = RouteAction{}
	if protoimpl.UnsafeEnabled {
		mi := &file_envoy_extensions_filters_network_generic_proxy_action_v3_action_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RouteAction) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RouteAction) ProtoMessage() {}

func (x *RouteAction) ProtoReflect() protoreflect.Message {
	mi := &file_envoy_extensions_filters_network_generic_proxy_action_v3_action_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RouteAction.ProtoReflect.Descriptor instead.
func (*RouteAction) Descriptor() ([]byte, []int) {
	return file_envoy_extensions_filters_network_generic_proxy_action_v3_action_proto_rawDescGZIP(), []int{0}
}

func (x *RouteAction) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (m *RouteAction) GetClusterSpecifier() isRouteAction_ClusterSpecifier {
	if m != nil {
		return m.ClusterSpecifier
	}
	return nil
}

func (x *RouteAction) GetCluster() string {
	if x, ok := x.GetClusterSpecifier().(*RouteAction_Cluster); ok {
		return x.Cluster
	}
	return ""
}

func (x *RouteAction) GetWeightedClusters() *v31.WeightedCluster {
	if x, ok := x.GetClusterSpecifier().(*RouteAction_WeightedClusters); ok {
		return x.WeightedClusters
	}
	return nil
}

func (x *RouteAction) GetMetadata() *v3.Metadata {
	if x != nil {
		return x.Metadata
	}
	return nil
}

func (x *RouteAction) GetPerFilterConfig() map[string]*anypb.Any {
	if x != nil {
		return x.PerFilterConfig
	}
	return nil
}

func (x *RouteAction) GetTimeout() *durationpb.Duration {
	if x != nil {
		return x.Timeout
	}
	return nil
}

func (x *RouteAction) GetRetryPolicy() *v3.RetryPolicy {
	if x != nil {
		return x.RetryPolicy
	}
	return nil
}

type isRouteAction_ClusterSpecifier interface {
	isRouteAction_ClusterSpecifier()
}

type RouteAction_Cluster struct {
	// Indicates the upstream cluster to which the request should be routed.
	Cluster string `protobuf:"bytes,1,opt,name=cluster,proto3,oneof"`
}

type RouteAction_WeightedClusters struct {
	// [#not-implemented-hide:]
	// Multiple upstream clusters can be specified for a given route. The request is routed to one
	// of the upstream clusters based on weights assigned to each cluster.
	// Currently ClusterWeight only supports the name and weight fields.
	WeightedClusters *v31.WeightedCluster `protobuf:"bytes,2,opt,name=weighted_clusters,json=weightedClusters,proto3,oneof"`
}

func (*RouteAction_Cluster) isRouteAction_ClusterSpecifier() {}

func (*RouteAction_WeightedClusters) isRouteAction_ClusterSpecifier() {}

var File_envoy_extensions_filters_network_generic_proxy_action_v3_action_proto protoreflect.FileDescriptor

var file_envoy_extensions_filters_network_generic_proxy_action_v3_action_proto_rawDesc = []byte{
	0x0a, 0x45, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x2f, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f,
	0x6e, 0x73, 0x2f, 0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x73, 0x2f, 0x6e, 0x65, 0x74, 0x77, 0x6f,
	0x72, 0x6b, 0x2f, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x69, 0x63, 0x5f, 0x70, 0x72, 0x6f, 0x78, 0x79,
	0x2f, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x76, 0x33, 0x2f, 0x61, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x38, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x2e, 0x65,
	0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x66, 0x69, 0x6c, 0x74, 0x65, 0x72,
	0x73, 0x2e, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x2e, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x69,
	0x63, 0x5f, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x2e, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x76,
	0x33, 0x1a, 0x1f, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2f,
	0x63, 0x6f, 0x72, 0x65, 0x2f, 0x76, 0x33, 0x2f, 0x62, 0x61, 0x73, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x2c, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x2f, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x2f, 0x76, 0x33, 0x2f, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x5f,
	0x63, 0x6f, 0x6d, 0x70, 0x6f, 0x6e, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x19, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2f, 0x61, 0x6e, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x75, 0x72,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x78, 0x64, 0x73,
	0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x76, 0x33, 0x2f,
	0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1d, 0x75, 0x64,
	0x70, 0x61, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x73,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x17, 0x76, 0x61, 0x6c,
	0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0xc8, 0x04, 0x0a, 0x0b, 0x52, 0x6f, 0x75, 0x74, 0x65, 0x41, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x05, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1a, 0x0a, 0x07, 0x63, 0x6c, 0x75, 0x73,
	0x74, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x07, 0x63, 0x6c, 0x75,
	0x73, 0x74, 0x65, 0x72, 0x12, 0x55, 0x0a, 0x11, 0x77, 0x65, 0x69, 0x67, 0x68, 0x74, 0x65, 0x64,
	0x5f, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x26, 0x2e, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x72,
	0x6f, 0x75, 0x74, 0x65, 0x2e, 0x76, 0x33, 0x2e, 0x57, 0x65, 0x69, 0x67, 0x68, 0x74, 0x65, 0x64,
	0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x48, 0x00, 0x52, 0x10, 0x77, 0x65, 0x69, 0x67, 0x68,
	0x74, 0x65, 0x64, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x73, 0x12, 0x3a, 0x0a, 0x08, 0x6d,
	0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1e, 0x2e,
	0x65, 0x6e, 0x76, 0x6f, 0x79, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x63, 0x6f, 0x72,
	0x65, 0x2e, 0x76, 0x33, 0x2e, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x52, 0x08, 0x6d,
	0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x12, 0x86, 0x01, 0x0a, 0x11, 0x70, 0x65, 0x72, 0x5f,
	0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x04, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x5a, 0x2e, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x2e, 0x65, 0x78, 0x74, 0x65,
	0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x73, 0x2e, 0x6e,
	0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x2e, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x69, 0x63, 0x5f, 0x70,
	0x72, 0x6f, 0x78, 0x79, 0x2e, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x76, 0x33, 0x2e, 0x52,
	0x6f, 0x75, 0x74, 0x65, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x50, 0x65, 0x72, 0x46, 0x69,
	0x6c, 0x74, 0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52,
	0x0f, 0x70, 0x65, 0x72, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x12, 0x33, 0x0a, 0x07, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x19, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x07, 0x74, 0x69,
	0x6d, 0x65, 0x6f, 0x75, 0x74, 0x12, 0x44, 0x0a, 0x0c, 0x72, 0x65, 0x74, 0x72, 0x79, 0x5f, 0x70,
	0x6f, 0x6c, 0x69, 0x63, 0x79, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x65, 0x6e,
	0x76, 0x6f, 0x79, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e,
	0x76, 0x33, 0x2e, 0x52, 0x65, 0x74, 0x72, 0x79, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x52, 0x0b,
	0x72, 0x65, 0x74, 0x72, 0x79, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x1a, 0x58, 0x0a, 0x14, 0x50,
	0x65, 0x72, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x2a, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x3a, 0x02, 0x38, 0x01, 0x42, 0x18, 0x0a, 0x11, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72,
	0x5f, 0x73, 0x70, 0x65, 0x63, 0x69, 0x66, 0x69, 0x65, 0x72, 0x12, 0x03, 0xf8, 0x42, 0x01, 0x42,
	0xd1, 0x01, 0xba, 0x80, 0xc8, 0xd1, 0x06, 0x02, 0x10, 0x02, 0xd2, 0xc6, 0xa4, 0xe1, 0x06, 0x02,
	0x08, 0x01, 0x0a, 0x46, 0x69, 0x6f, 0x2e, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x70, 0x72, 0x6f, 0x78,
	0x79, 0x2e, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x2e, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f,
	0x6e, 0x73, 0x2e, 0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x73, 0x2e, 0x6e, 0x65, 0x74, 0x77, 0x6f,
	0x72, 0x6b, 0x2e, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x69, 0x63, 0x5f, 0x70, 0x72, 0x6f, 0x78, 0x79,
	0x2e, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x76, 0x33, 0x42, 0x0b, 0x41, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x68, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x70, 0x72, 0x6f, 0x78, 0x79,
	0x2f, 0x67, 0x6f, 0x2d, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x2d, 0x70, 0x6c, 0x61, 0x6e,
	0x65, 0x2f, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x2f, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f,
	0x6e, 0x73, 0x2f, 0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x73, 0x2f, 0x6e, 0x65, 0x74, 0x77, 0x6f,
	0x72, 0x6b, 0x2f, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x69, 0x63, 0x5f, 0x70, 0x72, 0x6f, 0x78, 0x79,
	0x2f, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x76, 0x33, 0x3b, 0x61, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x76, 0x33, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_envoy_extensions_filters_network_generic_proxy_action_v3_action_proto_rawDescOnce sync.Once
	file_envoy_extensions_filters_network_generic_proxy_action_v3_action_proto_rawDescData = file_envoy_extensions_filters_network_generic_proxy_action_v3_action_proto_rawDesc
)

func file_envoy_extensions_filters_network_generic_proxy_action_v3_action_proto_rawDescGZIP() []byte {
	file_envoy_extensions_filters_network_generic_proxy_action_v3_action_proto_rawDescOnce.Do(func() {
		file_envoy_extensions_filters_network_generic_proxy_action_v3_action_proto_rawDescData = protoimpl.X.CompressGZIP(file_envoy_extensions_filters_network_generic_proxy_action_v3_action_proto_rawDescData)
	})
	return file_envoy_extensions_filters_network_generic_proxy_action_v3_action_proto_rawDescData
}

var file_envoy_extensions_filters_network_generic_proxy_action_v3_action_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_envoy_extensions_filters_network_generic_proxy_action_v3_action_proto_goTypes = []interface{}{
	(*RouteAction)(nil),         // 0: envoy.extensions.filters.network.generic_proxy.action.v3.RouteAction
	nil,                         // 1: envoy.extensions.filters.network.generic_proxy.action.v3.RouteAction.PerFilterConfigEntry
	(*v31.WeightedCluster)(nil), // 2: envoy.config.route.v3.WeightedCluster
	(*v3.Metadata)(nil),         // 3: envoy.config.core.v3.Metadata
	(*durationpb.Duration)(nil), // 4: google.protobuf.Duration
	(*v3.RetryPolicy)(nil),      // 5: envoy.config.core.v3.RetryPolicy
	(*anypb.Any)(nil),           // 6: google.protobuf.Any
}
var file_envoy_extensions_filters_network_generic_proxy_action_v3_action_proto_depIdxs = []int32{
	2, // 0: envoy.extensions.filters.network.generic_proxy.action.v3.RouteAction.weighted_clusters:type_name -> envoy.config.route.v3.WeightedCluster
	3, // 1: envoy.extensions.filters.network.generic_proxy.action.v3.RouteAction.metadata:type_name -> envoy.config.core.v3.Metadata
	1, // 2: envoy.extensions.filters.network.generic_proxy.action.v3.RouteAction.per_filter_config:type_name -> envoy.extensions.filters.network.generic_proxy.action.v3.RouteAction.PerFilterConfigEntry
	4, // 3: envoy.extensions.filters.network.generic_proxy.action.v3.RouteAction.timeout:type_name -> google.protobuf.Duration
	5, // 4: envoy.extensions.filters.network.generic_proxy.action.v3.RouteAction.retry_policy:type_name -> envoy.config.core.v3.RetryPolicy
	6, // 5: envoy.extensions.filters.network.generic_proxy.action.v3.RouteAction.PerFilterConfigEntry.value:type_name -> google.protobuf.Any
	6, // [6:6] is the sub-list for method output_type
	6, // [6:6] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_envoy_extensions_filters_network_generic_proxy_action_v3_action_proto_init() }
func file_envoy_extensions_filters_network_generic_proxy_action_v3_action_proto_init() {
	if File_envoy_extensions_filters_network_generic_proxy_action_v3_action_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_envoy_extensions_filters_network_generic_proxy_action_v3_action_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RouteAction); i {
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
	file_envoy_extensions_filters_network_generic_proxy_action_v3_action_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*RouteAction_Cluster)(nil),
		(*RouteAction_WeightedClusters)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_envoy_extensions_filters_network_generic_proxy_action_v3_action_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_envoy_extensions_filters_network_generic_proxy_action_v3_action_proto_goTypes,
		DependencyIndexes: file_envoy_extensions_filters_network_generic_proxy_action_v3_action_proto_depIdxs,
		MessageInfos:      file_envoy_extensions_filters_network_generic_proxy_action_v3_action_proto_msgTypes,
	}.Build()
	File_envoy_extensions_filters_network_generic_proxy_action_v3_action_proto = out.File
	file_envoy_extensions_filters_network_generic_proxy_action_v3_action_proto_rawDesc = nil
	file_envoy_extensions_filters_network_generic_proxy_action_v3_action_proto_goTypes = nil
	file_envoy_extensions_filters_network_generic_proxy_action_v3_action_proto_depIdxs = nil
}
