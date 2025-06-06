package reconcilers

import (
	"context"
	"reflect"
	"testing"

	"github.com/3scale-sre/marin3r/api/envoy"
	envoy_resources "github.com/3scale-sre/marin3r/api/envoy/resources"
	envoy_resources_v3 "github.com/3scale-sre/marin3r/api/envoy/resources/v3"
	envoy_serializer "github.com/3scale-sre/marin3r/api/envoy/serializer"
	marin3rv1alpha1 "github.com/3scale-sre/marin3r/api/marin3r/v1alpha1"
	xdss "github.com/3scale-sre/marin3r/internal/pkg/discoveryservice/xdss"
	xdss_v3 "github.com/3scale-sre/marin3r/internal/pkg/discoveryservice/xdss/v3"
	k8sutil "github.com/3scale-sre/marin3r/internal/pkg/util/k8s"
	testutil "github.com/3scale-sre/marin3r/internal/pkg/util/test"
	"github.com/davecgh/go-spew/spew"
	envoy_config_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_config_endpoint_v3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	envoy_config_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_extensions_transport_sockets_tls_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	envoy_service_runtime_v3 "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestNewCacheReconciler(t *testing.T) {
	type args struct {
		ctx       context.Context
		logger    logr.Logger
		client    client.Client
		xdsCache  xdss.Cache
		decoder   envoy_serializer.ResourceUnmarshaller
		generator envoy_resources.Generator
	}

	tests := []struct {
		name string
		args args
		want CacheReconciler
	}{
		{
			name: "Returns a CacheReconciler (v3)",
			args: args{
				ctx:       context.TODO(),
				logger:    ctrl.Log.WithName("test"),
				client:    fake.NewClientBuilder().Build(),
				xdsCache:  xdss_v3.NewCache(),
				decoder:   envoy_serializer.NewResourceUnmarshaller(envoy_serializer.JSON, envoy.APIv3),
				generator: envoy_resources_v3.Generator{},
			},
			want: CacheReconciler{
				logger:    ctrl.Log.WithName("test"),
				client:    fake.NewClientBuilder().Build(),
				xdsCache:  xdss_v3.NewCache(),
				decoder:   envoy_serializer.NewResourceUnmarshaller(envoy_serializer.JSON, envoy.APIv3),
				generator: envoy_resources_v3.Generator{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewCacheReconciler(tt.args.client, tt.args.xdsCache, tt.args.decoder, tt.args.generator, tt.args.logger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCacheReconciler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCacheReconciler_Reconcile(t *testing.T) {
	type fields struct {
		ctx       context.Context
		logger    logr.Logger
		client    client.Client
		xdsCache  xdss.Cache
		decoder   envoy_serializer.ResourceUnmarshaller
		generator envoy_resources.Generator
	}

	type args struct {
		req       types.NamespacedName
		resources []marin3rv1alpha1.Resource
		nodeID    string
		version   string
	}

	tests := []struct {
		name        string
		fields      fields
		args        args
		want        *marin3rv1alpha1.VersionTracker
		wantErr     bool
		wantSnap    xdss.Snapshot
		wantVersion string
	}{
		{
			name: "Reconciles cache (v3)",
			fields: fields{
				ctx:       context.TODO(),
				logger:    ctrl.Log.WithName("test"),
				client:    fake.NewClientBuilder().Build(),
				xdsCache:  xdss_v3.NewCache(),
				decoder:   envoy_serializer.NewResourceUnmarshaller(envoy_serializer.JSON, envoy.APIv3),
				generator: envoy_resources_v3.Generator{},
			},
			args: args{
				req: types.NamespacedName{Name: "xx", Namespace: "xx"},
				resources: []marin3rv1alpha1.Resource{
					{
						Type:  envoy.Endpoint,
						Value: k8sutil.StringtoRawExtension("{\"cluster_name\": \"endpoint\"}"),
					},
				},
				version: "xxxx",
				nodeID:  "node2",
			},

			want:    &marin3rv1alpha1.VersionTracker{Endpoints: "845f965864"},
			wantErr: false,
			wantSnap: xdss_v3.NewSnapshot().SetResources(envoy.Endpoint, []envoy.Resource{
				&envoy_config_endpoint_v3.ClusterLoadAssignment{ClusterName: "endpoint"},
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &CacheReconciler{
				logger:    tt.fields.logger,
				client:    tt.fields.client,
				xdsCache:  tt.fields.xdsCache,
				decoder:   tt.fields.decoder,
				generator: tt.fields.generator,
			}

			got, err := r.Reconcile(context.TODO(), tt.args.req, tt.args.resources, tt.args.nodeID, tt.args.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("CacheReconciler.Reconcile() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CacheReconciler.Reconcile() = %v, want %v", got, tt.want)
			}

			gotSnap, _ := r.xdsCache.GetSnapshot(tt.args.nodeID)
			if !testutil.SnapshotsAreEqual(gotSnap, tt.wantSnap) {
				t.Errorf("CacheReconciler.Reconcile() Snapshot = E:%s C:%s R:%s SR:%s VH:%s L:%s S:%s RU:%s, want E:%s C:%s R:%s SR:%s VH:%s L:%s S:%s RU:%s",
					gotSnap.GetVersion(envoy.Endpoint), gotSnap.GetVersion(envoy.Cluster), gotSnap.GetVersion(envoy.Route), gotSnap.GetVersion(envoy.ScopedRoute),
					gotSnap.GetVersion(envoy.VirtualHost), gotSnap.GetVersion(envoy.Listener), gotSnap.GetVersion(envoy.Secret), gotSnap.GetVersion(envoy.Runtime),
					tt.wantSnap.GetVersion(envoy.Endpoint), tt.wantSnap.GetVersion(envoy.Cluster), tt.wantSnap.GetVersion(envoy.Route), tt.wantSnap.GetVersion(envoy.ScopedRoute),
					tt.wantSnap.GetVersion(envoy.VirtualHost), tt.wantSnap.GetVersion(envoy.Listener), tt.wantSnap.GetVersion(envoy.Secret), tt.wantSnap.GetVersion(envoy.Runtime),
				)
			}
		})
	}
}

func TestCacheReconciler_GenerateSnapshot(t *testing.T) {
	type fields struct {
		ctx       context.Context
		logger    logr.Logger
		client    client.Client
		xdsCache  xdss.Cache
		decoder   envoy_serializer.ResourceUnmarshaller
		generator envoy_resources.Generator
	}

	type args struct {
		req       types.NamespacedName
		resources []marin3rv1alpha1.Resource
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    xdss.Snapshot
		wantErr bool
	}{
		{
			name: "Loads v3 resources into the snapshot",
			fields: fields{
				ctx:       context.TODO(),
				logger:    ctrl.Log.WithName("test"),
				client:    fake.NewClientBuilder().Build(),
				xdsCache:  xdss_v3.NewCache(),
				decoder:   envoy_serializer.NewResourceUnmarshaller(envoy_serializer.JSON, envoy.APIv3),
				generator: envoy_resources_v3.Generator{},
			},
			args: args{
				req: types.NamespacedName{Name: "xx", Namespace: "xx"},
				resources: []marin3rv1alpha1.Resource{
					{Type: envoy.Endpoint, Value: k8sutil.StringtoRawExtension("{\"cluster_name\": \"endpoint\"}")},
					{Type: envoy.Cluster, Value: k8sutil.StringtoRawExtension("{\"name\": \"cluster\"}")},
					{Type: envoy.Route, Value: k8sutil.StringtoRawExtension("{\"name\": \"route\"}")},
					{Type: envoy.ScopedRoute, Value: k8sutil.StringtoRawExtension("{\"name\": \"scoped_route\"}")},
					{Type: envoy.Listener, Value: k8sutil.StringtoRawExtension("{\"name\": \"listener\"}")},
					{Type: envoy.Runtime, Value: k8sutil.StringtoRawExtension("{\"name\": \"runtime\"}")},
				},
			},
			want: xdss_v3.NewSnapshot().
				SetResources(envoy.Endpoint, []envoy.Resource{
					&envoy_config_endpoint_v3.ClusterLoadAssignment{ClusterName: "endpoint"},
				}).
				SetResources(envoy.Cluster, []envoy.Resource{
					&envoy_config_cluster_v3.Cluster{Name: "cluster"},
				}).
				SetResources(envoy.Route, []envoy.Resource{
					&envoy_config_route_v3.RouteConfiguration{Name: "route"},
				}).
				SetResources(envoy.ScopedRoute, []envoy.Resource{
					&envoy_config_route_v3.ScopedRouteConfiguration{Name: "scoped_route"},
				}).
				SetResources(envoy.Listener, []envoy.Resource{
					&envoy_config_listener_v3.Listener{Name: "listener"},
				}).
				SetResources(envoy.Runtime, []envoy.Resource{
					&envoy_service_runtime_v3.Runtime{Name: "runtime"},
				}),
			wantErr: false,
		},
		{
			name: "Error, bad endpoint value",
			fields: fields{
				ctx:       context.TODO(),
				logger:    ctrl.Log.WithName("test"),
				client:    fake.NewClientBuilder().Build(),
				xdsCache:  xdss_v3.NewCache(),
				decoder:   envoy_serializer.NewResourceUnmarshaller(envoy_serializer.JSON, envoy.APIv3),
				generator: envoy_resources_v3.Generator{},
			},
			args: args{
				req: types.NamespacedName{Name: "xx", Namespace: "xx"},
				resources: []marin3rv1alpha1.Resource{
					{Type: envoy.Endpoint, Value: k8sutil.StringtoRawExtension("giberish")},
				},
			},
			wantErr: true,
			want:    xdss_v3.NewSnapshot(),
		},
		{
			name: "Error, bad cluster value",
			fields: fields{
				ctx:       context.TODO(),
				logger:    ctrl.Log.WithName("test"),
				client:    fake.NewClientBuilder().Build(),
				xdsCache:  xdss_v3.NewCache(),
				decoder:   envoy_serializer.NewResourceUnmarshaller(envoy_serializer.JSON, envoy.APIv3),
				generator: envoy_resources_v3.Generator{},
			},
			args: args{
				req: types.NamespacedName{Name: "xx", Namespace: "xx"},
				resources: []marin3rv1alpha1.Resource{
					{Type: envoy.Cluster, Value: k8sutil.StringtoRawExtension("giberish")},
				},
			},
			wantErr: true,
			want:    xdss_v3.NewSnapshot(),
		},
		{
			name: "Error, bad route value",
			fields: fields{
				ctx:       context.TODO(),
				logger:    ctrl.Log.WithName("test"),
				client:    fake.NewClientBuilder().Build(),
				xdsCache:  xdss_v3.NewCache(),
				decoder:   envoy_serializer.NewResourceUnmarshaller(envoy_serializer.JSON, envoy.APIv3),
				generator: envoy_resources_v3.Generator{},
			},
			args: args{
				req: types.NamespacedName{Name: "xx", Namespace: "xx"},
				resources: []marin3rv1alpha1.Resource{
					{Type: envoy.Route, Value: k8sutil.StringtoRawExtension("giberish")},
				},
			},
			wantErr: true,
			want:    xdss_v3.NewSnapshot(),
		},
		{
			name: "Error, bad scoped route value",
			fields: fields{
				ctx:       context.TODO(),
				logger:    ctrl.Log.WithName("test"),
				client:    fake.NewClientBuilder().Build(),
				xdsCache:  xdss_v3.NewCache(),
				decoder:   envoy_serializer.NewResourceUnmarshaller(envoy_serializer.JSON, envoy.APIv3),
				generator: envoy_resources_v3.Generator{},
			},
			args: args{
				req: types.NamespacedName{Name: "xx", Namespace: "xx"},
				resources: []marin3rv1alpha1.Resource{
					{Type: envoy.ScopedRoute, Value: k8sutil.StringtoRawExtension("giberish")},
				},
			},
			wantErr: true,
			want:    xdss_v3.NewSnapshot(),
		},
		{
			name: "Error, bad listener value",
			fields: fields{
				ctx:       context.TODO(),
				logger:    ctrl.Log.WithName("test"),
				client:    fake.NewClientBuilder().Build(),
				xdsCache:  xdss_v3.NewCache(),
				decoder:   envoy_serializer.NewResourceUnmarshaller(envoy_serializer.JSON, envoy.APIv3),
				generator: envoy_resources_v3.Generator{},
			},
			args: args{
				req: types.NamespacedName{Name: "xx", Namespace: "xx"},
				resources: []marin3rv1alpha1.Resource{
					{Type: envoy.Listener, Value: k8sutil.StringtoRawExtension("giberish")},
				},
			},
			wantErr: true,
			want:    xdss_v3.NewSnapshot(),
		},
		{
			name: "Error, bad runtime value",
			fields: fields{
				ctx:       context.TODO(),
				logger:    ctrl.Log.WithName("test"),
				client:    fake.NewClientBuilder().Build(),
				xdsCache:  xdss_v3.NewCache(),
				decoder:   envoy_serializer.NewResourceUnmarshaller(envoy_serializer.JSON, envoy.APIv3),
				generator: envoy_resources_v3.Generator{},
			},
			args: args{
				req: types.NamespacedName{Name: "xx", Namespace: "xx"},
				resources: []marin3rv1alpha1.Resource{
					{Type: envoy.Runtime, Value: k8sutil.StringtoRawExtension("giberish")},
				},
			},
			wantErr: true,
			want:    xdss_v3.NewSnapshot(),
		},
		{
			name: "Loads secret:tlsCertificate resources into the snapshot (v3)",
			fields: fields{
				ctx:    context.TODO(),
				logger: ctrl.Log.WithName("test"),
				client: fake.NewClientBuilder().WithObjects(&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{Name: "secret", Namespace: "xx"},
					Type:       corev1.SecretTypeTLS,
					Data:       map[string][]byte{"tls.crt": []byte("cert"), "tls.key": []byte("key")},
				}).Build(),
				xdsCache:  xdss_v3.NewCache(),
				decoder:   envoy_serializer.NewResourceUnmarshaller(envoy_serializer.JSON, envoy.APIv3),
				generator: envoy_resources_v3.Generator{},
			},
			args: args{
				req: types.NamespacedName{Name: "xx", Namespace: "xx"},
				resources: []marin3rv1alpha1.Resource{
					{Type: envoy.Secret, GenerateFromTlsSecret: ptr.To("secret")},
				},
			},
			wantErr: false,
			want: xdss_v3.NewSnapshot().
				SetResources(envoy.Secret, []envoy.Resource{
					&envoy_extensions_transport_sockets_tls_v3.Secret{
						Name: "secret",
						Type: &envoy_extensions_transport_sockets_tls_v3.Secret_TlsCertificate{
							TlsCertificate: &envoy_extensions_transport_sockets_tls_v3.TlsCertificate{
								PrivateKey: &envoy_config_core_v3.DataSource{
									Specifier: &envoy_config_core_v3.DataSource_InlineBytes{InlineBytes: []byte("key")},
								},
								CertificateChain: &envoy_config_core_v3.DataSource{
									Specifier: &envoy_config_core_v3.DataSource_InlineBytes{InlineBytes: []byte("cert")},
								}}}}}),
		},
		{
			name: "Loads secret:validationContext resources into the snapshot (v3)",
			fields: fields{
				ctx:    context.TODO(),
				logger: ctrl.Log.WithName("test"),
				client: fake.NewClientBuilder().WithObjects(&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{Name: "secret", Namespace: "xx"},
					Type:       corev1.SecretTypeTLS,
					Data:       map[string][]byte{"tls.crt": []byte("cert"), "tls.key": []byte("key")},
				}).Build(),
				xdsCache:  xdss_v3.NewCache(),
				decoder:   envoy_serializer.NewResourceUnmarshaller(envoy_serializer.JSON, envoy.APIv3),
				generator: envoy_resources_v3.Generator{},
			},
			args: args{
				req: types.NamespacedName{Name: "xx", Namespace: "xx"},
				resources: []marin3rv1alpha1.Resource{
					{
						Type:                  envoy.Secret,
						GenerateFromTlsSecret: ptr.To("secret"),
						Blueprint:             ptr.To(marin3rv1alpha1.TlsValidationContext),
					},
				},
			},
			wantErr: false,
			want: xdss_v3.NewSnapshot().
				SetResources(envoy.Secret, []envoy.Resource{
					&envoy_extensions_transport_sockets_tls_v3.Secret{
						Name: "secret",
						Type: &envoy_extensions_transport_sockets_tls_v3.Secret_ValidationContext{
							ValidationContext: &envoy_extensions_transport_sockets_tls_v3.CertificateValidationContext{
								TrustedCa: &envoy_config_core_v3.DataSource{
									Specifier: &envoy_config_core_v3.DataSource_InlineBytes{InlineBytes: []byte("cert")},
								},
							},
						},
					}}),
		},
		{
			name: "Fails with wrong secret type",
			fields: fields{
				client: fake.NewClientBuilder().WithObjects(&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{Name: "secret", Namespace: "xx"},
					Type:       corev1.SecretTypeBasicAuth,
					Data:       map[string][]byte{"tls.crt": []byte("cert"), "tls.key": []byte("key")},
				}).Build(),
				ctx:       context.TODO(),
				logger:    ctrl.Log.WithName("test"),
				xdsCache:  xdss_v3.NewCache(),
				decoder:   envoy_serializer.NewResourceUnmarshaller(envoy_serializer.JSON, envoy.APIv3),
				generator: envoy_resources_v3.Generator{},
			},
			args: args{
				req: types.NamespacedName{Name: "xx", Namespace: "xx"},
				resources: []marin3rv1alpha1.Resource{
					{Type: envoy.Secret, GenerateFromTlsSecret: ptr.To("secret")},
				},
			},
			wantErr: true,
			want:    xdss_v3.NewSnapshot(),
		},
		{
			name: "Fails when secret does not exist",
			fields: fields{
				client:    fake.NewClientBuilder().Build(),
				ctx:       context.TODO(),
				logger:    ctrl.Log.WithName("test"),
				xdsCache:  xdss_v3.NewCache(),
				decoder:   envoy_serializer.NewResourceUnmarshaller(envoy_serializer.JSON, envoy.APIv3),
				generator: envoy_resources_v3.Generator{},
			},
			args: args{
				req: types.NamespacedName{Name: "xx", Namespace: "xx"},
				resources: []marin3rv1alpha1.Resource{
					{Type: envoy.Secret, GenerateFromTlsSecret: ptr.To("secret")},
				},
			},
			wantErr: true,
			want:    xdss_v3.NewSnapshot(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &CacheReconciler{
				logger:    tt.fields.logger,
				client:    tt.fields.client,
				xdsCache:  tt.fields.xdsCache,
				decoder:   tt.fields.decoder,
				generator: tt.fields.generator,
			}

			got, err := r.GenerateSnapshot(context.TODO(), tt.args.req, tt.args.resources)
			if (err != nil) != tt.wantErr {
				t.Errorf("CacheReconciler.GenerateSnapshot() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !tt.wantErr && !testutil.SnapshotsAreEqual(got, tt.want) {
				spew.Dump(got)
				t.Errorf("CacheReconciler.Reconcile() Snapshot = E:%s C:%s R:%s SR:%s VH:%s L:%s S:%s RU:%s, want E:%s C:%s R:%s SR:%s VH:%s L:%s S:%s RU:%s",
					got.GetVersion(envoy.Endpoint), got.GetVersion(envoy.Cluster), got.GetVersion(envoy.Route), got.GetVersion(envoy.ScopedRoute),
					got.GetVersion(envoy.VirtualHost), got.GetVersion(envoy.Listener), got.GetVersion(envoy.Secret), got.GetVersion(envoy.Runtime),
					tt.want.GetVersion(envoy.Endpoint), tt.want.GetVersion(envoy.Cluster), tt.want.GetVersion(envoy.Route), tt.want.GetVersion(envoy.ScopedRoute),
					tt.want.GetVersion(envoy.VirtualHost), tt.want.GetVersion(envoy.Listener), tt.want.GetVersion(envoy.Secret), tt.want.GetVersion(envoy.Runtime),
				)
			}
		})
	}
}
