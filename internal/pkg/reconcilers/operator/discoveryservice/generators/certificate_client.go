package generators

import (
	operatorv1alpha1 "github.com/3scale-sre/marin3r/api/operator.marin3r/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (cfg *GeneratorOptions) ClientCertificate() *operatorv1alpha1.DiscoveryServiceCertificate {
	return &operatorv1alpha1.DiscoveryServiceCertificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfg.ClientCertName(),
			Namespace: cfg.Namespace,
			Labels:    cfg.labels(),
		},
		Spec: operatorv1alpha1.DiscoveryServiceCertificateSpec{
			CommonName: cfg.ClientCertName(),
			ValidFor:   int64(cfg.ClientCertificateDuration.Seconds()),
			Signer: operatorv1alpha1.DiscoveryServiceCertificateSigner{
				CASigned: &operatorv1alpha1.CASignedConfig{
					SecretRef: corev1.SecretReference{
						Name:      cfg.RootCertName(),
						Namespace: cfg.Namespace,
					}},
			},
			SecretRef: corev1.SecretReference{
				Name: cfg.ClientCertName(),
			},
		},
	}
}
