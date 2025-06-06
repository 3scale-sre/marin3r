package generators

import (
	"testing"
	"time"

	operatorv1alpha1 "github.com/3scale-sre/marin3r/api/operator.marin3r/v1alpha1"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestGeneratorOptions_Service(t *testing.T) {
	type args struct {
		hash string
	}

	tests := []struct {
		name string
		opts GeneratorOptions
		args args
		want *corev1.Service
	}{
		{"Generates a Service (ClusterIP)",
			GeneratorOptions{
				InstanceName:                      "test",
				Namespace:                         "default",
				RootCertificateNamePrefix:         "ca-cert",
				RootCertificateCommonNamePrefix:   "test",
				RootCertificateDuration:           10 * time.Second, // 3 years
				ServerCertificateNamePrefix:       "server-cert",
				ServerCertificateCommonNamePrefix: "test",
				ServerCertificateDuration:         10 * time.Second, // 90 days,
				ClientCertificateDuration:         10 * time.Second,
				XdsServerPort:                     1000,
				MetricsServerPort:                 1001,
				ServiceType:                       operatorv1alpha1.ClusterIPType,
				DeploymentImage:                   "test:latest",
				DeploymentResources:               corev1.ResourceRequirements{},
				Debug:                             true,
			},
			args{hash: "hash"},
			&corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "marin3r-test",
					Namespace: "default",
					Labels: map[string]string{
						"app.kubernetes.io/name":       "marin3r",
						"app.kubernetes.io/managed-by": "marin3r-operator",
						"app.kubernetes.io/component":  "discovery-service",
						"app.kubernetes.io/instance":   "test",
					},
				},
				Spec: corev1.ServiceSpec{
					Type:      corev1.ServiceType(operatorv1alpha1.ClusterIPType),
					ClusterIP: "",
					Selector: map[string]string{
						"app.kubernetes.io/name":       "marin3r",
						"app.kubernetes.io/managed-by": "marin3r-operator",
						"app.kubernetes.io/component":  "discovery-service",
						"app.kubernetes.io/instance":   "test",
					},
					SessionAffinity: corev1.ServiceAffinityNone,
					Ports: []corev1.ServicePort{
						{
							Name:       "discovery",
							Port:       1000,
							Protocol:   corev1.ProtocolTCP,
							TargetPort: intstr.FromString("discovery"),
						},
						{
							Name:       "metrics",
							Port:       1001,
							Protocol:   corev1.ProtocolTCP,
							TargetPort: intstr.FromString("metrics"),
						},
					},
				},
			},
		},
		{"Generates a Service (Headless)",
			GeneratorOptions{
				InstanceName:                      "test",
				Namespace:                         "default",
				RootCertificateNamePrefix:         "ca-cert",
				RootCertificateCommonNamePrefix:   "test",
				RootCertificateDuration:           10 * time.Second, // 3 years
				ServerCertificateNamePrefix:       "server-cert",
				ServerCertificateCommonNamePrefix: "test",
				ServerCertificateDuration:         10 * time.Second, // 90 days,
				ClientCertificateDuration:         10 * time.Second,
				XdsServerPort:                     1000,
				MetricsServerPort:                 1001,
				ServiceType:                       operatorv1alpha1.HeadlessType,
				DeploymentImage:                   "test:latest",
				DeploymentResources:               corev1.ResourceRequirements{},
				Debug:                             true,
			},
			args{hash: "hash"},
			&corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "marin3r-test",
					Namespace: "default",
					Labels: map[string]string{
						"app.kubernetes.io/name":       "marin3r",
						"app.kubernetes.io/managed-by": "marin3r-operator",
						"app.kubernetes.io/component":  "discovery-service",
						"app.kubernetes.io/instance":   "test",
					},
				},
				Spec: corev1.ServiceSpec{
					Type:      corev1.ServiceType(operatorv1alpha1.ClusterIPType),
					ClusterIP: "None",
					Selector: map[string]string{
						"app.kubernetes.io/name":       "marin3r",
						"app.kubernetes.io/managed-by": "marin3r-operator",
						"app.kubernetes.io/component":  "discovery-service",
						"app.kubernetes.io/instance":   "test",
					},
					SessionAffinity: corev1.ServiceAffinityNone,
					Ports: []corev1.ServicePort{
						{
							Name:       "discovery",
							Port:       1000,
							Protocol:   corev1.ProtocolTCP,
							TargetPort: intstr.FromString("discovery"),
						},
						{
							Name:       "metrics",
							Port:       1001,
							Protocol:   corev1.ProtocolTCP,
							TargetPort: intstr.FromString("metrics"),
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.opts.Service(), tt.want); len(diff) > 0 {
				t.Errorf("GeneratorOptions.Service() DIFF:\n %v", diff)
			}
		})
	}
}
