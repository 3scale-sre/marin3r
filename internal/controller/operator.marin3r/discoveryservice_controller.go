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
	"time"

	"github.com/3scale-sre/basereconciler/mutators"
	"github.com/3scale-sre/basereconciler/reconciler"
	"github.com/3scale-sre/basereconciler/resource"
	operatorv1alpha1 "github.com/3scale-sre/marin3r/api/operator.marin3r/v1alpha1"
	"github.com/3scale-sre/marin3r/internal/pkg/reconcilers/operator/discoveryservice/generators"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DiscoveryServiceReconciler reconciles a DiscoveryService object
type DiscoveryServiceReconciler struct {
	*reconciler.Reconciler
}

// +kubebuilder:rbac:groups=operator.marin3r.3scale.net,namespace=placeholder,resources=*,verbs=*
// +kubebuilder:rbac:groups=marin3r.3scale.net,namespace=placeholder,resources=*,verbs=*
// +kubebuilder:rbac:groups="core",namespace=placeholder,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="core",namespace=placeholder,resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="apps",namespace=placeholder,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",namespace=placeholder,resources=roles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",namespace=placeholder,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.marin3r.3scale.net,namespace=placeholder,resources=discoveryservicecertificates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="core",namespace=placeholder,resources=pods,verbs=list;watch;get
// +kubebuilder:rbac:groups="core",namespace=placeholder,resources=secrets,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups="discovery.k8s.io",namespace=placeholder,resources=endpointslices,verbs=get;list;watch

func (r *DiscoveryServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx, _ = r.Logger(ctx, "name", req.Name, "namespace", req.Namespace)
	ds := &operatorv1alpha1.DiscoveryService{}

	result := r.ManageResourceLifecycle(ctx, req, ds)
	if result.ShouldReturn() {
		return result.Values()
	}

	gen := generators.GeneratorOptions{
		InstanceName:                      ds.GetName(),
		Namespace:                         ds.GetNamespace(),
		RootCertificateNamePrefix:         "marin3r-ca-cert",
		RootCertificateCommonNamePrefix:   "marin3r-ca",
		RootCertificateDuration:           ds.GetRootCertificateAuthorityOptions().Duration.Duration,
		ServerCertificateNamePrefix:       "marin3r-server-cert",
		ServerCertificateCommonNamePrefix: "marin3r-server",
		ServerCertificateDuration:         ds.GetServerCertificateOptions().Duration.Duration,
		ClientCertificateDuration: func() (d time.Duration) {
			d, _ = time.ParseDuration("48h")

			return
		}(),
		XdsServerPort:       int32(ds.GetXdsServerPort()),
		MetricsServerPort:   int32(ds.GetMetricsPort()),
		ProbePort:           int32(ds.GetProbePort()),
		ServiceType:         operatorv1alpha1.ClusterIPType,
		DeploymentImage:     ds.GetImage(),
		DeploymentResources: ds.Resources(),
		Debug:               ds.Debug(),
		PodPriorityClass:    ds.GetPriorityClass(),
		Affinity:            ds.Affinity(),
	}

	serverCertHash, err := r.calculateServerCertificateHash(ctx, types.NamespacedName{Name: gen.ServerCertName(), Namespace: gen.Namespace})
	if err != nil {
		return ctrl.Result{}, err
	}

	resources := []resource.TemplateInterface{
		resource.NewTemplateFromObjectFunction(gen.RootCertificationAuthority).Apply(dscDefaulter),
		resource.NewTemplateFromObjectFunction(gen.ServerCertificate).Apply(dscDefaulter),
		resource.NewTemplateFromObjectFunction(gen.ClientCertificate).Apply(dscDefaulter),
		resource.NewTemplateFromObjectFunction(gen.ServiceAccount),
		resource.NewTemplateFromObjectFunction(gen.Role),
		resource.NewTemplateFromObjectFunction(gen.RoleBinding),
		resource.NewTemplateFromObjectFunction(gen.Service).WithMutation(mutators.SetServiceLiveValues()),
		resource.NewTemplateFromObjectFunction(gen.Deployment(serverCertHash)).WithEnabled(serverCertHash != ""),
	}

	result = r.ReconcileOwnedResources(ctx, ds, resources)
	if result.ShouldReturn() {
		return result.Values()
	}
	// requeue if the server certificate is not ready
	if serverCertHash == "" {
		return ctrl.Result{Requeue: true}, nil
	}

	// reconcile the status
	result = r.ReconcileStatus(ctx, ds, []types.NamespacedName{{Name: gen.ResourceName(), Namespace: ds.GetNamespace()}}, nil,
		func() bool {
			if ds.Status.DeploymentName == nil || *ds.Status.DeploymentName != gen.ResourceName() {
				ds.Status.DeploymentName = ptr.To(gen.ResourceName())

				return true
			}

			return false
		})
	if result.ShouldReturn() {
		return result.Values()
	}

	return ctrl.Result{}, nil
}

func (r *DiscoveryServiceReconciler) calculateServerCertificateHash(ctx context.Context, key types.NamespacedName) (string, error) {
	// Fetch the server certificate to calculate the hash and
	// populate the deployment's label.
	// This will trigger rollouts on server certificate changes.
	serverDSC := &operatorv1alpha1.DiscoveryServiceCertificate{}

	err := r.Client.Get(ctx, key, serverDSC)
	if err != nil {
		if errors.IsNotFound(err) {
			// The server certificate hasn't been created yet
			return "", nil
		}

		return "", err
	}

	return serverDSC.Status.GetCertificateHash(), nil
}

func dscDefaulter(o client.Object) (*operatorv1alpha1.DiscoveryServiceCertificate, error) {
	dsc := o.(*operatorv1alpha1.DiscoveryServiceCertificate)
	dsc.Default()

	return dsc, nil
}

// SetupWithManager adds the controller to the manager
func (r *DiscoveryServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.DiscoveryService{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&operatorv1alpha1.DiscoveryServiceCertificate{}).
		Complete(r)
}
