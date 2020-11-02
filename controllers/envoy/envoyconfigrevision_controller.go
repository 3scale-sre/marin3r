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

package controllers

import (
	"context"
	"fmt"
	"hash/fnv"

	envoyv1alpha1 "github.com/3scale/marin3r/apis/envoy/v1alpha1"
	common "github.com/3scale/marin3r/pkg/common"
	xdss "github.com/3scale/marin3r/pkg/discoveryservice/xdss"
	envoy "github.com/3scale/marin3r/pkg/envoy"
	envoy_resources "github.com/3scale/marin3r/pkg/envoy/resources"
	envoy_resources_v2 "github.com/3scale/marin3r/pkg/envoy/resources/v2"
	envoy_serializer_v2 "github.com/3scale/marin3r/pkg/envoy/serializer/v2"

	envoy_api_v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_api_v2_route "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	envoy_service_discovery_v2 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"

	"github.com/go-logr/logr"
	"github.com/operator-framework/operator-lib/status"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	secretCertificate = "tls.crt"
	secretPrivateKey  = "tls.key"
)

// EnvoyConfigRevisionReconciler reconciles a EnvoyConfigRevision object
type EnvoyConfigRevisionReconciler struct {
	Client   client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	XdsCache xdss.Cache
}

func (r *EnvoyConfigRevisionReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("envoyconfigrevision", req.NamespacedName)
	r.Log.Info("Reconciling EnvoyConfigRevision")

	// Fetch the EnvoyConfigRevision instance
	ecr := &envoyv1alpha1.EnvoyConfigRevision{}
	err := r.Client.Get(ctx, req.NamespacedName, ecr)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	// If this ecr has the RevisionPublishedCondition set to "True" pusblish the resources
	// to the xds server cache
	if ecr.Status.Conditions.IsTrueFor(envoyv1alpha1.RevisionPublishedCondition) {

		nodeID := ecr.Spec.NodeID
		version := ecr.Spec.Version
		snap := r.XdsCache.NewSnapshot(version)

		// Deserialize envoy resources from the spec and create a new snapshot with them
		if err := r.loadResources(ctx, req.Name, req.Namespace,
			ecr.Spec.Serialization, ecr.Spec.EnvoyResources, field.NewPath("spec", "resources"), snap); err != nil {
			// Requeue with delay, as the envoy resources syntax is probably wrong
			// and that is not a transitory error (some other higher level resource
			// probaly needs fixing)
			r.Log.Error(err, "Errors occured while loading resources from CR")
			if err := r.taintSelf(ctx, ecr, "FailedLoadingResources", err.Error()); err != nil {
				return ctrl.Result{}, err
			}
			// This is an unrecoverable error because resources are wrong
			// so do not reque
			return ctrl.Result{}, nil
		}

		oldSnap, err := r.XdsCache.GetSnapshot(nodeID)
		// Publish the generated snapshot when the version is different from the published one. We look specifically
		// for the version of the "Secret" resources because secrets can change even when the spec hasn't changed.
		// Publish the snapshot when an error retrieving the published one occurs as it means that no snpshot has already
		// been written to the cache for that specific nodeID.
		if snap.GetVersion(envoy_resources.Secret) != oldSnap.GetVersion(envoy_resources.Secret) || err != nil {
			r.Log.Info("Publishing new snapshot for nodeID", "Version", version, "NodeID", nodeID)
			r.Log.V(1).Info("Writting to xDS cache", "CurrentVersion", oldSnap.GetVersion(envoy_resources.Secret),
				"NewVersion", snap.GetVersion(envoy_resources.Secret))
			if err := r.XdsCache.SetSnapshot(nodeID, snap); err != nil {
				return ctrl.Result{}, err
			}

		} else {
			r.Log.V(1).Info("Generated snapshot is equal to published one, avoiding push to xds server cache", "NodeID", nodeID)
		}
	}

	// Update status
	if err := r.updateStatus(ctx, ecr); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *EnvoyConfigRevisionReconciler) loadResources(ctx context.Context, name, namespace, serialization string,
	resources *envoyv1alpha1.EnvoyResources, resPath *field.Path, snap xdss.Snapshot) error {

	var ds envoy_serializer_v2.ResourceUnmarshaller
	switch serialization {
	case "b64json":
		ds = envoy_serializer_v2.B64JSON{}

	case "yaml":
		ds = envoy_serializer_v2.YAML{}
	default:
		// "json" is the default
		ds = envoy_serializer_v2.JSON{}
	}

	for idx, endpoint := range resources.Endpoints {
		res := &envoy_api_v2.ClusterLoadAssignment{}
		if err := ds.Unmarshal(endpoint.Value, res); err != nil {
			return resourceLoaderError(name, namespace, "Endpoints", endpoint.Value, resPath, idx)
		}
		snap.SetResource(endpoint.Name, res)
	}

	for idx, cluster := range resources.Clusters {
		res := &envoy_api_v2.Cluster{}
		if err := ds.Unmarshal(cluster.Value, res); err != nil {
			return resourceLoaderError(name, namespace, "Clusters", cluster.Value, resPath, idx)
		}
		snap.SetResource(cluster.Name, res)
	}

	for idx, route := range resources.Routes {
		res := &envoy_api_v2_route.Route{}
		if err := ds.Unmarshal(route.Value, res); err != nil {
			return resourceLoaderError(name, namespace, "Routes", route.Value, resPath, idx)
		}
		snap.SetResource(route.Name, res)
	}

	for idx, listener := range resources.Listeners {
		res := &envoy_api_v2.Listener{}
		if err := ds.Unmarshal(listener.Value, res); err != nil {
			return resourceLoaderError(name, namespace, "Listeners", listener.Value, resPath, idx)
		}
		snap.SetResource(listener.Name, res)
	}

	for idx, runtime := range resources.Runtimes {
		res := &envoy_service_discovery_v2.Runtime{}
		if err := ds.Unmarshal(runtime.Value, res); err != nil {
			return resourceLoaderError(name, namespace, "Runtimes", runtime.Value, resPath, idx)
		}
		snap.SetResource(runtime.Name, res)
	}

	for idx, secret := range resources.Secrets {
		s := &corev1.Secret{}
		key := types.NamespacedName{
			Name:      secret.Ref.Name,
			Namespace: secret.Ref.Namespace,
		}
		if err := r.Client.Get(ctx, key, s); err != nil {
			if errors.IsNotFound(err) {
				return err
			}
			return err
		}

		// Validate secret holds a certificate
		if s.Type == "kubernetes.io/tls" {
			res := envoy_resources_v2.NewSecret(secret.Name, string(s.Data[secretPrivateKey]), string(s.Data[secretCertificate]))
			snap.SetResource(secret.Name, res)
		} else {
			return errors.NewInvalid(
				schema.GroupKind{Group: "caches", Kind: "EnvoyConfig"},
				fmt.Sprintf("%s/%s", namespace, name),
				field.ErrorList{
					field.Invalid(
						resPath.Child("Secrets").Index(idx).Child("Ref"),
						secret.Ref,
						fmt.Sprint("Only 'kubernetes.io/tls' type secrets allowed"),
					),
				},
			)
		}
	}

	// Secrets are a special case of resources as its values are not defined in the spec. This
	// could cause that new values in the same set of secrets wouldn't trigger updates to the envoy
	// gateways as the version (the Spec.Resources hash) would be the same. To avoid this problem, we
	// append the hash of the values of the secrets to the version of the secret resources that will
	// only trigger secret updates in the envoy gateways when necessary
	secretsHash := calculateSecretsHash(snap.GetResources(envoy_resources.Secret))
	snap.SetVersion(
		envoy_resources.Secret,
		fmt.Sprintf("%s-%s", snap.GetVersion(envoy_resources.Secret), secretsHash),
	)

	return nil
}

func resourceLoaderError(name, namespace, rtype, rvalue string, resPath *field.Path, idx int) error {
	return errors.NewInvalid(
		schema.GroupKind{Group: "caches", Kind: "EnvoyConfig"},
		fmt.Sprintf("%s/%s", namespace, name),
		field.ErrorList{
			field.Invalid(
				resPath.Child(rtype).Index(idx).Child("Value"),
				rvalue,
				fmt.Sprint("Invalid envoy resource value"),
			),
		},
	)
}

func (r *EnvoyConfigRevisionReconciler) taintSelf(ctx context.Context, ecr *envoyv1alpha1.EnvoyConfigRevision, reason, msg string) error {
	if !ecr.Status.Conditions.IsTrueFor(envoyv1alpha1.RevisionTaintedCondition) {
		patch := client.MergeFrom(ecr.DeepCopy())
		ecr.Status.Conditions.SetCondition(status.Condition{
			Type:    envoyv1alpha1.RevisionTaintedCondition,
			Status:  corev1.ConditionTrue,
			Reason:  status.ConditionReason(reason),
			Message: msg,
		})
		ecr.Status.Tainted = true

		if err := r.Client.Status().Patch(ctx, ecr, patch); err != nil {
			return err
		}
	}
	return nil
}

func (r *EnvoyConfigRevisionReconciler) updateStatus(ctx context.Context, ecr *envoyv1alpha1.EnvoyConfigRevision) error {

	changed := false
	patch := client.MergeFrom(ecr.DeepCopy())

	// Clear ResourcesOutOfSyncCondition
	if ecr.Status.Conditions.IsTrueFor(envoyv1alpha1.ResourcesOutOfSyncCondition) {
		ecr.Status.Conditions.SetCondition(status.Condition{
			Type:    envoyv1alpha1.ResourcesOutOfSyncCondition,
			Reason:  "NodeConficRevisionSynced",
			Status:  corev1.ConditionFalse,
			Message: "EnvoyConfigRevision successfully synced",
		})
		changed = true

	}

	// Set status.published and status.lastPublishedAt fields
	if ecr.Status.Conditions.IsTrueFor(envoyv1alpha1.RevisionPublishedCondition) && !ecr.Status.Published {
		ecr.Status.Published = true
		ecr.Status.LastPublishedAt = metav1.Now()
		// We also initialise the "tainted" status property to false
		ecr.Status.Tainted = false
		changed = true
	} else if !ecr.Status.Conditions.IsTrueFor(envoyv1alpha1.RevisionPublishedCondition) && ecr.Status.Published {
		ecr.Status.Published = false
		changed = true
	}

	// Set status.failed field
	if ecr.Status.Conditions.IsTrueFor(envoyv1alpha1.RevisionTaintedCondition) && !ecr.Status.Tainted {
		ecr.Status.Tainted = true
		changed = true
	} else if !ecr.Status.Conditions.IsTrueFor(envoyv1alpha1.RevisionTaintedCondition) && ecr.Status.Tainted {
		ecr.Status.Tainted = false
		changed = true
	}

	if changed {
		if err := r.Client.Status().Patch(ctx, ecr, patch); err != nil {
			return err
		}
	}

	return nil
}

func calculateSecretsHash(resources map[string]envoy.Resource) string {
	resourcesHasher := fnv.New32a()
	common.DeepHashObject(resourcesHasher, resources)
	return rand.SafeEncodeString(fmt.Sprint(resourcesHasher.Sum32()))
}

func (r *EnvoyConfigRevisionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&envoyv1alpha1.EnvoyConfigRevision{}).
		Complete(r)
}
