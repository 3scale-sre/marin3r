package container

import (
	"fmt"
	"strconv"

	"github.com/3scale-sre/marin3r/api/envoy/defaults"
	operatorv1alpha1 "github.com/3scale-sre/marin3r/api/operator.marin3r/v1alpha1"
	"github.com/3scale-sre/marin3r/internal/pkg/envoy/container/shutdownmanager"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

type ContainerConfig struct {
	// Envoy container configuration
	Name               string
	Image              string
	ConfigBasePath     string
	ConfigFileName     string
	ConfigVolume       string
	TLSBasePath        string
	TLSVolume          string
	NodeID             string
	ClusterID          string
	ClientCertSecret   string
	ExtraArgs          []string
	Resources          corev1.ResourceRequirements
	AdminBindAddress   string
	AdminPort          int32
	AdminAccessLogPath string
	Ports              []corev1.ContainerPort
	LivenessProbe      operatorv1alpha1.ProbeSpec
	ReadinessProbe     operatorv1alpha1.ProbeSpec

	// Init manager container configuration
	InitManagerImage string
	XdssHost         string
	XdssPort         int
	APIVersion       string

	// Shutdown manager container configuration
	ShutdownManagerEnabled       bool
	ShutdownManagerPort          int32
	ShutdownManagerImage         string
	ShutdownManagerDrainSeconds  int64
	ShutdownManagerDrainStrategy defaults.DrainStrategy
}

func (cc *ContainerConfig) Containers() []corev1.Container {
	containers := []corev1.Container{{
		Name:    cc.Name,
		Image:   cc.Image,
		Command: []string{"envoy"},
		Args: func() []string {
			args := []string{"-c",
				fmt.Sprintf("%s/%s", cc.ConfigBasePath, cc.ConfigFileName),
				"--service-node",
				cc.NodeID,
				"--service-cluster",
				cc.ClusterID,
			}
			if cc.ShutdownManagerEnabled {
				args = append(args,
					"--drain-time-s", strconv.FormatInt(cc.ShutdownManagerDrainSeconds, 10),
					"--drain-strategy", string(cc.ShutdownManagerDrainStrategy),
				)
			}
			if len(cc.ExtraArgs) > 0 {
				args = append(args, cc.ExtraArgs...)
			}

			return args
		}(),
		Resources: cc.Resources,
		Ports: append(cc.Ports, corev1.ContainerPort{
			Name:          "admin",
			ContainerPort: cc.AdminPort,
			Protocol:      corev1.ProtocolTCP,
		}),
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      cc.TLSVolume,
				ReadOnly:  true,
				MountPath: cc.TLSBasePath,
			},
			{
				Name:      cc.ConfigVolume,
				ReadOnly:  true,
				MountPath: cc.ConfigBasePath,
			},
		},
		LivenessProbe: &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   "/ready",
					Port:   intstr.IntOrString{IntVal: cc.AdminPort},
					Scheme: corev1.URISchemeHTTP,
				},
			},
			InitialDelaySeconds: cc.LivenessProbe.InitialDelaySeconds,
			TimeoutSeconds:      cc.LivenessProbe.TimeoutSeconds,
			PeriodSeconds:       cc.LivenessProbe.PeriodSeconds,
			SuccessThreshold:    cc.LivenessProbe.SuccessThreshold,
			FailureThreshold:    cc.LivenessProbe.FailureThreshold,
		},
		ReadinessProbe: &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   "/ready",
					Port:   intstr.IntOrString{IntVal: cc.AdminPort},
					Scheme: corev1.URISchemeHTTP,
				},
			},
			InitialDelaySeconds: cc.ReadinessProbe.InitialDelaySeconds,
			TimeoutSeconds:      cc.ReadinessProbe.TimeoutSeconds,
			PeriodSeconds:       cc.ReadinessProbe.PeriodSeconds,
			SuccessThreshold:    cc.ReadinessProbe.SuccessThreshold,
			FailureThreshold:    cc.ReadinessProbe.FailureThreshold,
		},
		TerminationMessagePath:   corev1.TerminationMessagePathDefault,
		TerminationMessagePolicy: corev1.TerminationMessageReadFile,
		ImagePullPolicy:          corev1.PullIfNotPresent,
	}}

	if cc.ShutdownManagerEnabled {
		containers = append(containers, corev1.Container{
			Name:  "envoy-shtdn-mgr",
			Image: cc.ShutdownManagerImage,
			Args: []string{
				"shutdown-manager",
				"--port",
				strconv.Itoa(int(cc.ShutdownManagerPort)),
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
						Port:   intstr.FromInt(int(cc.ShutdownManagerPort)),
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
						Port:   intstr.FromInt(int(cc.ShutdownManagerPort)),
						Scheme: corev1.URISchemeHTTP,
					},
				},
			},
			TerminationMessagePath:   corev1.TerminationMessagePathDefault,
			TerminationMessagePolicy: corev1.TerminationMessageReadFile,
			ImagePullPolicy:          corev1.PullIfNotPresent,
		})

		containers[0].Lifecycle = &corev1.Lifecycle{
			PreStop: &corev1.LifecycleHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   shutdownmanager.ShutdownEndpoint,
					Port:   intstr.FromInt(int(cc.ShutdownManagerPort)),
					Scheme: corev1.URISchemeHTTP,
				},
			},
		}
	}

	return containers
}

func (cc *ContainerConfig) InitContainers() []corev1.Container {
	containers := []corev1.Container{{
		Name:  "envoy-init-mgr",
		Image: cc.InitManagerImage,
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
			"--admin-access-log-path", cc.AdminAccessLogPath,
			"--admin-bind-address", fmt.Sprintf("%s:%d", cc.AdminBindAddress, cc.AdminPort),
			"--api-version", cc.APIVersion,
			"--client-certificate-path", cc.TLSBasePath,
			"--config-file", fmt.Sprintf("%s/%s", cc.ConfigBasePath, cc.ConfigFileName),
			"--resources-path", cc.ConfigBasePath,
			"--rtds-resource-name", defaults.InitMgrRtdsLayerResourceName,
			"--xdss-host", cc.XdssHost,
			"--xdss-port", strconv.Itoa(cc.XdssPort),
			"--envoy-image", cc.Image,
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      cc.ConfigVolume,
				ReadOnly:  false,
				MountPath: cc.ConfigBasePath,
			},
		},
		ImagePullPolicy:          corev1.PullIfNotPresent,
		TerminationMessagePath:   corev1.TerminationMessagePathDefault,
		TerminationMessagePolicy: corev1.TerminationMessageReadFile,
	}}

	return containers
}

func (cc *ContainerConfig) Volumes() []corev1.Volume {
	return []corev1.Volume{
		{
			Name: cc.TLSVolume,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  cc.ClientCertSecret,
					DefaultMode: ptr.To(int32(420)),
				},
			},
		},
		{
			Name: cc.ConfigVolume,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}
}
