package mutators

import (
	"context"
	"fmt"

	"github.com/3scale-sre/basereconciler/config"
	"github.com/3scale-sre/basereconciler/resource"
	"github.com/3scale-sre/basereconciler/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// RolloutTrigger defines a configuration source that should trigger a
// rollout whenever the data within that configuration source changes.
// Example usage:
//
//	 &resource.Template[*appsv1.Deployment]{
//	 	TemplateBuilder: deployment(),
//	 	IsEnabled:       true,
//	 	TemplateMutations: []resource.TemplateMutationFunction{
//	 		mutators.RolloutTrigger{Name: "secret", SecretName: pointer.String("secret")}.Add(),
//	 	},
//	 },

type RolloutTrigger struct {
	Name          string
	ConfigMapName *string
	SecretName    *string
}

// Add adds the trigger to the Deployment/StatefulSet
func (trigger RolloutTrigger) Add(params ...string) resource.TemplateMutationFunction {
	var domain string
	if len(params) == 0 {
		domain = config.GetAnnotationsDomain()
	} else {
		domain = params[0]
	}
	return func(ctx context.Context, cl client.Client, desired client.Object) error {

		hash, err := trigger.GetHash(ctx, cl, desired.GetNamespace())
		if err != nil {
			return err
		}
		trigger := map[string]string{trigger.GetAnnotationKey(domain): hash}

		switch o := desired.(type) {
		case *appsv1.Deployment:
			o.Spec.Template.ObjectMeta.Annotations =
				util.MergeMaps(map[string]string{}, o.Spec.Template.ObjectMeta.Annotations, trigger)
		case *appsv1.StatefulSet:
			o.Spec.Template.ObjectMeta.Annotations =
				util.MergeMaps(map[string]string{}, o.Spec.Template.ObjectMeta.Annotations, trigger)
		}

		return nil
	}
}

// GetHash returns the hash of the data contained in the RolloutTrigger
// config source
func (rt RolloutTrigger) GetHash(ctx context.Context, cl client.Client, namespace string) (string, error) {

	if rt.SecretName != nil {
		secret := &corev1.Secret{}
		key := types.NamespacedName{Name: *rt.SecretName, Namespace: namespace}
		if err := cl.Get(ctx, key, secret); err != nil {
			if errors.IsNotFound(err) {
				return "", nil
			}
			return "", err
		}
		return util.Hash(secret.Data), nil

	} else if rt.ConfigMapName != nil {
		cm := &corev1.ConfigMap{}
		key := types.NamespacedName{Name: *rt.ConfigMapName, Namespace: namespace}
		if err := cl.Get(ctx, key, cm); err != nil {
			if errors.IsNotFound(err) {
				return "", nil
			}
			return "", err
		}
		return util.Hash(cm.Data), nil

	} else {
		return "", fmt.Errorf("empty rollout trigger")
	}
}

// GetAnnotationKey returns the annotation key to be used in the Pods that read
// from the config source defined in the RolloutTrigger.
func (rt RolloutTrigger) GetAnnotationKey(annotationsDomain string) string {
	if rt.SecretName != nil {
		return fmt.Sprintf("%s/%s.%s", annotationsDomain, rt.Name, "secret-hash")
	}
	return fmt.Sprintf("%s/%s.%s", annotationsDomain, rt.Name, "configmap-hash")
}
