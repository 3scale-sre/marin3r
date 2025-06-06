package reconcilers

import (
	"context"
	"fmt"

	envoy "github.com/3scale-sre/marin3r/api/envoy"
	envoy_resources "github.com/3scale-sre/marin3r/api/envoy/resources"
	envoy_serializer "github.com/3scale-sre/marin3r/api/envoy/serializer"
	marin3rv1alpha1 "github.com/3scale-sre/marin3r/api/marin3r/v1alpha1"
	xdss "github.com/3scale-sre/marin3r/internal/pkg/discoveryservice/xdss"
	"github.com/3scale-sre/marin3r/internal/pkg/reconcilers/marin3r/envoyconfigrevision/discover"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	secretCertificate = "tls.crt"
	secretPrivateKey  = "tls.key"
)

type CacheReconciler struct {
	logger    logr.Logger
	client    client.Client
	xdsCache  xdss.Cache
	decoder   envoy_serializer.ResourceUnmarshaller
	generator envoy_resources.Generator
}

func NewCacheReconciler(client client.Client, xdsCache xdss.Cache, decoder envoy_serializer.ResourceUnmarshaller,
	generator envoy_resources.Generator, logger logr.Logger) CacheReconciler {
	return CacheReconciler{logger, client, xdsCache, decoder, generator}
}

func (r *CacheReconciler) Reconcile(ctx context.Context, req types.NamespacedName, resources []marin3rv1alpha1.Resource,
	nodeID, version string) (*marin3rv1alpha1.VersionTracker, error) {
	snap, err := r.GenerateSnapshot(ctx, req, resources)

	if err != nil {
		return nil, err
	}

	oldSnap, err := r.xdsCache.GetSnapshot(nodeID)
	if err != nil || areDifferent(snap, oldSnap) {
		r.logger.Info("Writing new snapshot to xDS cache", "Revision", version, "NodeID", nodeID)

		if err := r.xdsCache.SetSnapshot(ctx, nodeID, snap); err != nil {
			return nil, err
		}
	}

	return &marin3rv1alpha1.VersionTracker{
		Endpoints:        snap.GetVersion(envoy.Endpoint),
		Clusters:         snap.GetVersion(envoy.Cluster),
		Routes:           snap.GetVersion(envoy.Route),
		ScopedRoutes:     snap.GetVersion(envoy.ScopedRoute),
		Listeners:        snap.GetVersion(envoy.Listener),
		Secrets:          snap.GetVersion(envoy.Secret),
		Runtimes:         snap.GetVersion(envoy.Runtime),
		ExtensionConfigs: snap.GetVersion(envoy.ExtensionConfig),
	}, nil
}

func (r *CacheReconciler) GenerateSnapshot(ctx context.Context, req types.NamespacedName, resources []marin3rv1alpha1.Resource) (xdss.Snapshot, error) {
	snap := r.xdsCache.NewSnapshot()

	endpoints := make([]envoy.Resource, 0, len(resources))
	clusters := make([]envoy.Resource, 0, len(resources))
	routes := make([]envoy.Resource, 0, len(resources))
	scopedRoutes := make([]envoy.Resource, 0, len(resources))
	listeners := make([]envoy.Resource, 0, len(resources))
	runtimes := make([]envoy.Resource, 0, len(resources))
	extensionConfigs := make([]envoy.Resource, 0, len(resources))
	secrets := make([]envoy.Resource, 0, len(resources))

	for idx, resourceDefinition := range resources {
		switch resourceDefinition.Type {
		case envoy.Endpoint:
			if resourceDefinition.GenerateFromEndpointSlices != nil {
				// Endpoint discovery enabled
				endpoint, err := discover.Endpoints(ctx, r.client, req.Namespace,
					resourceDefinition.GenerateFromEndpointSlices.ClusterName,
					resourceDefinition.GenerateFromEndpointSlices.TargetPort,
					resourceDefinition.GenerateFromEndpointSlices.Selector,
					r.generator, r.logger)
				if err != nil {
					return nil, err
				}

				endpoints = append(endpoints, endpoint)
			} else {
				// Raw value provided
				res := r.generator.New(envoy.Endpoint)
				if err := r.decoder.Unmarshal(string(resourceDefinition.Value.Raw), res); err != nil {
					return nil,
						resourceLoaderError(
							req, string(resourceDefinition.Value.Raw), field.NewPath("spec", "resources").Index(idx).Child("value"),
							fmt.Sprintf("Invalid envoy resource value: '%s'", err),
						)
				}

				endpoints = append(endpoints, res)
			}

		case envoy.Cluster:
			res := r.generator.New(envoy.Cluster)
			if err := r.decoder.Unmarshal(string(resourceDefinition.Value.Raw), res); err != nil {
				return nil,
					resourceLoaderError(
						req, string(resourceDefinition.Value.Raw), field.NewPath("spec", "resources").Index(idx).Child("value"),
						fmt.Sprintf("Invalid envoy resource value: '%s'", err),
					)
			}

			clusters = append(clusters, res)

		case envoy.Route:
			res := r.generator.New(envoy.Route)
			if err := r.decoder.Unmarshal(string(resourceDefinition.Value.Raw), res); err != nil {
				return nil,
					resourceLoaderError(
						req, string(resourceDefinition.Value.Raw), field.NewPath("spec", "resources").Index(idx).Child("value"),
						fmt.Sprintf("Invalid envoy resource value: '%s'", err),
					)
			}

			routes = append(routes, res)

		case envoy.ScopedRoute:
			res := r.generator.New(envoy.ScopedRoute)
			if err := r.decoder.Unmarshal(string(resourceDefinition.Value.Raw), res); err != nil {
				return nil,
					resourceLoaderError(
						req, string(resourceDefinition.Value.Raw), field.NewPath("spec", "resources").Index(idx).Child("value"),
						fmt.Sprintf("Invalid envoy resource value: '%s'", err),
					)
			}

			scopedRoutes = append(scopedRoutes, res)

		case envoy.Listener:
			res := r.generator.New(envoy.Listener)
			if err := r.decoder.Unmarshal(string(resourceDefinition.Value.Raw), res); err != nil {
				return nil,
					resourceLoaderError(
						req, string(resourceDefinition.Value.Raw), field.NewPath("spec", "resources").Index(idx).Child("value"),
						fmt.Sprintf("Invalid envoy resource value: '%s'", err),
					)
			}

			listeners = append(listeners, res)

		case envoy.Secret:
			var res envoy.Resource

			s := &corev1.Secret{}

			if resourceDefinition.GenerateFromTlsSecret != nil {
				name := *resourceDefinition.GenerateFromTlsSecret

				key := types.NamespacedName{Name: name, Namespace: req.Namespace}
				if err := r.client.Get(ctx, key, s); err != nil {
					return nil, fmt.Errorf("%s", err.Error())
				}

				if s.Type != corev1.SecretTypeTLS {
					return nil, fmt.Errorf("expected Secret of '%s' type", corev1.SecretTypeTLS)
				}

				switch resourceDefinition.GetBlueprint() {
				case marin3rv1alpha1.TlsCertificate:
					res = r.generator.NewTlsCertificateSecret(name, string(s.Data[secretPrivateKey]), string(s.Data[secretCertificate]))
				case marin3rv1alpha1.TlsValidationContext:
					res = r.generator.NewValidationContextSecret(name, string(s.Data[secretCertificate]))
				}
			} else if resourceDefinition.GenerateFromOpaqueSecret != nil {
				name := resourceDefinition.GenerateFromOpaqueSecret.Name

				key := types.NamespacedName{Name: name, Namespace: req.Namespace}
				if err := r.client.Get(ctx, key, s); err != nil {
					return nil, fmt.Errorf("%s", err.Error())
				}

				if s.Type != corev1.SecretTypeOpaque {
					return nil, fmt.Errorf("expected Secret of '%s' type", corev1.SecretTypeOpaque)
				}

				res = r.generator.NewGenericSecret(resourceDefinition.GenerateFromOpaqueSecret.Alias, string(s.Data[resourceDefinition.GenerateFromOpaqueSecret.Key]))
				m := envoy_serializer.NewResourceMarshaller(envoy_serializer.YAML, envoy.APIv3)
				yaml, _ := m.Marshal(res)

				fmt.Println("###################################")
				spew.Dump(yaml)
				fmt.Println("###################################")
			} else {
				return nil, resourceLoaderError(
					req, resourceDefinition, field.NewPath("spec", "resources").Index(idx),
					"one of 'generateFromOpaqueSecret', 'generateFromTlsSecret' must be set",
				)
			}

			secrets = append(secrets, res)

		case envoy.Runtime:
			res := r.generator.New(envoy.Runtime)
			if err := r.decoder.Unmarshal(string(resourceDefinition.Value.Raw), res); err != nil {
				return nil,
					resourceLoaderError(
						req, string(resourceDefinition.Value.Raw), field.NewPath("spec", "resources").Index(idx).Child("value"),
						fmt.Sprintf("Invalid envoy resource value: '%s'", err),
					)
			}

			runtimes = append(runtimes, res)

		case envoy.ExtensionConfig:
			res := r.generator.New(envoy.ExtensionConfig)
			if err := r.decoder.Unmarshal(string(resourceDefinition.Value.Raw), res); err != nil {
				return nil,
					resourceLoaderError(
						req, string(resourceDefinition.Value.Raw), field.NewPath("spec", "resources").Index(idx).Child("value"),
						fmt.Sprintf("Invalid envoy resource value: '%s'", err),
					)
			}

			extensionConfigs = append(extensionConfigs, res)

		default:
		}
	}

	snap.SetResources(envoy.Endpoint, endpoints)
	snap.SetResources(envoy.Cluster, clusters)
	snap.SetResources(envoy.Route, routes)
	snap.SetResources(envoy.ScopedRoute, scopedRoutes)
	snap.SetResources(envoy.Listener, listeners)
	snap.SetResources(envoy.Secret, secrets)
	snap.SetResources(envoy.Runtime, runtimes)
	snap.SetResources(envoy.ExtensionConfig, extensionConfigs)

	return snap, nil
}

func resourceLoaderError(req types.NamespacedName, value interface{}, resPath *field.Path, msg string) error {
	return errors.NewInvalid(
		schema.GroupKind{Group: "envoy", Kind: "EnvoyConfig"},
		fmt.Sprintf("%s/%s", req.Namespace, req.Name),
		field.ErrorList{field.Invalid(resPath, value, msg)},
	)
}

func areDifferent(a, b xdss.Snapshot) bool {
	for _, rType := range []envoy.Type{envoy.Endpoint, envoy.Cluster, envoy.Route, envoy.ScopedRoute,
		envoy.Listener, envoy.Secret, envoy.Runtime, envoy.ExtensionConfig} {
		if a.GetVersion(rType) != b.GetVersion(rType) {
			return true
		}
	}

	return false
}
