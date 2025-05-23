package generators

import (
	"testing"

	operatorv1alpha1 "github.com/3scale-sre/marin3r/api/operator.marin3r/v1alpha1"
	"github.com/google/go-cmp/cmp"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestGeneratorOptions_PDB(t *testing.T) {
	tests := []struct {
		name string
		opts GeneratorOptions
		want *policyv1.PodDisruptionBudget
	}{
		{
			name: "Generate an HPA",
			opts: GeneratorOptions{
				InstanceName: "instance",
				Namespace:    "default",
				PodDisruptionBudget: operatorv1alpha1.PodDisruptionBudgetSpec{
					MinAvailable: &intstr.IntOrString{Type: intstr.Int, IntVal: 1},
				},
			},
			want: &policyv1.PodDisruptionBudget{
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
				Spec: policyv1.PodDisruptionBudgetSpec{
					MinAvailable: &intstr.IntOrString{Type: intstr.Int, IntVal: 1},
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app.kubernetes.io/name":       "marin3r",
							"app.kubernetes.io/managed-by": "marin3r-operator",
							"app.kubernetes.io/component":  "envoy-deployment",
							"app.kubernetes.io/instance":   "instance",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.opts.PDB(), tt.want); len(diff) > 0 {
				t.Errorf("GeneratorOptions.PDB() DIFF:\n %v", diff)
			}
		})
	}
}
