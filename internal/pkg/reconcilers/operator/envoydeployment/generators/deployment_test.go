package generators

import (
	"fmt"
	"testing"
	"time"

	defaults "github.com/3scale-sre/marin3r/api/envoy/defaults"
	operatorv1alpha1 "github.com/3scale-sre/marin3r/api/operator.marin3r/v1alpha1"
	"github.com/google/go-cmp/cmp"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

func TestGeneratorOptions_Deployment(t *testing.T) {
	tests := []struct {
		name string
		opts GeneratorOptions
		want *appsv1.Deployment
	}{
		{
			name: "EnvoyDeployment's Deployment generation",
			opts: GeneratorOptions{
				InstanceName:              "instance",
				Namespace:                 "default",
				XdssAdress:                "example.com",
				XdssPort:                  10000,
				EnvoyAPIVersion:           "v3",
				EnvoyNodeID:               "test",
				EnvoyClusterID:            "test",
				ClientCertificateDuration: 20 * time.Second,
				DeploymentImage:           "test:latest",
				DeploymentResources:       corev1.ResourceRequirements{},
				ExposedPorts:              []operatorv1alpha1.ContainerPort{{Name: "port", Port: 8080}},
				AdminPort:                 9901,
				AdminAccessLogPath:        "/dev/null",
				Replicas:                  operatorv1alpha1.ReplicasSpec{Static: ptr.To(int32(1))},
				LivenessProbe:             operatorv1alpha1.ProbeSpec{InitialDelaySeconds: 30, TimeoutSeconds: 1, PeriodSeconds: 10, SuccessThreshold: 1, FailureThreshold: 10},
				ReadinessProbe:            operatorv1alpha1.ProbeSpec{InitialDelaySeconds: 15, TimeoutSeconds: 1, PeriodSeconds: 5, SuccessThreshold: 1, FailureThreshold: 1},
				InitManager:               &operatorv1alpha1.InitManager{Image: ptr.To("init-manager:latest")},
			},
			want: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "marin3r-envoydeployment-instance",
					Namespace: "default",
					Labels: map[string]string{
						"app.kubernetes.io/name":       "marin3r",
						"app.kubernetes.io/managed-by": "marin3r-operator",
						"app.kubernetes.io/component":  "envoy-deployment",
						"app.kubernetes.io/instance":   "instance",
					},
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: ptr.To(int32(1)),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app.kubernetes.io/name":       "marin3r",
							"app.kubernetes.io/managed-by": "marin3r-operator",
							"app.kubernetes.io/component":  "envoy-deployment",
							"app.kubernetes.io/instance":   "instance",
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							CreationTimestamp: metav1.Time{},
							Labels: map[string]string{
								"app.kubernetes.io/name":       "marin3r",
								"app.kubernetes.io/managed-by": "marin3r-operator",
								"app.kubernetes.io/component":  "envoy-deployment",
								"app.kubernetes.io/instance":   "instance",
							}},
						Spec: corev1.PodSpec{
							Volumes: []corev1.Volume{
								{
									Name: defaults.DeploymentTLSVolume,
									VolumeSource: corev1.VolumeSource{
										Secret: &corev1.SecretVolumeSource{
											SecretName:  defaults.DeploymentClientCertificate + "-instance",
											DefaultMode: ptr.To(int32(420)),
										},
									},
								},
								{
									Name: defaults.DeploymentConfigVolume,
									VolumeSource: corev1.VolumeSource{
										EmptyDir: &corev1.EmptyDirVolumeSource{},
									},
								},
							},
							InitContainers: []corev1.Container{{
								Name:  "envoy-init-mgr",
								Image: "init-manager:latest",
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
									"--admin-access-log-path", "/dev/null",
									"--admin-bind-address", "0.0.0.0:9901",
									"--api-version", "v3",
									"--client-certificate-path", defaults.EnvoyTLSBasePath,
									"--config-file", fmt.Sprintf("%s/%s", defaults.EnvoyConfigBasePath, defaults.EnvoyConfigFileName),
									"--resources-path", defaults.EnvoyConfigBasePath,
									"--rtds-resource-name", defaults.InitMgrRtdsLayerResourceName,
									"--xdss-host", "example.com",
									"--xdss-port", "10000",
									"--envoy-image", "test:latest",
								},
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      defaults.DeploymentConfigVolume,
										ReadOnly:  false,
										MountPath: defaults.EnvoyConfigBasePath,
									},
								},
								TerminationMessagePath:   corev1.TerminationMessagePathDefault,
								TerminationMessagePolicy: corev1.TerminationMessageReadFile,
								ImagePullPolicy:          corev1.PullIfNotPresent,
							}},
							Containers: []corev1.Container{
								{
									Name:    defaults.DeploymentContainerName,
									Image:   "test:latest",
									Command: []string{"envoy"},
									Args: []string{
										"-c",
										fmt.Sprintf("%s/%s", defaults.EnvoyConfigBasePath, defaults.EnvoyConfigFileName),
										"--service-node",
										"test",
										"--service-cluster",
										"test",
									},
									Resources: corev1.ResourceRequirements{},
									Ports: []corev1.ContainerPort{
										{
											Name:          "port",
											ContainerPort: int32(8080),
										},
										{
											Name:          "admin",
											ContainerPort: int32(9901),
											Protocol:      corev1.ProtocolTCP,
										},
									},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      defaults.DeploymentTLSVolume,
											ReadOnly:  true,
											MountPath: defaults.EnvoyTLSBasePath,
										},
										{
											Name:      defaults.DeploymentConfigVolume,
											ReadOnly:  true,
											MountPath: defaults.EnvoyConfigBasePath,
										},
									},
									LivenessProbe: &corev1.Probe{
										ProbeHandler: corev1.ProbeHandler{
											HTTPGet: &corev1.HTTPGetAction{
												Path:   "/ready",
												Port:   intstr.IntOrString{IntVal: 9901},
												Scheme: corev1.URISchemeHTTP,
											},
										},
										InitialDelaySeconds: 30,
										TimeoutSeconds:      1,
										PeriodSeconds:       10,
										SuccessThreshold:    1,
										FailureThreshold:    10,
									},
									ReadinessProbe: &corev1.Probe{
										ProbeHandler: corev1.ProbeHandler{
											HTTPGet: &corev1.HTTPGetAction{
												Path:   "/ready",
												Port:   intstr.IntOrString{IntVal: 9901},
												Scheme: corev1.URISchemeHTTP,
											},
										},
										InitialDelaySeconds: 15,
										TimeoutSeconds:      1,
										PeriodSeconds:       5,
										SuccessThreshold:    1,
										FailureThreshold:    1,
									},
									TerminationMessagePath:   corev1.TerminationMessagePathDefault,
									TerminationMessagePolicy: corev1.TerminationMessageReadFile,
									ImagePullPolicy:          corev1.PullIfNotPresent,
								},
							},
							TerminationGracePeriodSeconds: ptr.To(int64(corev1.DefaultTerminationGracePeriodSeconds)),
							ServiceAccountName:            "default",
							DeprecatedServiceAccount:      "default",
						},
					},
					Strategy: appsv1.DeploymentStrategy{
						Type: appsv1.RollingUpdateDeploymentStrategyType,
						RollingUpdate: &appsv1.RollingUpdateDeployment{
							MaxUnavailable: &intstr.IntOrString{
								Type:   intstr.String,
								StrVal: "25%",
							},
							MaxSurge: &intstr.IntOrString{
								Type:   intstr.String,
								StrVal: "25%",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.opts.Deployment(), tt.want); len(diff) > 0 {
				t.Errorf("GeneratorOptions.Deployment() DIFF:\n %v", diff)
			}
		})
	}
}
