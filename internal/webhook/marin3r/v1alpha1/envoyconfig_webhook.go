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
	"fmt"

	"github.com/3scale-sre/basereconciler/util"
	"github.com/3scale-sre/marin3r/api/envoy"
	envoy_resources "github.com/3scale-sre/marin3r/api/envoy/resources"
	envoy_serializer "github.com/3scale-sre/marin3r/api/envoy/serializer"
	marin3rv1alpha1 "github.com/3scale-sre/marin3r/api/marin3r/v1alpha1"
	errorutil "github.com/3scale-sre/marin3r/pkg/util/error"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// nolint:unused
// log is for logging in this package.
var envoyconfiglog = logf.Log.WithName("envoyconfig-resource")

// SetupEnvoyConfigWebhookWithManager registers the webhook for EnvoyConfig in the manager.
func SetupEnvoyConfigWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&marin3rv1alpha1.EnvoyConfig{}).
		WithValidator(&EnvoyConfigCustomValidator{}).
		Complete()
}

type EnvoyConfigCustomValidator struct{}

var _ webhook.CustomValidator = &EnvoyConfigCustomValidator{}

// +kubebuilder:webhook:path=/validate-marin3r-3scale-net-v1alpha1-envoyconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=marin3r.3scale.net,resources=envoyconfigs,verbs=create;update,versions=v1alpha1,name=envoyconfig.marin3r.3scale.net-v1alpha1,admissionReviewVersions=v1

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (validator *EnvoyConfigCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ec := obj.(*marin3rv1alpha1.EnvoyConfig)
	envoyconfiglog.Info("ValidateCreate", "type", "EnvoyConfig", "resource", util.ObjectKey(ec).String())
	if err := validate(ec); err != nil {
		return nil, err
	}
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (validator *EnvoyConfigCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	ec := newObj.(*marin3rv1alpha1.EnvoyConfig)
	envoyconfiglog.Info("validateUpdate", "type", "EnvoyConfig", "resource", util.ObjectKey(ec).String())
	if err := validate(ec); err != nil {
		return nil, err
	}
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (validator *EnvoyConfigCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

// Validates the EnvoyConfig resource
func validate(ec *marin3rv1alpha1.EnvoyConfig) error {
	if (ec.Spec.EnvoyResources == nil && ec.Spec.Resources == nil) || (ec.Spec.EnvoyResources != nil && ec.Spec.Resources != nil) {
		return fmt.Errorf("one and only one of 'spec.EnvoyResources', 'spec.Resources' must be set")
	}

	if ec.Spec.EnvoyResources != nil {
		if err := validateEnvoyResources(ec); err != nil {
			return err
		}

	} else {
		if err := validateResources(ec); err != nil {
			return err
		}
	}

	return nil
}

// Validate Envoy Resources against schema
func validateResources(ec *marin3rv1alpha1.EnvoyConfig) error {
	errList := []error{}

	for _, res := range ec.Spec.Resources {

		switch res.Type {

		case envoy.Secret:
			if res.GenerateFromTlsSecret == nil && res.GenerateFromOpaqueSecret == nil {
				errList = append(errList, fmt.Errorf("one of 'generateFromTlsSecret', 'generateFromOpaqueSecret' must be set for type '%s'", envoy.Secret))
			}
			if res.Value != nil {
				errList = append(errList, fmt.Errorf("'value' cannot be used for type '%s'", envoy.Secret))
			}
			if res.GenerateFromEndpointSlices != nil {
				errList = append(errList, fmt.Errorf("'generateFromEndpointSlice' can only be used type '%s'", envoy.Endpoint))
			}

		case envoy.Endpoint:
			if res.GenerateFromEndpointSlices != nil && res.Value != nil {
				errList = append(errList, fmt.Errorf("only one of 'generateFromEndpointSlice', 'value' allowed for type '%s'", envoy.Secret))
			}
			if res.GenerateFromEndpointSlices == nil && res.Value == nil {
				errList = append(errList, fmt.Errorf("one of 'generateFromEndpointSlice', 'value' must be set for type '%s'", envoy.Secret))
			}
			if res.Value != nil {
				if err := envoy_resources.Validate(string(res.Value.Raw), envoy_serializer.JSON, ec.GetEnvoyAPIVersion(), envoy.Type(res.Type)); err != nil {
					errList = append(errList, err)
				}
			}
			if res.GenerateFromTlsSecret != nil {
				errList = append(errList, fmt.Errorf("'generateFromTlsSecret' can only be used type '%s'", envoy.Secret))
			}
			if res.Blueprint != nil {
				errList = append(errList, fmt.Errorf("'blueprint' can only be used type '%s'", envoy.Secret))
			}

		default:
			if res.GenerateFromEndpointSlices != nil {
				errList = append(errList, fmt.Errorf("'generateFromEndpointSlice' can only be used type '%s'", envoy.Endpoint))
			}
			if res.GenerateFromTlsSecret != nil {
				errList = append(errList, fmt.Errorf("'generateFromTlsSecret' can only be used type '%s'", envoy.Secret))
			}
			if res.Blueprint != nil {
				errList = append(errList, fmt.Errorf("'blueprint' cannot be empty for type '%s'", envoy.Secret))
			}
			if res.Value != nil {
				if err := envoy_resources.Validate(string(res.Value.Raw), envoy_serializer.JSON, ec.GetEnvoyAPIVersion(), envoy.Type(res.Type)); err != nil {
					errList = append(errList, err)
				}
			} else {
				errList = append(errList, fmt.Errorf("'value' cannot be empty for type '%s'", res.Type))
			}
		}

	}

	if len(errList) > 0 {
		return errorutil.NewMultiError(errList)
	}
	return nil
}

// Validate EnvoyResources against schema
func validateEnvoyResources(ec *marin3rv1alpha1.EnvoyConfig) error {
	errList := []error{}

	for _, endpoint := range ec.Spec.EnvoyResources.Endpoints {
		if err := envoy_resources.Validate(endpoint.Value, ec.GetSerialization(), ec.GetEnvoyAPIVersion(), envoy.Endpoint); err != nil {
			errList = append(errList, err)
		}
	}

	for _, cluster := range ec.Spec.EnvoyResources.Clusters {
		if err := envoy_resources.Validate(cluster.Value, ec.GetSerialization(), ec.GetEnvoyAPIVersion(), envoy.Cluster); err != nil {
			errList = append(errList, err)
		}
	}

	for _, route := range ec.Spec.EnvoyResources.Routes {
		if err := envoy_resources.Validate(route.Value, ec.GetSerialization(), ec.GetEnvoyAPIVersion(), envoy.Route); err != nil {
			errList = append(errList, err)
		}
	}

	for _, route := range ec.Spec.EnvoyResources.ScopedRoutes {
		if err := envoy_resources.Validate(route.Value, ec.GetSerialization(), ec.GetEnvoyAPIVersion(), envoy.ScopedRoute); err != nil {
			errList = append(errList, err)
		}
	}

	for _, listener := range ec.Spec.EnvoyResources.Listeners {
		if err := envoy_resources.Validate(listener.Value, ec.GetSerialization(), ec.GetEnvoyAPIVersion(), envoy.Listener); err != nil {
			errList = append(errList, err)
		}
	}

	for _, runtime := range ec.Spec.EnvoyResources.Runtimes {
		if err := envoy_resources.Validate(runtime.Value, ec.GetSerialization(), ec.GetEnvoyAPIVersion(), envoy.Runtime); err != nil {
			errList = append(errList, err)
		}
	}

	for _, secret := range ec.Spec.EnvoyResources.Secrets {
		if err := secret.Validate(ec.GetNamespace()); err != nil {
			errList = append(errList, err)
		}
	}

	if len(errList) > 0 {
		return errorutil.NewMultiError(errList)
	}
	return nil
}
