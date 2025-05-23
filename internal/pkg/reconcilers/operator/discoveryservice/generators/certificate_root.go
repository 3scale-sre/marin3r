package generators

import (
	"fmt"

	operatorv1alpha1 "github.com/3scale-sre/marin3r/api/operator.marin3r/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func (cfg *GeneratorOptions) RootCertificationAuthority() *operatorv1alpha1.DiscoveryServiceCertificate {
	return &operatorv1alpha1.DiscoveryServiceCertificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfg.RootCertName(),
			Namespace: cfg.Namespace,
			Labels:    cfg.labels(),
		},
		Spec: operatorv1alpha1.DiscoveryServiceCertificateSpec{
			CommonName: fmt.Sprintf("%s-%s", cfg.RootCertificateCommonNamePrefix, cfg.InstanceName),
			IsCA:       ptr.To(true),
			ValidFor:   int64(cfg.RootCertificateDuration.Seconds()),
			Signer: operatorv1alpha1.DiscoveryServiceCertificateSigner{
				SelfSigned: &operatorv1alpha1.SelfSignedConfig{},
			},
			SecretRef: corev1.SecretReference{
				Name:      cfg.RootCertName(),
				Namespace: cfg.Namespace,
			},
			CertificateRenewalConfig: &operatorv1alpha1.CertificateRenewalConfig{
				Enabled: true,
			},
		},
	}
}
