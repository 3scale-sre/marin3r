package generators

import (
	"testing"
	"time"

	operatorv1alpha1 "github.com/3scale-sre/marin3r/api/operator.marin3r/v1alpha1"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestGeneratorOptions_ClientCertificate(t *testing.T) {
	tests := []struct {
		name string
		opts GeneratorOptions
		want client.Object
	}{
		{
			name: "Generates DSC resource",
			opts: GeneratorOptions{
				InstanceName:              "instance",
				Namespace:                 "default",
				ClientCertificateName:     "cert",
				ClientCertificateDuration: 20 * time.Second,
				SigningCertificateName:    "signing-cert",
			},
			want: &operatorv1alpha1.DiscoveryServiceCertificate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cert",
					Namespace: "default",
					Labels: map[string]string{
						"app.kubernetes.io/name":       "marin3r",
						"app.kubernetes.io/managed-by": "marin3r-operator",
						"app.kubernetes.io/component":  "envoy-deployment",
						"app.kubernetes.io/instance":   "instance",
					},
				},
				Spec: operatorv1alpha1.DiscoveryServiceCertificateSpec{
					CommonName: "cert",
					ValidFor:   int64((20 * time.Second).Seconds()),
					Signer: operatorv1alpha1.DiscoveryServiceCertificateSigner{
						CASigned: &operatorv1alpha1.CASignedConfig{
							SecretRef: corev1.SecretReference{
								Name:      "signing-cert",
								Namespace: "default",
							}},
					},
					SecretRef: corev1.SecretReference{
						Name: "cert",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.opts.ClientCertificate(), tt.want); len(diff) > 0 {
				t.Errorf("GeneratorOptions.ClientCertificate() DIFF:\n %v", diff)
			}
		})
	}
}
