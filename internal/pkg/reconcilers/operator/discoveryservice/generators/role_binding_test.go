package generators

import (
	"testing"
	"time"

	operatorv1alpha1 "github.com/3scale-sre/marin3r/api/operator.marin3r/v1alpha1"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGeneratorOptions_RoleBinding(t *testing.T) {
	type args struct {
		hash string
	}

	tests := []struct {
		name string
		opts GeneratorOptions
		args args
		want *rbacv1.RoleBinding
	}{
		{"Generates a RoleBinding",
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
			&rbacv1.RoleBinding{
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
				RoleRef: rbacv1.RoleRef{
					APIGroup: rbacv1.SchemeGroupVersion.Group,
					Kind:     "Role",
					Name:     "marin3r-test",
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      rbacv1.ServiceAccountKind,
						Name:      "marin3r-test",
						Namespace: "default",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.opts.RoleBinding(), tt.want); len(diff) > 0 {
				t.Errorf("GeneratorOptions.RoleBinding() DIFF:\n %v", diff)
			}
		})
	}
}
