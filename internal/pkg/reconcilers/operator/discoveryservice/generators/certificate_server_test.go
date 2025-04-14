package generators

import (
	"testing"
	"time"

	reconcilerutil "github.com/3scale-sre/basereconciler/util"
	operatorv1alpha1 "github.com/3scale-sre/marin3r/api/operator.marin3r/v1alpha1"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGeneratorOptions_ServerCertificate(t *testing.T) {
	type args struct {
		hash string
	}
	tests := []struct {
		name string
		opts GeneratorOptions
		args args
		want *operatorv1alpha1.DiscoveryServiceCertificate
	}{
		{"Generates DiscoveryServiceCertificate for the server certificate",
			GeneratorOptions{
				InstanceName:                      "test",
				Namespace:                         "default",
				RootCertificateNamePrefix:         "ca-cert",
				RootCertificateCommonNamePrefix:   "test",
				RootCertificateDuration:           time.Duration(10 * time.Second), // 3 years
				ServerCertificateNamePrefix:       "server-cert",
				ServerCertificateCommonNamePrefix: "test",
				ServerCertificateDuration:         time.Duration(10 * time.Second), // 90 days,
				ClientCertificateDuration:         time.Duration(10 * time.Second),
				XdsServerPort:                     1000,
				MetricsServerPort:                 1001,
				ServiceType:                       operatorv1alpha1.ClusterIPType,
				DeploymentImage:                   "test:latest",
				DeploymentResources:               corev1.ResourceRequirements{},
				Debug:                             true,
			},
			args{hash: "hash"},
			&operatorv1alpha1.DiscoveryServiceCertificate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "server-cert-test",
					Namespace: "default",
					Labels: map[string]string{
						"app.kubernetes.io/name":       "marin3r",
						"app.kubernetes.io/managed-by": "marin3r-operator",
						"app.kubernetes.io/component":  "discovery-service",
						"app.kubernetes.io/instance":   "test",
					},
				},
				Spec: operatorv1alpha1.DiscoveryServiceCertificateSpec{
					CommonName:          "test-test",
					IsServerCertificate: reconcilerutil.Pointer(true),
					ValidFor:            int64(10),
					Signer: operatorv1alpha1.DiscoveryServiceCertificateSigner{
						CASigned: &operatorv1alpha1.CASignedConfig{
							SecretRef: corev1.SecretReference{
								Name:      "ca-cert-test",
								Namespace: "default",
							}}},
					SecretRef: corev1.SecretReference{
						Name:      "server-cert-test",
						Namespace: "default",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.opts.ServerCertificate(), tt.want); len(diff) > 0 {
				t.Errorf("GeneratorOptions.ServerCertificate() DIFF:\n %v", diff)
			}
		})
	}
}
