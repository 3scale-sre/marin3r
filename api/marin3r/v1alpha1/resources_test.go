package v1alpha1

import (
	"reflect"
	"testing"

	reconcilerutil "github.com/3scale-sre/basereconciler/util"
	envoy_serializer "github.com/3scale-sre/marin3r/api/envoy/serializer"
	"k8s.io/apimachinery/pkg/runtime"
)

func stringtoRawExtension(value string) *runtime.RawExtension {
	return &runtime.RawExtension{
		Raw: []byte(value),
	}
}

func TestEnvoyResources_Resources(t *testing.T) {
	type fields struct {
		Endpoints        []EnvoyResource
		Clusters         []EnvoyResource
		Routes           []EnvoyResource
		ScopedRoutes     []EnvoyResource
		Listeners        []EnvoyResource
		Runtimes         []EnvoyResource
		Secrets          []EnvoySecretResource
		ExtensionConfigs []EnvoyResource
	}
	type args struct {
		serialization envoy_serializer.Serialization
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []Resource
		wantErr bool
	}{
		{
			name: "Converts resources from v1alpha1 to v1alpha2 (json)",
			fields: fields{
				Endpoints: []EnvoyResource{
					{Value: "{\"cluster_name\": \"endpoint\"}"},
				},
				Clusters: []EnvoyResource{
					{Value: "{\"name\": \"cluster\"}"},
				},
				Secrets: []EnvoySecretResource{
					{Name: "secret"},
				},
			},
			args: args{
				serialization: envoy_serializer.JSON,
			},
			want: []Resource{
				{Type: "endpoint", Value: stringtoRawExtension("{\"cluster_name\": \"endpoint\"}")},
				{Type: "cluster", Value: stringtoRawExtension("{\"name\": \"cluster\"}")},
				{Type: "secret", GenerateFromTlsSecret: reconcilerutil.Pointer("secret"), Blueprint: reconcilerutil.Pointer(TlsCertificate)},
			},
			wantErr: false,
		},
		{
			name: "Converts resources from v1alpha1 to v1alpha2 (yaml)",
			fields: fields{
				Endpoints: []EnvoyResource{
					{Value: "cluster_name: endpoint"},
				},
				Clusters: []EnvoyResource{
					{Value: "name: cluster"},
				},
				Secrets: []EnvoySecretResource{
					{Name: "secret"},
				},
			},
			args: args{
				serialization: envoy_serializer.YAML,
			},
			want: []Resource{
				{Type: "endpoint", Value: stringtoRawExtension("{\"cluster_name\":\"endpoint\"}")},
				{Type: "cluster", Value: stringtoRawExtension("{\"name\":\"cluster\"}")},
				{Type: "secret", GenerateFromTlsSecret: reconcilerutil.Pointer("secret"), Blueprint: reconcilerutil.Pointer(TlsCertificate)},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := &EnvoyResources{
				Endpoints:        tt.fields.Endpoints,
				Clusters:         tt.fields.Clusters,
				Routes:           tt.fields.Routes,
				ScopedRoutes:     tt.fields.ScopedRoutes,
				Listeners:        tt.fields.Listeners,
				Runtimes:         tt.fields.Runtimes,
				Secrets:          tt.fields.Secrets,
				ExtensionConfigs: tt.fields.ExtensionConfigs,
			}
			got, err := in.Resources(tt.args.serialization)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnvoyResources.Resources() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EnvoyResources.Resources() = %v, want %v", got, tt.want)
			}
		})
	}
}
