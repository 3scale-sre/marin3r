package v1

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/3scale-sre/marin3r/api/envoy/defaults"
	operatorv1alpha1 "github.com/3scale-sre/marin3r/api/operator.marin3r/v1alpha1"
	envoy_container "github.com/3scale-sre/marin3r/internal/pkg/envoy/container"
	"github.com/3scale-sre/marin3r/internal/pkg/envoy/container/shutdownmanager"
	"github.com/go-test/deep"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func init() {
	operatorv1alpha1.AddToScheme(scheme.Scheme)

	deep.CompareUnexportedFields = true
}

func Test_envoySidecarConfig_PopulateFromAnnotation(t *testing.T) {
	type args struct {
		ctx         context.Context
		clnt        client.Client
		namespace   string
		annotations map[string]string
	}

	tests := []struct {
		name    string
		esc     *envoySidecarConfig
		args    args
		want    *envoySidecarConfig
		wantErr bool
	}{
		{
			"Populate ContainerConfig from annotations",
			&envoySidecarConfig{},
			args{
				ctx: context.TODO(),
				clnt: fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(
					&operatorv1alpha1.DiscoveryService{ObjectMeta: metav1.ObjectMeta{Name: "ds", Namespace: "test"}},
				).WithStatusSubresource(&operatorv1alpha1.DiscoveryService{}).Build(),
				namespace: "test",
				annotations: map[string]string{
					"marin3r.3scale.net/node-id":                                                "node-id",
					"marin3r.3scale.net/ports":                                                  "xxxx:1111",
					"marin3r.3scale.net/host-port-mappings":                                     "xxxx:3000",
					"marin3r.3scale.net/container-name":                                         "container",
					"marin3r.3scale.net/envoy-image":                                            "image",
					"marin3r.3scale.net/ads-configmap":                                          "cm",
					"marin3r.3scale.net/cluster-id":                                             "cluster-id",
					"marin3r.3scale.net/config-volume":                                          "config-volume",
					"marin3r.3scale.net/tls-volume":                                             "tls-volume",
					"marin3r.3scale.net/client-certificate":                                     "client-cert",
					"marin3r.3scale.net/envoy-extra-args":                                       "--log-level debug",
					"marin3r.3scale.net/admin.port":                                             "2000",
					"marin3r.3scale.net/admin.bind-address":                                     "127.0.0.1",
					"marin3r.3scale.net/admin.access-log-path":                                  "/dev/stdout",
					"marin3r.3scale.net/envoy-api-version":                                      "v3",
					fmt.Sprintf("%s/%s", marin3rAnnotationsDomain, paramResourceRequestsCPU):    "500m",
					fmt.Sprintf("%s/%s", marin3rAnnotationsDomain, paramResourceRequestsMemory): "700Mi",
					fmt.Sprintf("%s/%s", marin3rAnnotationsDomain, paramResourceLimitsCPU):      "1000m",
					fmt.Sprintf("%s/%s", marin3rAnnotationsDomain, paramResourceLimitsMemory):   "900Mi",
				}},
			&envoySidecarConfig{
				generator: envoy_container.ContainerConfig{
					Name:                         "container",
					Image:                        "image",
					ConfigBasePath:               defaults.EnvoyConfigBasePath,
					ConfigFileName:               defaults.EnvoyConfigFileName,
					ConfigVolume:                 "config-volume",
					TLSBasePath:                  defaults.EnvoyTLSBasePath,
					TLSVolume:                    "tls-volume",
					NodeID:                       "node-id",
					ClusterID:                    "cluster-id",
					ClientCertSecret:             "client-cert",
					ExtraArgs:                    []string{"--log-level", "debug"},
					Resources:                    corev1.ResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("500m"), corev1.ResourceMemory: resource.MustParse("700Mi")}, Limits: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1000m"), corev1.ResourceMemory: resource.MustParse("900Mi")}},
					AdminBindAddress:             "127.0.0.1",
					AdminPort:                    2000,
					AdminAccessLogPath:           "/dev/stdout",
					Ports:                        []corev1.ContainerPort{{Name: "xxxx", ContainerPort: 1111, HostPort: 3000}},
					LivenessProbe:                operatorv1alpha1.ProbeSpec{InitialDelaySeconds: 30, TimeoutSeconds: 1, PeriodSeconds: 10, SuccessThreshold: 1, FailureThreshold: 10},
					ReadinessProbe:               operatorv1alpha1.ProbeSpec{InitialDelaySeconds: 15, TimeoutSeconds: 1, PeriodSeconds: 5, SuccessThreshold: 1, FailureThreshold: 1},
					InitManagerImage:             defaults.InitMgrImage(),
					XdssHost:                     "marin3r-ds.test.svc",
					XdssPort:                     18000,
					APIVersion:                   "v3",
					ShutdownManagerEnabled:       false,
					ShutdownManagerPort:          int32(defaults.ShtdnMgrDefaultServerPort),
					ShutdownManagerImage:         defaults.ShtdnMgrImage(),
					ShutdownManagerDrainSeconds:  300,
					ShutdownManagerDrainStrategy: defaults.DrainStrategyGradual,
				},
			},
			false,
		},
		{
			"Populate ContainerConfig from annotations (all default)",
			&envoySidecarConfig{},
			args{
				ctx: context.TODO(),
				clnt: fake.NewClientBuilder().WithObjects(
					&operatorv1alpha1.DiscoveryService{ObjectMeta: metav1.ObjectMeta{Name: "ds", Namespace: "test"}},
				).WithStatusSubresource(&operatorv1alpha1.DiscoveryService{}).Build(),
				namespace: "test",
				annotations: map[string]string{
					"marin3r.3scale.net/node-id": "node-id",
				}},
			&envoySidecarConfig{
				generator: envoy_container.ContainerConfig{
					Name:                         defaults.SidecarContainerName,
					Image:                        defaults.Image,
					ConfigBasePath:               defaults.EnvoyConfigBasePath,
					ConfigFileName:               defaults.EnvoyConfigFileName,
					ConfigVolume:                 defaults.SidecarConfigVolume,
					TLSBasePath:                  defaults.EnvoyTLSBasePath,
					TLSVolume:                    defaults.SidecarTLSVolume,
					NodeID:                       "node-id",
					ClusterID:                    "node-id",
					ClientCertSecret:             defaults.SidecarClientCertificate,
					Resources:                    corev1.ResourceRequirements{},
					AdminBindAddress:             defaults.EnvoyAdminBindAddress,
					AdminPort:                    int32(defaults.EnvoyAdminPort),
					AdminAccessLogPath:           defaults.EnvoyAdminAccessLogPath,
					Ports:                        []corev1.ContainerPort{},
					LivenessProbe:                operatorv1alpha1.ProbeSpec{InitialDelaySeconds: 30, TimeoutSeconds: 1, PeriodSeconds: 10, SuccessThreshold: 1, FailureThreshold: 10},
					ReadinessProbe:               operatorv1alpha1.ProbeSpec{InitialDelaySeconds: 15, TimeoutSeconds: 1, PeriodSeconds: 5, SuccessThreshold: 1, FailureThreshold: 1},
					InitManagerImage:             defaults.InitMgrImage(),
					XdssHost:                     "marin3r-ds.test.svc",
					XdssPort:                     18000,
					APIVersion:                   defaults.EnvoyAPIVersion,
					ShutdownManagerEnabled:       false,
					ShutdownManagerPort:          int32(defaults.ShtdnMgrDefaultServerPort),
					ShutdownManagerImage:         defaults.ShtdnMgrImage(),
					ShutdownManagerDrainSeconds:  300,
					ShutdownManagerDrainStrategy: defaults.DrainStrategyGradual,
				},
			},
			false,
		},
		{
			"Populate ContainerConfig from annotations (shtdnmgr enabled)",
			&envoySidecarConfig{},
			args{
				ctx: context.TODO(),
				clnt: fake.NewClientBuilder().WithObjects(
					&operatorv1alpha1.DiscoveryService{ObjectMeta: metav1.ObjectMeta{Name: "ds", Namespace: "test"}},
				).WithStatusSubresource(&operatorv1alpha1.DiscoveryService{}).Build(),
				namespace: "test",
				annotations: map[string]string{
					"marin3r.3scale.net/node-id":                         "node-id",
					"marin3r.3scale.net/shutdown-manager.enabled":        "true",
					"marin3r.3scale.net/shutdown-manager.port":           "30000",
					"marin3r.3scale.net/shutdown-manager.image":          "image:test",
					"marin3r.3scale.net/shutdown-manager.drain-time":     "50",
					"marin3r.3scale.net/shutdown-manager.drain-strategy": "immediate",
				}},
			&envoySidecarConfig{
				generator: envoy_container.ContainerConfig{
					Name:                         defaults.SidecarContainerName,
					Image:                        defaults.Image,
					ConfigBasePath:               defaults.EnvoyConfigBasePath,
					ConfigFileName:               defaults.EnvoyConfigFileName,
					ConfigVolume:                 defaults.SidecarConfigVolume,
					TLSBasePath:                  defaults.EnvoyTLSBasePath,
					TLSVolume:                    defaults.SidecarTLSVolume,
					NodeID:                       "node-id",
					ClusterID:                    "node-id",
					ClientCertSecret:             defaults.SidecarClientCertificate,
					Resources:                    corev1.ResourceRequirements{},
					AdminBindAddress:             defaults.EnvoyAdminBindAddress,
					AdminPort:                    int32(defaults.EnvoyAdminPort),
					AdminAccessLogPath:           defaults.EnvoyAdminAccessLogPath,
					Ports:                        []corev1.ContainerPort{},
					LivenessProbe:                operatorv1alpha1.ProbeSpec{InitialDelaySeconds: 30, TimeoutSeconds: 1, PeriodSeconds: 10, SuccessThreshold: 1, FailureThreshold: 10},
					ReadinessProbe:               operatorv1alpha1.ProbeSpec{InitialDelaySeconds: 15, TimeoutSeconds: 1, PeriodSeconds: 5, SuccessThreshold: 1, FailureThreshold: 1},
					InitManagerImage:             defaults.InitMgrImage(),
					XdssHost:                     "marin3r-ds.test.svc",
					XdssPort:                     18000,
					APIVersion:                   defaults.EnvoyAPIVersion,
					ShutdownManagerEnabled:       true,
					ShutdownManagerPort:          30000,
					ShutdownManagerImage:         "image:test",
					ShutdownManagerDrainSeconds:  50,
					ShutdownManagerDrainStrategy: defaults.DrainStrategyImmediate,
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.esc.PopulateFromAnnotations(tt.args.ctx, tt.args.clnt, tt.args.namespace, tt.args.annotations); (err != nil) != tt.wantErr {
				t.Errorf("envoySidecarConfig.PopulateFromAnnotations() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := deep.Equal(tt.esc, tt.want); len(diff) > 0 {
				t.Errorf("envoySidecarConfig.PopulateFromAnnotations() = diff %v", diff)
			}
		})
	}
}

func Test_getContainerPorts(t *testing.T) {
	type args struct {
		annotations map[string]string
	}

	tests := []struct {
		name    string
		args    args
		want    []corev1.ContainerPort
		wantErr bool
	}{
		{
			"Slice of ContainerPorts from annotation, one port",
			args{map[string]string{
				"marin3r.3scale.net/ports": "xxxx:1111",
			}},
			[]corev1.ContainerPort{
				{Name: "xxxx", ContainerPort: 1111},
			},
			false,
		}, {
			"Slice of ContainerPorts from annotation, multiple ports",
			args{map[string]string{
				"marin3r.3scale.net/ports": "xxxx:1111,yyyy:2222,zzzz:3333",
			}},
			[]corev1.ContainerPort{
				{Name: "xxxx", ContainerPort: 1111},
				{Name: "yyyy", ContainerPort: 2222},
				{Name: "zzzz", ContainerPort: 3333},
			},
			false,
		}, {
			"Wrong annotations produces empty slice",
			args{map[string]string{
				"marin3r.3scale.net/xxxx": "xxxx:1111,yyyy:2222,zzzz:3333",
			}},
			[]corev1.ContainerPort{},
			false,
		}, {
			"No annotation produces empty slice",
			args{map[string]string{}},
			[]corev1.ContainerPort{},
			false,
		}, {
			"Mix spec with proto and spec without proto",
			args{map[string]string{
				"marin3r.3scale.net/ports": "xxxx:1111:UDP,yyyy:2222",
			}},
			[]corev1.ContainerPort{
				{Name: "xxxx", ContainerPort: 1111, Protocol: "UDP"},
				{Name: "yyyy", ContainerPort: 2222},
			},
			false,
		}, {
			"With host-port-mapping annotation",
			args{map[string]string{
				"marin3r.3scale.net/ports":              "xxxx:1111:UDP,yyyy:2222,zzzz:3333",
				"marin3r.3scale.net/host-port-mappings": "xxxx:5000,yyyy:6000",
			}},
			[]corev1.ContainerPort{
				{Name: "xxxx", ContainerPort: 1111, Protocol: "UDP", HostPort: 5000},
				{Name: "yyyy", ContainerPort: 2222, HostPort: 6000},
				{Name: "zzzz", ContainerPort: 3333},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getContainerPorts(tt.args.annotations)
			if (err != nil) != tt.wantErr {
				t.Errorf("getContainerPorts() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getContainerPorts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parsePortSpec(t *testing.T) {
	type args struct {
		spec string
	}

	tests := []struct {
		name    string
		args    args
		want    *corev1.ContainerPort
		wantErr bool
	}{
		{
			"ContainerPort struct from annotation string, without protocol",
			args{"http:3000"},
			&corev1.ContainerPort{
				Name:          "http",
				ContainerPort: 3000,
			},
			false,
		}, {
			"ContainerPort struct from annotation string, with TCP protocol",
			args{"http:8080:TCP"},
			&corev1.ContainerPort{
				Name:          "http",
				ContainerPort: 8080,
				Protocol:      corev1.Protocol("TCP"),
			},
			false,
		}, {
			"ContainerPort struct from annotation string, with UDP protocol",
			args{"http:8080:UDP"},
			&corev1.ContainerPort{
				Name:          "http",
				ContainerPort: 8080,
				Protocol:      corev1.Protocol("UDP"),
			},
			false,
		}, {
			"ContainerPort struct from annotation string, with SCTP protocol",
			args{"http:8080:SCTP"},
			&corev1.ContainerPort{
				Name:          "http",
				ContainerPort: 8080,
				Protocol:      corev1.Protocol("SCTP"),
			},
			false,
		}, {
			"Error, wrong protocol",
			args{"http:8080:XXX"},
			nil,
			true,
		}, {
			"Error, privileged port",
			args{"http:80:TCP"},
			nil,
			true,
		},
		{
			"Error, missing params",
			args{"http"},
			nil,
			true,
		}, {
			"Error, not a port",
			args{"http:xxxx"},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePortSpec(tt.args.spec)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePortSpec() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parsePortSpec() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_hostPortMapping(t *testing.T) {
	type args struct {
		portName    string
		annotations map[string]string
	}

	tests := []struct {
		name    string
		args    args
		want    int32
		wantErr bool
	}{
		{
			"Parse a single 'host-port-mapping' spec",
			args{"http", map[string]string{"marin3r.3scale.net/host-port-mappings": "http:3000"}},
			3000,
			false,
		}, {
			"Parse several 'host-port-mapping' specs",
			args{"admin", map[string]string{"marin3r.3scale.net/host-port-mappings": "http:3000,admin:6000"}},
			6000,
			false,
		}, {
			"Incorrect 'host-port-mapping' spec 1",
			args{"http", map[string]string{"marin3r.3scale.net/host-port-mappings": "admin,6000"}},
			0,
			true,
		}, {
			"Not a port",
			args{"admin", map[string]string{"marin3r.3scale.net/host-port-mappings": "admin:4000i"}},
			0,
			true,
		}, {
			"Privileged port not allowed",
			args{"admin", map[string]string{"marin3r.3scale.net/host-port-mappings": "admin:80"}},
			0,
			true,
		}, {
			"Port not found in spec",
			args{"http", map[string]string{"marin3r.3scale.net/host-port-mappings": "admin:80"}},
			0,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := hostPortMapping(tt.args.portName, tt.args.annotations)
			if (err != nil) != tt.wantErr {
				t.Errorf("hostPortMapping() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if got != tt.want {
				t.Errorf("hostPortMapping() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_portNumber(t *testing.T) {
	type args struct {
		sport string
	}

	tests := []struct {
		name    string
		args    args
		want    int32
		wantErr bool
	}{
		{"Parse port string", args{"1111"}, 1111, false},
		{"Error, not a number", args{"xxxx"}, 0, true},
		{"Port 1024 is allowed", args{"1023"}, 0, true},
		{"Error, privileged port 1024", args{"1023"}, 0, true},
		{"Error, privileged port 0", args{"0"}, 0, true},
		{"Error, privileged port 100", args{"100"}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := portNumber(tt.args.sport)
			if (err != nil) != tt.wantErr {
				t.Errorf("portNumber() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if got != tt.want {
				t.Errorf("portNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getStringParam(t *testing.T) {
	type args struct {
		key         string
		annotations map[string]string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"Return string value from annotation",
			args{"envoy-image", map[string]string{"marin3r.3scale.net/envoy-image": "image"}},
			"image",
		}, {
			"Return string value from default",
			args{"envoy-image", map[string]string{}},
			defaults.Image,
		}, {
			"Return cluster-id from annotation",
			args{"cluster-id", map[string]string{"marin3r.3scale.net/cluster-id": "cluster-id"}},
			"cluster-id",
		}, {
			"Return cluster-id from default (defaults to node-id)",
			args{"cluster-id", map[string]string{"marin3r.3scale.net/node-id": "test"}},
			"test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getStringParam(tt.args.key, tt.args.annotations); got != tt.want {
				t.Errorf("getStringParam() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getNodeID(t *testing.T) {
	type args struct {
		annotations map[string]string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"Return node-id from annotation",
			args{map[string]string{"marin3r.3scale.net/node-id": "test-id"}},
			"test-id",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getNodeID(tt.args.annotations); got != tt.want {
				t.Errorf("getNodeID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_lookupMarin3rAnnotation(t *testing.T) {
	type args struct {
		annotations map[string]string
		key         string
	}

	tests := []struct {
		name   string
		args   args
		want   string
		wantOk bool
	}{
		{
			"Marin3r annotation exists",
			args{
				annotations: map[string]string{
					fmt.Sprintf("%s/%s", marin3rAnnotationsDomain, "existingkey"): "examplevalue",
				},
				key: "existingkey",
			},
			"examplevalue",
			true,
		},
		{
			"Marin3r annotation exists and is empty",
			args{
				annotations: map[string]string{
					fmt.Sprintf("%s/%s", marin3rAnnotationsDomain, "existingkey"): "",
				},
				key: "existingkey",
			},
			"",
			true,
		},
		{
			"Marin3r annotation does not exist",
			args{
				annotations: map[string]string{
					fmt.Sprintf("%s/%s", marin3rAnnotationsDomain, "existingkey"): "myval",
				},
				key: "unexistingkey",
			},
			"",
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := lookupMarin3rAnnotation(tt.args.key, tt.args.annotations)
			if ok != tt.wantOk {
				t.Errorf("lookupMarin3rAnnotation() ok = %v, wantOk %v", got, tt.wantOk)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("lookupMarin3rAnnotation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getContainerResourceRequirements(t *testing.T) {
	type args struct {
		annotations map[string]string
	}

	tests := []struct {
		name    string
		args    args
		want    corev1.ResourceRequirements
		wantErr bool
	}{
		{
			"No resource requirement annotations",
			args{map[string]string{}},
			corev1.ResourceRequirements{},
			false,
		},
		{
			"invalid resource requirement annotation value",
			args{map[string]string{
				fmt.Sprintf("%s/%s", marin3rAnnotationsDomain, paramResourceLimitsMemory): "invalidMemoryValue",
			}},
			corev1.ResourceRequirements{},
			true,
		},
		{
			"resource requirement annotation set but invalid empty value",
			args{map[string]string{
				fmt.Sprintf("%s/%s", marin3rAnnotationsDomain, paramResourceRequestsCPU): "",
			}},
			corev1.ResourceRequirements{},
			true,
		},
		{
			"resource cpu request set",
			args{map[string]string{
				fmt.Sprintf("%s/%s", marin3rAnnotationsDomain, paramResourceRequestsCPU): "100m",
			}},
			corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU: resource.MustParse("100m"),
				},
			},
			false,
		},
		{
			"resource cpu limit set",
			args{map[string]string{
				fmt.Sprintf("%s/%s", marin3rAnnotationsDomain, paramResourceLimitsCPU): "200m",
			}},
			corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU: resource.MustParse("200m"),
				},
			},
			false,
		},
		{
			"resource memory request set",
			args{map[string]string{
				fmt.Sprintf("%s/%s", marin3rAnnotationsDomain, paramResourceRequestsMemory): "100Mi",
			}},
			corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("100Mi"),
				},
			},
			false,
		},
		{
			"resource memory limit set",
			args{map[string]string{
				fmt.Sprintf("%s/%s", marin3rAnnotationsDomain, paramResourceLimitsMemory): "200Mi",
			}},
			corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("200Mi"),
				},
			},
			false,
		},
		{
			"resource requests and limits set",
			args{map[string]string{
				fmt.Sprintf("%s/%s", marin3rAnnotationsDomain, paramResourceRequestsCPU):    "500m",
				fmt.Sprintf("%s/%s", marin3rAnnotationsDomain, paramResourceRequestsMemory): "700Mi",
				fmt.Sprintf("%s/%s", marin3rAnnotationsDomain, paramResourceLimitsCPU):      "1000m",
				fmt.Sprintf("%s/%s", marin3rAnnotationsDomain, paramResourceLimitsMemory):   "900Mi",
			}},
			corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("500m"),
					corev1.ResourceMemory: resource.MustParse("700Mi"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1000m"),
					corev1.ResourceMemory: resource.MustParse("900Mi"),
				},
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getContainerResourceRequirements(tt.args.annotations)
			if (err != nil) != tt.wantErr {
				t.Errorf("getContainerResourceRequirements() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !equality.Semantic.DeepEqual(got, tt.want) {
				t.Errorf("getContainerResourceRequirements() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isShtdnMgrEnabled(t *testing.T) {
	type args struct {
		annotations map[string]string
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Returns true (value: true)",
			args: args{
				annotations: map[string]string{
					fmt.Sprintf("%s/%s", marin3rAnnotationsDomain, "other-stuff"):        "aaaa",
					fmt.Sprintf("%s/%s", marin3rAnnotationsDomain, paramShtdnMgrEnabled): "true",
				},
			},
			want: true,
		},
		{
			name: "Returns false (value: false)",
			args: args{
				annotations: map[string]string{
					fmt.Sprintf("%s/%s", marin3rAnnotationsDomain, "other-stuff"):        "aaaa",
					fmt.Sprintf("%s/%s", marin3rAnnotationsDomain, paramShtdnMgrEnabled): "false",
				},
			},
			want: false,
		},
		{
			name: "Returns false (bad value)",
			args: args{
				annotations: map[string]string{
					fmt.Sprintf("%s/%s", marin3rAnnotationsDomain, "other-stuff"):        "aaaa",
					fmt.Sprintf("%s/%s", marin3rAnnotationsDomain, paramShtdnMgrEnabled): "bad_value",
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isShtdnMgrEnabled(tt.args.annotations); got != tt.want {
				t.Errorf("isShtdnMgrEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getPortOrDefault(t *testing.T) {
	type args struct {
		key         string
		annotations map[string]string
		defaultPort uint32
	}

	tests := []struct {
		name string
		args args
		want int32
	}{
		{
			name: "Returns the port in the annotation",
			args: args{
				key: "port",
				annotations: map[string]string{
					fmt.Sprintf("%s/%s", marin3rAnnotationsDomain, "other-stuff"): "aaaa",
					fmt.Sprintf("%s/%s", marin3rAnnotationsDomain, "port"):        "5000",
				},
				defaultPort: 1000,
			},
			want: 5000,
		},
		{
			name: "Returns the default port if unset",
			args: args{
				key:         "port",
				annotations: map[string]string{},
				defaultPort: 1000,
			},
			want: 1000,
		},
		{
			name: "Returns the default port if bad port provided",
			args: args{
				key: "port",
				annotations: map[string]string{
					fmt.Sprintf("%s/%s", marin3rAnnotationsDomain, "other-stuff"): "aaaa",
					fmt.Sprintf("%s/%s", marin3rAnnotationsDomain, "port"):        "80",
				},
				defaultPort: 1000,
			},
			want: 1000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getPortOrDefault(tt.args.key, tt.args.annotations, tt.args.defaultPort); got != tt.want {
				t.Errorf("getPortOrDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_envoySidecarConfig_GetDiscoveryServiceAddress(t *testing.T) {
	type args struct {
		ctx         context.Context
		clnt        client.Client
		namespace   string
		annotations map[string]string
	}

	tests := []struct {
		name       string
		args       args
		wantServer string
		wantPort   int
		wantErr    bool
	}{
		{
			name: "Returns the address using the 'discovery-service.name' annotations",
			args: args{
				ctx: context.TODO(),
				clnt: fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(
					&operatorv1alpha1.DiscoveryService{
						ObjectMeta: metav1.ObjectMeta{Name: "ds", Namespace: "test"},
						Spec: operatorv1alpha1.DiscoveryServiceSpec{
							XdsServerPort: func() *uint32 {
								var p uint32 = 20000

								return &p
							}(),
							ServiceConfig: &operatorv1alpha1.ServiceConfig{
								Name: "example",
							},
						},
					},
				).WithStatusSubresource(&operatorv1alpha1.DiscoveryService{}).Build(),
				namespace: "test",
				annotations: map[string]string{
					"marin3r.3scale.net/discovery-service.name": "ds",
				},
			},
			wantServer: "example.test.svc",
			wantPort:   20000,
			wantErr:    false,
		},
		{
			name: "Returns the address without the 'discovery-service.name' annotation",
			args: args{
				ctx: context.TODO(),
				clnt: fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(
					&operatorv1alpha1.DiscoveryService{
						ObjectMeta: metav1.ObjectMeta{Name: "ds", Namespace: "test"},
						Spec: operatorv1alpha1.DiscoveryServiceSpec{
							XdsServerPort: func() *uint32 {
								var p uint32 = 20000

								return &p
							}(),
							ServiceConfig: &operatorv1alpha1.ServiceConfig{
								Name: "example",
							},
						},
					},
				).WithStatusSubresource(&operatorv1alpha1.DiscoveryService{}).Build(),
				namespace:   "test",
				annotations: map[string]string{},
			},
			wantServer: "example.test.svc",
			wantPort:   20000,
			wantErr:    false,
		},
		{
			name: "'discovery-service.name' annotation points to inexistent resource",
			args: args{
				ctx:       context.TODO(),
				clnt:      fake.NewClientBuilder().WithScheme(scheme.Scheme).Build(),
				namespace: "test",
				annotations: map[string]string{
					"marin3r.3scale.net/discovery-service.name": "ds",
				},
			},
			wantServer: "",
			wantPort:   -1,
			wantErr:    true,
		},
		{
			name: "wrong number of discoveryservices found when 'discovery-service.name' not set",
			args: args{
				ctx: context.TODO(),
				clnt: fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(
					&operatorv1alpha1.DiscoveryService{ObjectMeta: metav1.ObjectMeta{Name: "ds", Namespace: "test"}},
					&operatorv1alpha1.DiscoveryService{ObjectMeta: metav1.ObjectMeta{Name: "other", Namespace: "test"}},
				).WithStatusSubresource(&operatorv1alpha1.DiscoveryService{}).Build(),
				namespace:   "test",
				annotations: map[string]string{},
			},
			wantServer: "",
			wantPort:   -1,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotServer, gotPort, err := getDiscoveryServiceAddress(tt.args.ctx, tt.args.clnt, tt.args.namespace, tt.args.annotations)
			if (err != nil) != tt.wantErr {
				t.Errorf("envoySidecarConfig.GetDiscoveryServiceAddress() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if gotServer != tt.wantServer {
				t.Errorf("envoySidecarConfig.GetDiscoveryServiceAddress() got = %v, want %v", gotServer, tt.wantServer)
			}

			if gotPort != tt.wantPort {
				t.Errorf("envoySidecarConfig.GetDiscoveryServiceAddress() got1 = %v, want %v", gotPort, tt.wantPort)
			}
		})
	}
}

func Test_parseExtraLifecycleHooksAnnotation(t *testing.T) {
	type args struct {
		annotations map[string]string
	}

	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Parses a list of containers",
			args: args{
				annotations: map[string]string{
					"marin3r.3scale.net/shutdown-manager.extra-lifecycle-hooks": "container1,container2",
				},
			},
			want: []string{"container1", "container2"},
		},
		{
			name: "Returns empty slice if unset",
			args: args{
				annotations: map[string]string{},
			},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseExtraLifecycleHooksAnnotation(tt.args.annotations); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extra-lifecycle-hooks() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getContainerByName(t *testing.T) {
	type args struct {
		name       string
		containers []corev1.Container
	}

	tests := []struct {
		name    string
		args    args
		want1   corev1.Container
		want2   int
		wantErr bool
	}{
		{
			name: "Returns the container",
			args: args{
				name: "c2",
				containers: []corev1.Container{
					{Name: "c1"}, {Name: "c2"},
				},
			},
			want1:   corev1.Container{Name: "c2"},
			want2:   1,
			wantErr: false,
		},
		{
			name: "Container not found",
			args: args{
				name: "c3",
				containers: []corev1.Container{
					{Name: "c1"}, {Name: "c2"},
				},
			},
			want1:   corev1.Container{},
			want2:   -1,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got1, _, err := getContainerByName(tt.args.name, tt.args.containers)
			if (err != nil) != tt.wantErr {
				t.Errorf("getContainerByName() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("getContainerByName() = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_envoySidecarConfig_addExtraLifecycleHooks(t *testing.T) {
	type fields struct {
		generator envoy_container.ContainerConfig
	}

	type args struct {
		containers  []corev1.Container
		annotations map[string]string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []corev1.Container
		wantErr bool
	}{
		{
			name: "Injects the lifecycle hook into containers that match provided names",
			fields: fields{
				generator: envoy_container.ContainerConfig{
					ShutdownManagerPort: int32(defaults.ShtdnMgrDefaultServerPort),
				},
			},
			args: args{
				containers: []corev1.Container{
					{Name: "c1"}, {Name: "c2"},
				},
				annotations: map[string]string{
					"marin3r.3scale.net/shutdown-manager.extra-lifecycle-hooks": "c1",
				},
			},
			want: []corev1.Container{
				{Name: "c1",
					Lifecycle: &corev1.Lifecycle{
						PreStop: &corev1.LifecycleHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Path:   shutdownmanager.DrainEndpoint,
								Port:   intstr.FromInt(int(defaults.ShtdnMgrDefaultServerPort)),
								Scheme: corev1.URISchemeHTTP,
							},
						},
					}},
				{Name: "c2"},
			},
			wantErr: false,
		},
		{
			name: "Specified container does not exist",
			fields: fields{
				generator: envoy_container.ContainerConfig{
					ShutdownManagerPort: int32(defaults.ShtdnMgrDefaultServerPort),
				},
			},
			args: args{
				containers: []corev1.Container{
					{Name: "c1"}, {Name: "c2"},
				},
				annotations: map[string]string{
					"marin3r.3scale.net/shutdown-manager.extra-lifecycle-hooks": "zzzz",
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			esc := &envoySidecarConfig{
				generator: tt.fields.generator,
			}

			got, err := esc.addExtraLifecycleHooks(tt.args.containers, tt.args.annotations)
			if (err != nil) != tt.wantErr {
				t.Errorf("envoySidecarConfig.addExtraLifecycleHooks() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("envoySidecarConfig.addExtraLifecycleHooks() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getInt64Param(t *testing.T) {
	type args struct {
		key         string
		annotations map[string]string
	}

	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "Returns an int",
			args: args{
				key: "shutdown-manager.drain-time",
				annotations: map[string]string{
					"marin3r.3scale.net/shutdown-manager.drain-time": "100",
				},
			},
			want: 100,
		},
		{
			name: "Returns default if not present",
			args: args{
				key:         "shutdown-manager.drain-time",
				annotations: map[string]string{},
			},
			want: defaults.GracefulShutdownTimeoutSeconds,
		},
		{
			name: "Returns default if error",
			args: args{
				key: "shutdown-manager.drain-time",
				annotations: map[string]string{
					"marin3r.3scale.net/shutdown-manager.drain-time": "xxx",
				},
			},
			want: defaults.GracefulShutdownTimeoutSeconds,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getInt64Param(tt.args.key, tt.args.annotations); got != tt.want {
				t.Errorf("getInt64Param() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getDrainStrategy(t *testing.T) {
	type args struct {
		annotations map[string]string
	}

	tests := []struct {
		name string
		args args
		want defaults.DrainStrategy
	}{
		{
			name: "Returns drain strategy",
			args: args{
				annotations: map[string]string{
					"marin3r.3scale.net/shutdown-manager.drain-strategy": "immediate",
				},
			},
			want: defaults.DrainStrategyImmediate,
		},
		{
			name: "Returns default value if not annotation present",
			args: args{
				annotations: map[string]string{},
			},
			want: defaults.DrainStrategyGradual,
		},
		{
			name: "Returns default value if user provides wrong value",
			args: args{
				annotations: map[string]string{
					"marin3r.3scale.net/shutdown-manager.drain-strategy": "xxx",
				},
			},
			want: defaults.DrainStrategyGradual,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getDrainStrategy(tt.args.annotations); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getDrainStrategy() = %v, want %v", got, tt.want)
			}
		})
	}
}
