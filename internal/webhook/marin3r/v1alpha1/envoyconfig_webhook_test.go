/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"testing"

	"github.com/3scale-sre/marin3r/api/envoy"
	envoy_serializer "github.com/3scale-sre/marin3r/api/envoy/serializer"
	marin3rv1alpha1 "github.com/3scale-sre/marin3r/api/marin3r/v1alpha1"
	"github.com/3scale-sre/marin3r/internal/pkg/util/pointer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestEnvoyConfig_ValidateResources(t *testing.T) {
	tests := []struct {
		name    string
		r       *marin3rv1alpha1.EnvoyConfig
		wantErr bool
	}{
		{
			name: "Succeeds: type cluster",
			r: &marin3rv1alpha1.EnvoyConfig{
				Spec: marin3rv1alpha1.EnvoyConfigSpec{
					NodeID: "test",
					Resources: []marin3rv1alpha1.Resource{{
						Type: "cluster",
						Value: &runtime.RawExtension{
							Raw: []byte(`{"name": "cluster"}`),
						},
					}},
				},
			},
			wantErr: false,
		},
		{
			name: "Fails: incorrect timeout",
			r: &marin3rv1alpha1.EnvoyConfig{
				Spec: marin3rv1alpha1.EnvoyConfigSpec{
					NodeID: "test",
					Resources: []marin3rv1alpha1.Resource{{
						Type: "cluster",
						Value: &runtime.RawExtension{
							Raw: []byte(`{"name":"cluster1","type":"STRICT_DNS","connect_timeout":"xx"}`),
						},
					}},
				},
			}, wantErr: true,
		},
		{
			name: "Fails: missing resource value",
			r: &marin3rv1alpha1.EnvoyConfig{
				Spec: marin3rv1alpha1.EnvoyConfigSpec{
					NodeID: "test",
					Resources: []marin3rv1alpha1.Resource{{
						Type: "cluster",
					}},
				},
			}, wantErr: true,
		},
		{
			name: "Fails: blueprint cannot be used for cluster",
			r: &marin3rv1alpha1.EnvoyConfig{
				Spec: marin3rv1alpha1.EnvoyConfigSpec{
					NodeID: "test",
					Resources: []marin3rv1alpha1.Resource{{
						Type: "cluster",
						Value: &runtime.RawExtension{
							Raw: []byte(`{"name": "cluster"}`),
						},
						Blueprint: new(marin3rv1alpha1.Blueprint),
					}},
				},
			}, wantErr: true,
		},
		{
			name: "Fails: generateFromEndpointSlice cannot be used for cluster",
			r: &marin3rv1alpha1.EnvoyConfig{
				Spec: marin3rv1alpha1.EnvoyConfigSpec{
					NodeID: "test",
					Resources: []marin3rv1alpha1.Resource{{
						Type: "cluster",
						Value: &runtime.RawExtension{
							Raw: []byte(`{"name": "cluster"}`),
						},
						GenerateFromEndpointSlices: &marin3rv1alpha1.GenerateFromEndpointSlices{},
					}},
				},
			}, wantErr: true,
		},
		{
			name: "Fails: generateFromTlsSecret cannot be used for cluster",
			r: &marin3rv1alpha1.EnvoyConfig{
				Spec: marin3rv1alpha1.EnvoyConfigSpec{
					NodeID: "test",
					Resources: []marin3rv1alpha1.Resource{{
						Type: "cluster",
						Value: &runtime.RawExtension{
							Raw: []byte(`{"name": "cluster"}`),
						},
						GenerateFromTlsSecret: new(string),
					}},
				},
			}, wantErr: true,
		},
		{
			name: "Succeeds: type secret",
			r: &marin3rv1alpha1.EnvoyConfig{
				Spec: marin3rv1alpha1.EnvoyConfigSpec{
					NodeID: "test",
					Resources: []marin3rv1alpha1.Resource{{
						Type:                  "secret",
						GenerateFromTlsSecret: new(string),
					}},
				},
			}, wantErr: false,
		},
		{
			name: "Fails: generateFromTlsSecret' cannot be empty for secret",
			r: &marin3rv1alpha1.EnvoyConfig{
				Spec: marin3rv1alpha1.EnvoyConfigSpec{
					NodeID: "test",
					Resources: []marin3rv1alpha1.Resource{{
						Type: "secret",
					}},
				},
			}, wantErr: true,
		},
		{
			name: "Fails: value cannot be used for secret",
			r: &marin3rv1alpha1.EnvoyConfig{
				Spec: marin3rv1alpha1.EnvoyConfigSpec{
					NodeID: "test",
					Resources: []marin3rv1alpha1.Resource{{
						Type:  "secret",
						Value: &runtime.RawExtension{},
					}},
				},
			}, wantErr: true,
		},
		{
			name: "Fails: generateFromEndpointSlice can only be used for endpoints",
			r: &marin3rv1alpha1.EnvoyConfig{
				Spec: marin3rv1alpha1.EnvoyConfigSpec{
					NodeID: "test",
					Resources: []marin3rv1alpha1.Resource{{
						Type:                       "secret",
						GenerateFromEndpointSlices: &marin3rv1alpha1.GenerateFromEndpointSlices{},
					}},
				},
			}, wantErr: true,
		},
		{
			name: "Succeeds: type endpoint",
			r: &marin3rv1alpha1.EnvoyConfig{
				Spec: marin3rv1alpha1.EnvoyConfigSpec{
					NodeID: "test",
					Resources: []marin3rv1alpha1.Resource{{
						Type: "endpoint",
						GenerateFromEndpointSlices: &marin3rv1alpha1.GenerateFromEndpointSlices{
							Selector:    &metav1.LabelSelector{MatchLabels: map[string]string{"label": "value"}},
							ClusterName: "test",
							TargetPort:  "port",
						},
					}},
				},
			}, wantErr: false,
		},
		{
			name: "Fails: one of value/generateFromEndpointSlice for endpoint",
			r: &marin3rv1alpha1.EnvoyConfig{
				Spec: marin3rv1alpha1.EnvoyConfigSpec{
					NodeID: "test",
					Resources: []marin3rv1alpha1.Resource{{
						Type: "endpoint",
						GenerateFromEndpointSlices: &marin3rv1alpha1.GenerateFromEndpointSlices{
							Selector:    &metav1.LabelSelector{MatchLabels: map[string]string{"label": "value"}},
							ClusterName: "test",
							TargetPort:  "port",
						}, Value: &runtime.RawExtension{},
					}},
				},
			}, wantErr: true,
		},
		{
			name: "Fails: missing value/generateFromEndpointSlice for endpoint",
			r: &marin3rv1alpha1.EnvoyConfig{
				Spec: marin3rv1alpha1.EnvoyConfigSpec{
					NodeID: "test",
					Resources: []marin3rv1alpha1.Resource{{
						Type: "endpoint",
					}},
				},
			}, wantErr: true,
		},
		{
			name: "Fails: generateFromTlsSecret not allowed for endpoint",
			r: &marin3rv1alpha1.EnvoyConfig{
				Spec: marin3rv1alpha1.EnvoyConfigSpec{
					NodeID: "test",
					Resources: []marin3rv1alpha1.Resource{{
						Type:                  "endpoint",
						GenerateFromTlsSecret: new(string),
					}},
				},
			}, wantErr: true,
		},
		{
			name: "Fails: blueprint not allowed for endpoint",
			r: &marin3rv1alpha1.EnvoyConfig{
				Spec: marin3rv1alpha1.EnvoyConfigSpec{
					NodeID: "test",
					Resources: []marin3rv1alpha1.Resource{{
						Type:      "endpoint",
						Blueprint: new(marin3rv1alpha1.Blueprint),
					}},
				},
			}, wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateResources(tt.r); (err != nil) != tt.wantErr {
				t.Errorf("EnvoyConfig.ValidateResources() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validateEnvoyResources(t *testing.T) {
	tests := []struct {
		name    string
		r       *marin3rv1alpha1.EnvoyConfig
		wantErr bool
	}{
		{
			name: "fails for an EnvoyConfig with a syntax error in one of the envoy resources (from json)",
			r: &marin3rv1alpha1.EnvoyConfig{
				Spec: marin3rv1alpha1.EnvoyConfigSpec{
					NodeID:        "test",
					Serialization: pointer.New(envoy_serializer.JSON),
					EnvoyAPI:      pointer.New(envoy.APIv3),
					EnvoyResources: &marin3rv1alpha1.EnvoyResources{
						Clusters: []marin3rv1alpha1.EnvoyResource{{
							Name: pointer.New("cluster"),
							// the connect_timeout value unit is wrong
							Value: `{"name":"cluster1","type":"STRICT_DNS","connect_timeout":"2xs","load_assignment":{"cluster_name":"cluster1"}}`,
						}},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "fails for an EnvoyConfig with a syntax error in one of the envoy resources (from yaml)",
			r: &marin3rv1alpha1.EnvoyConfig{
				Spec: marin3rv1alpha1.EnvoyConfigSpec{
					NodeID:        "test",
					Serialization: pointer.New(envoy_serializer.YAML),
					EnvoyAPI:      pointer.New(envoy.APIv3),
					EnvoyResources: &marin3rv1alpha1.EnvoyResources{
						Listeners: []marin3rv1alpha1.EnvoyResource{{
							Name: pointer.New("test"),
							// the "port" property should be "port_value"
							Value: `
                              name: listener1
                              address:
                                socket_address:
                                  address: 0.0.0.0
                                  port: 8443
                            `,
						}},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateEnvoyResources(tt.r); (err != nil) != tt.wantErr {
				t.Errorf("EnvoyConfig.ValidateEnvoyResources() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnvoyConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		r       *marin3rv1alpha1.EnvoyConfig
		wantErr bool
	}{
		{
			name: "Ok, using spec.EnvoyResources",
			r: &marin3rv1alpha1.EnvoyConfig{
				Spec: marin3rv1alpha1.EnvoyConfigSpec{
					NodeID: "test",
					EnvoyResources: &marin3rv1alpha1.EnvoyResources{
						Clusters: []marin3rv1alpha1.EnvoyResource{{
							Name:  pointer.New("cluster"),
							Value: `{"name":"cluster1","type":"STRICT_DNS","connect_timeout":"2s","load_assignment":{"cluster_name":"cluster1"}}`,
						}},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Ok, using spec.Resources",
			r: &marin3rv1alpha1.EnvoyConfig{
				Spec: marin3rv1alpha1.EnvoyConfigSpec{
					NodeID: "test",
					Resources: []marin3rv1alpha1.Resource{{
						Type: "cluster",
						Value: &runtime.RawExtension{
							Raw: []byte(`{"name":"cluster1","type":"STRICT_DNS","connect_timeout":"2s","load_assignment":{"cluster_name":"cluster1"}}`),
						},
					}},
				},
			},
			wantErr: false,
		},
		{
			name: "Fail, cannot use EnvoyResources and Resources both",
			r: &marin3rv1alpha1.EnvoyConfig{
				Spec: marin3rv1alpha1.EnvoyConfigSpec{
					NodeID:         "test",
					Resources:      []marin3rv1alpha1.Resource{},
					EnvoyResources: &marin3rv1alpha1.EnvoyResources{},
				},
			},
			wantErr: true,
		},
		{
			name: "Fail, must use one of EnvoyResources, Resources",
			r: &marin3rv1alpha1.EnvoyConfig{
				Spec: marin3rv1alpha1.EnvoyConfigSpec{
					NodeID: "test",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validate(tt.r); (err != nil) != tt.wantErr {
				t.Errorf("EnvoyConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
