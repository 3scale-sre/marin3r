// Copyright 2020 rvazquez@redhat.com
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package discoveryservice

import (
	"context"
	"time"

	"github.com/3scale-sre/marin3r/api/envoy"
	envoy_resources_v3 "github.com/3scale-sre/marin3r/api/envoy/resources/v3"
	envoy_serializer "github.com/3scale-sre/marin3r/api/envoy/serializer"
	"github.com/3scale-sre/marin3r/internal/pkg/discoveryservice/xdss/stats"
	"github.com/3scale-sre/marin3r/internal/pkg/util/backoff"
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_service_discovery_v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	server_v3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/go-logr/logr"
)

// Callbacks is a type that implements go-control-plane/pkg/server/Callbacks
type Callbacks struct {
	Stats  *stats.Stats
	Logger logr.Logger
}

var _ server_v3.Callbacks = &Callbacks{}

// OnStreamOpen implements go-control-plane/pkg/server/Callbacks.OnStreamOpen
// Returning an error will end processing and close the stream. OnStreamClosed will still be called.
func (cb *Callbacks) OnStreamOpen(ctx context.Context, id int64, typ string) error {
	cb.Logger.V(1).Info("Stream opened", "StreamId", id)

	return nil
}

// OnStreamClosed implements go-control-plane/pkg/server/Callbacks.OnStreamClosed
// OnStreamClosed is called immediately prior to closing an xDS stream with a stream ID.
func (cb *Callbacks) OnStreamClosed(id int64, node *envoy_config_core_v3.Node) {
	cb.Logger.V(1).Info("Stream closed", "StreamID", id)
}

// OnStreamRequest implements go-control-plane/pkg/server/Callbacks.OnStreamRequest
// OnStreamRequest is called once a request is received on a stream.
// Returning an error will end processing and close the stream. OnStreamClosed will still be called.
func (cb *Callbacks) OnStreamRequest(id int64, req *envoy_service_discovery_v3.DiscoveryRequest) error {
	// Try to get the Pod name associated with the request
	podName, err := stats.GetStringValueFromMetadata(req.GetNode().GetMetadata().AsMap(), "pod_name")
	if err != nil {
		cb.Logger.Error(err, "an error occurred, Pod name could not be retrieved", "NodeID", req.GetNode().GetId(), "StreamID", id)

		podName = "unknown"
	}

	log := cb.Logger.WithValues("TypeURL", req.GetTypeUrl(), "NodeID", req.GetNode().GetId(), "StreamID", id,
		"Pod", podName, "ResourceNames", req.GetResourceNames(), "LastAcceptedVersion", req.GetVersionInfo())

	if req.GetResponseNonce() != "" {
		if req.GetErrorDetail() != nil {
			log.Info("Discovery NACK")

			failures, err := cb.Stats.ReportNACK(req.GetNode().GetId(), req.GetTypeUrl(), podName, req.GetResponseNonce())
			if err != nil {
				log.Error(err, "error trying to report a response NACK")
			}

			// Backoff
			if failures == 0 {
				time.Sleep(100 * time.Millisecond)
			} else {
				time.Sleep(backoff.Default.Duration(int(failures)))
			}
		} else {
			log.Info("Discovery ACK")
			cb.Stats.ReportACK(req.GetNode().GetId(), req.GetTypeUrl(), req.GetVersionInfo(), podName)
		}
	} else {
		log.Info("Discovery Request")
		cb.Stats.ReportRequest(req.GetNode().GetId(), req.GetTypeUrl(), podName)
	}

	return nil
}

// OnStreamResponse implements go-control-plane/pkgserver/Callbacks.OnStreamResponse
// OnStreamResponse is called immediately prior to sending a response on a stream.
func (cb *Callbacks) OnStreamResponse(ctx context.Context, id int64, req *envoy_service_discovery_v3.DiscoveryRequest,
	rsp *envoy_service_discovery_v3.DiscoveryResponse) {
	log := cb.Logger.WithValues("TypeURL", req.GetTypeUrl(), "NodeID", req.GetNode().GetId(), "StreamID", id, "Version", rsp.GetVersionInfo())

	// Track the nonce of this response in the stats cache
	podName, err := stats.GetStringValueFromMetadata(req.GetNode().GetMetadata().AsMap(), "pod_name")
	if err != nil {
		log.Error(err, "an error occurred, nonce won't be tracked")
	} else {
		cb.Stats.WriteResponseNonce(req.GetNode().GetId(), rsp.GetTypeUrl(), rsp.GetVersionInfo(), podName, rsp.GetNonce())
	}

	// Log resources when in debug mode
	resources := []string{}

	for _, r := range rsp.GetResources() {
		j, _ := envoy_serializer.NewResourceMarshaller(envoy_serializer.JSON, envoy.APIv3).Marshal(r)
		resources = append(resources, j)
	}

	if rsp.GetTypeUrl() == envoy_resources_v3.Mappings()[envoy.Secret] {
		// Do not log secret contents
		log.V(1).Info("Discovery Response", "ResourcesNames", req.GetResourceNames(), "Pod", podName)
	} else {
		log.V(1).Info("Discovery Response", "Resources", resources, "Pod", podName)
	}
}

// OnFetchRequest implements go-control-plane/pkg/server/Callbacks.OnFetchRequest
// OnFetchRequest is called for each Fetch request. Returning an error will end processing of the
// request and respond with an error.
func (cb *Callbacks) OnFetchRequest(context.Context, *envoy_service_discovery_v3.DiscoveryRequest) error {
	return nil
}

// OnFetchResponse implements go-control-plane/pkg/server/Callbacks.OnFetchRequest
// OnFetchResponse is called immediately prior to sending a response.
func (cb *Callbacks) OnFetchResponse(*envoy_service_discovery_v3.DiscoveryRequest, *envoy_service_discovery_v3.DiscoveryResponse) {
}

// OnDeltaStreamOpen is called once an incremental xDS stream is open with a stream ID and the type URL (or "" for ADS).
// Returning an error will end processing and close the stream. OnStreamClosed will still be called.
func (cb *Callbacks) OnDeltaStreamOpen(context.Context, int64, string) error { return nil }

// OnDeltaStreamClosed is called immediately prior to closing an xDS stream with a stream ID.
func (cb *Callbacks) OnDeltaStreamClosed(int64, *envoy_config_core_v3.Node) {}

// OnStreamDeltaRequest is called once a request is received on a stream.
// Returning an error will end processing and close the stream. OnStreamClosed will still be called.
func (cb *Callbacks) OnStreamDeltaRequest(int64, *envoy_service_discovery_v3.DeltaDiscoveryRequest) error {
	return nil
}

// OnStreamDelatResponse is called immediately prior to sending a response on a stream.
func (cb *Callbacks) OnStreamDeltaResponse(int64, *envoy_service_discovery_v3.DeltaDiscoveryRequest, *envoy_service_discovery_v3.DeltaDiscoveryResponse) {
}
