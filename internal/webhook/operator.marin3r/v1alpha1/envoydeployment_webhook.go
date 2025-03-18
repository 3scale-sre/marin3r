/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"context"

	operatorv1alpha1 "github.com/3scale-ops/marin3r/api/operator.marin3r/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// nolint:unused
// log is for logging in this package.
var envoydeploymentlog = logf.Log.WithName("envoydeployment-resource")

// SetupEnvoyDeploymentWebhookWithManager registers the webhook for EnvoyDeployment in the manager.
func SetupEnvoyDeploymentWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&operatorv1alpha1.EnvoyDeployment{}).
		WithValidator(&EnvoyDeploymentCustomValidator{}).
		Complete()
}

// +kubebuilder:webhook:path=/validate-operator-marin3r-3scale-net-v1alpha1-envoydeployment,mutating=false,failurePolicy=fail,sideEffects=None,groups=operator.marin3r.3scale.net,resources=envoydeployments,verbs=create;update,versions=v1alpha1,name=envoydeployment.operator.marin3r.3scale.net,admissionReviewVersions=v1
type EnvoyDeploymentCustomValidator struct{}

var _ webhook.CustomValidator = &EnvoyDeploymentCustomValidator{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (validator *EnvoyDeploymentCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ed := obj.(*operatorv1alpha1.EnvoyDeployment)
	envoydeploymentlog.V(1).Info("validate create", "name", ed.Name)
	return nil, validate(ed)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (validator *EnvoyDeploymentCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	ed := newObj.(*operatorv1alpha1.EnvoyDeployment)
	envoydeploymentlog.V(1).Info("validate update", "name", ed.Name)
	return nil, validate(ed)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (validator *EnvoyDeploymentCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

// Validate checks that the spec of the EnvoyDeployment resource is correct
func validate(ed *operatorv1alpha1.EnvoyDeployment) error {

	if ed.Spec.Replicas != nil {
		if err := ed.Spec.Replicas.Validate(); err != nil {
			return err
		}
	}

	if ed.Spec.PodDisruptionBudget != nil {
		if err := ed.Spec.PodDisruptionBudget.Validate(); err != nil {
			return err
		}
	}

	return nil
}
