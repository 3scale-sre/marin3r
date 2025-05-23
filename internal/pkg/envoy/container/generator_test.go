package container

import (
	"strconv"
	"testing"

	"github.com/3scale-sre/marin3r/api/envoy/defaults"
	operatorv1alpha1 "github.com/3scale-sre/marin3r/api/operator.marin3r/v1alpha1"
	"github.com/3scale-sre/marin3r/internal/pkg/envoy/container/shutdownmanager"
	"github.com/go-test/deep"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

func TestContainerConfig_Containers(t *testing.T) {
	tests := []struct {
		name string
		cc   ContainerConfig
		want []corev1.Container
	}{
		{
			name: "Generates an Envoy container for the given config",
			cc: ContainerConfig{
				Name:             "envoy",
				Image:            "envoy:test",
				ConfigBasePath:   "/config",
				ConfigFileName:   "config.json",
				ConfigVolume:     "config",
				TLSBasePath:      "/tls",
				TLSVolume:        "tls",
				NodeID:           "test-id",
				ClusterID:        "test-id",
				ClientCertSecret: "client-secret",
				ExtraArgs:        []string{"--some-arg", "some-value"},
				Resources:        corev1.ResourceRequirements{},
				AdminPort:        5000,
				Ports: []corev1.ContainerPort{
					{
						Name:          "udp",
						ContainerPort: 6000,
						Protocol:      corev1.Protocol("UDP"),
					},
					{
						Name:          "https",
						ContainerPort: 8443,
					},
				},
				LivenessProbe: operatorv1alpha1.ProbeSpec{
					InitialDelaySeconds: 1,
					TimeoutSeconds:      1,
					PeriodSeconds:       1,
					SuccessThreshold:    1,
					FailureThreshold:    1,
				},
				ReadinessProbe: operatorv1alpha1.ProbeSpec{
					InitialDelaySeconds: 1,
					TimeoutSeconds:      1,
					PeriodSeconds:       1,
					SuccessThreshold:    1,
					FailureThreshold:    1,
				},
			},
			want: []corev1.Container{{
				Name:    "envoy",
				Image:   "envoy:test",
				Command: []string{"envoy"},
				Args: []string{
					"-c",
					"/config/config.json",
					"--service-node",
					"test-id",
					"--service-cluster",
					"test-id",
					"--some-arg",
					"some-value",
				},
				Ports: []corev1.ContainerPort{
					{
						Name:          "udp",
						ContainerPort: 6000,
						Protocol:      corev1.Protocol("UDP"),
					},
					{
						Name:          "https",
						ContainerPort: 8443,
					},
					{
						Name:          "admin",
						ContainerPort: 5000,
						Protocol:      corev1.ProtocolTCP,
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "tls",
						ReadOnly:  true,
						MountPath: "/tls",
					},
					{
						Name:      "config",
						ReadOnly:  true,
						MountPath: "/config",
					},
				},
				LivenessProbe: &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Path:   "/ready",
							Port:   intstr.IntOrString{IntVal: 5000},
							Scheme: corev1.URISchemeHTTP,
						},
					},
					InitialDelaySeconds: 1,
					TimeoutSeconds:      1,
					PeriodSeconds:       1,
					SuccessThreshold:    1,
					FailureThreshold:    1,
				},
				ReadinessProbe: &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Path:   "/ready",
							Port:   intstr.IntOrString{IntVal: 5000},
							Scheme: corev1.URISchemeHTTP,
						},
					},
					InitialDelaySeconds: 1,
					TimeoutSeconds:      1,
					PeriodSeconds:       1,
					SuccessThreshold:    1,
					FailureThreshold:    1,
				},
				TerminationMessagePath:   corev1.TerminationMessagePathDefault,
				TerminationMessagePolicy: corev1.TerminationMessageReadFile,
				ImagePullPolicy:          corev1.PullIfNotPresent,
			}},
		},
		{
			name: "Generates containers for the given config (shtdnmgr enabled)",
			cc: ContainerConfig{
				Name:             "envoy",
				Image:            "envoy:test",
				ConfigBasePath:   "/config",
				ConfigFileName:   "config.json",
				ConfigVolume:     "config",
				TLSBasePath:      "/tls",
				TLSVolume:        "tls",
				NodeID:           "test-id",
				ClusterID:        "test-id",
				ClientCertSecret: "client-secret",
				ExtraArgs:        []string{"--some-arg", "some-value"},
				Resources:        corev1.ResourceRequirements{},
				AdminPort:        5000,
				Ports: []corev1.ContainerPort{
					{
						Name:          "udp",
						ContainerPort: 6000,
						Protocol:      corev1.Protocol("UDP"),
					},
					{
						Name:          "https",
						ContainerPort: 8443,
					},
				},
				LivenessProbe: operatorv1alpha1.ProbeSpec{
					InitialDelaySeconds: 1,
					TimeoutSeconds:      1,
					PeriodSeconds:       1,
					SuccessThreshold:    1,
					FailureThreshold:    1,
				},
				ReadinessProbe: operatorv1alpha1.ProbeSpec{
					InitialDelaySeconds: 1,
					TimeoutSeconds:      1,
					PeriodSeconds:       1,
					SuccessThreshold:    1,
					FailureThreshold:    1,
				},
				ShutdownManagerEnabled:       true,
				ShutdownManagerPort:          30000,
				ShutdownManagerImage:         "image:shtdnmgr",
				ShutdownManagerDrainSeconds:  360,
				ShutdownManagerDrainStrategy: defaults.DrainStrategyGradual,
			},
			want: []corev1.Container{
				{
					Name:    "envoy",
					Image:   "envoy:test",
					Command: []string{"envoy"},
					Args: []string{
						"-c",
						"/config/config.json",
						"--service-node",
						"test-id",
						"--service-cluster",
						"test-id",
						"--drain-time-s",
						"360",
						"--drain-strategy",
						"gradual",
						"--some-arg",
						"some-value",
					},
					Ports: []corev1.ContainerPort{
						{
							Name:          "udp",
							ContainerPort: 6000,
							Protocol:      corev1.Protocol("UDP"),
						},
						{
							Name:          "https",
							ContainerPort: 8443,
						},
						{
							Name:          "admin",
							ContainerPort: 5000,
							Protocol:      corev1.ProtocolTCP,
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "tls",
							ReadOnly:  true,
							MountPath: "/tls",
						},
						{
							Name:      "config",
							ReadOnly:  true,
							MountPath: "/config",
						},
					},
					LivenessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Path:   "/ready",
								Port:   intstr.IntOrString{IntVal: 5000},
								Scheme: corev1.URISchemeHTTP,
							},
						},
						InitialDelaySeconds: 1,
						TimeoutSeconds:      1,
						PeriodSeconds:       1,
						SuccessThreshold:    1,
						FailureThreshold:    1,
					},
					ReadinessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Path:   "/ready",
								Port:   intstr.IntOrString{IntVal: 5000},
								Scheme: corev1.URISchemeHTTP,
							},
						},
						InitialDelaySeconds: 1,
						TimeoutSeconds:      1,
						PeriodSeconds:       1,
						SuccessThreshold:    1,
						FailureThreshold:    1,
					},
					TerminationMessagePath:   corev1.TerminationMessagePathDefault,
					TerminationMessagePolicy: corev1.TerminationMessageReadFile,
					ImagePullPolicy:          corev1.PullIfNotPresent,
					Lifecycle: &corev1.Lifecycle{
						PreStop: &corev1.LifecycleHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Path:   shutdownmanager.ShutdownEndpoint,
								Port:   intstr.FromInt(30000),
								Scheme: corev1.URISchemeHTTP,
							},
						},
					},
				},
				{
					Name:  "envoy-shtdn-mgr",
					Image: "image:shtdnmgr",
					Args: []string{
						"shutdown-manager",
						"--port",
						strconv.Itoa(30000),
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse(defaults.ShtdnMgrDefaultCPURequests),
							corev1.ResourceMemory: resource.MustParse(defaults.ShtdnMgrDefaultMemoryRequests),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse(defaults.ShtdnMgrDefaultCPULimits),
							corev1.ResourceMemory: resource.MustParse(defaults.ShtdnMgrDefaultMemoryLimits),
						},
					},
					LivenessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Path:   shutdownmanager.HealthEndpoint,
								Port:   intstr.FromInt(30000),
								Scheme: corev1.URISchemeHTTP,
							},
						},
						InitialDelaySeconds: 3,
						PeriodSeconds:       10,
						TimeoutSeconds:      1,
						SuccessThreshold:    1,
						FailureThreshold:    3,
					},
					Lifecycle: &corev1.Lifecycle{
						PreStop: &corev1.LifecycleHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Path:   shutdownmanager.DrainEndpoint,
								Port:   intstr.FromInt(30000),
								Scheme: corev1.URISchemeHTTP,
							},
						},
					},
					TerminationMessagePath:   corev1.TerminationMessagePathDefault,
					TerminationMessagePolicy: corev1.TerminationMessageReadFile,
					ImagePullPolicy:          corev1.PullIfNotPresent,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := deep.Equal(tt.cc.Containers(), tt.want); len(diff) > 0 {
				t.Errorf("ContainerConfig.Container() = diff %v", diff)
			}
		})
	}
}

func TestContainerConfig_Volumes(t *testing.T) {
	tests := []struct {
		name string
		cc   ContainerConfig
		want []corev1.Volume
	}{
		{
			name: "Generates required volumes for an Envoy container with the given config",
			cc: ContainerConfig{
				ConfigVolume:     "config",
				TLSVolume:        "tls",
				ClientCertSecret: "client-secret",
			},
			want: []corev1.Volume{
				{
					Name: "tls",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName:  "client-secret",
							DefaultMode: ptr.To(int32(420)),
						},
					},
				},
				{
					Name: "config",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := deep.Equal(tt.cc.Volumes(), tt.want); len(diff) > 0 {
				t.Errorf("ContainerConfig.Volumes() = diff %v", diff)
			}
		})
	}
}

func TestContainerConfig_InitContainers(t *testing.T) {
	tests := []struct {
		name string
		cc   ContainerConfig
		want []corev1.Container
	}{
		{
			name: "Generates init manager init-container for the given config",
			cc: ContainerConfig{
				Image:              "envoy:test",
				ConfigBasePath:     "/config",
				ConfigFileName:     "config.json",
				ConfigVolume:       "config",
				TLSBasePath:        "/tls",
				NodeID:             "test-id",
				ClusterID:          "test-id",
				ClientCertSecret:   "client-secret",
				AdminAccessLogPath: "/dev/stdout",
				AdminBindAddress:   "127.0.0.1",
				AdminPort:          5000,
				XdssHost:           "discovery-service.com",
				XdssPort:           30000,
				APIVersion:         "v3",
				InitManagerImage:   "init-manager:test",
			},
			want: []corev1.Container{{
				Name:  "envoy-init-mgr",
				Image: "init-manager:test",
				Env: []corev1.EnvVar{
					{
						Name: "POD_NAME",
						ValueFrom: &corev1.EnvVarSource{
							FieldRef: &corev1.ObjectFieldSelector{
								FieldPath:  "metadata.name",
								APIVersion: "v1",
							},
						},
					},
					{
						Name: "POD_NAMESPACE",
						ValueFrom: &corev1.EnvVarSource{
							FieldRef: &corev1.ObjectFieldSelector{
								FieldPath:  "metadata.namespace",
								APIVersion: "v1",
							},
						},
					},
					{
						Name: "HOST_NAME",
						ValueFrom: &corev1.EnvVarSource{
							FieldRef: &corev1.ObjectFieldSelector{
								FieldPath:  "spec.nodeName",
								APIVersion: "v1",
							},
						},
					},
				},
				Args: []string{
					"init-manager",
					"--admin-access-log-path", "/dev/stdout",
					"--admin-bind-address", "127.0.0.1:5000",
					"--api-version", "v3",
					"--client-certificate-path", "/tls",
					"--config-file", "/config/config.json",
					"--resources-path", "/config",
					"--rtds-resource-name", defaults.InitMgrRtdsLayerResourceName,
					"--xdss-host", "discovery-service.com",
					"--xdss-port", "30000",
					"--envoy-image", "envoy:test",
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "config",
						ReadOnly:  false,
						MountPath: "/config",
					},
				},
				ImagePullPolicy:          corev1.PullIfNotPresent,
				TerminationMessagePath:   corev1.TerminationMessagePathDefault,
				TerminationMessagePolicy: corev1.TerminationMessageReadFile,
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := deep.Equal(tt.cc.InitContainers(), tt.want); len(diff) > 0 {
				t.Errorf("ContainerConfig.InitContainers() = diff %v", diff)
			}
		})
	}
}
