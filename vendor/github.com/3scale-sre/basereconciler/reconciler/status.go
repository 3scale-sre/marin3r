package reconciler

import (
	"context"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReconcileStatus can reconcile the status of a custom resource when the resource implements
// the ObjectWithAppStatus interface. It is specifically targeted for the status of custom
// resources that deploy Deployments/StatefulSets, as it can aggregate the status of those into the
// status of the custom resource. It also accepts functions with signature "func() bool" that can
// reconcile the status of the custom resource and return whether update is required or not.
func (r *Reconciler) ReconcileStatus(ctx context.Context, instance ObjectWithAppStatus,
	deployments, statefulsets []types.NamespacedName, mutators ...func() bool) Result {
	logger := logr.FromContextOrDiscard(ctx)
	update := false
	status := instance.GetStatus()

	// Aggregate the status of all Deployments owned
	// by this instance
	for _, key := range deployments {
		deployment := &appsv1.Deployment{}
		deploymentStatus := status.GetDeploymentStatus(key)
		if err := r.Client.Get(ctx, key, deployment); err != nil {
			return Result{Error: err}
		}

		if !equality.Semantic.DeepEqual(deploymentStatus, deployment.Status) {
			status.SetDeploymentStatus(key, &deployment.Status)
			update = true
		}
	}

	// Aggregate the status of all StatefulSets owned
	// by this instance
	for _, key := range statefulsets {
		sts := &appsv1.StatefulSet{}
		stsStatus := status.GetStatefulSetStatus(key)
		if err := r.Client.Get(ctx, key, sts); err != nil {
			return Result{Error: err}
		}

		if !equality.Semantic.DeepEqual(stsStatus, sts.Status) {
			status.SetStatefulSetStatus(key, &sts.Status)
			update = true
		}
	}

	// TODO: calculate health

	// call mutators
	for _, fn := range mutators {
		if fn() {
			update = true
		}
	}

	if update {
		if err := r.Client.Status().Update(ctx, instance); err != nil {
			logger.Error(err, "unable to update status")
			return Result{Error: err}
		}
	}

	return Result{Action: ContinueAction}
}

// ObjectWithAppStatus is an interface that implements
// both client.Object and AppStatus
type ObjectWithAppStatus interface {
	client.Object
	GetStatus() AppStatus
}

// Health not yet implemented
type Health string

const (
	Health_Healthy     Health = "Healthy"
	Health_Progressing Health = "Progressing"
	Health_Degraded    Health = "Degraded"
	Health_Suspended   Health = "Suspended"
	Health_Unknown     Health = "Unknown"
)

// AppStatus is an interface describing a custom resource with
// an status that can be reconciled by the reconciler
type AppStatus interface {
	// GetHealth(types.NamespacedName) Health
	// SetHealth(types.NamespacedName, Health)
	GetDeploymentStatus(types.NamespacedName) *appsv1.DeploymentStatus
	SetDeploymentStatus(types.NamespacedName, *appsv1.DeploymentStatus)
	GetStatefulSetStatus(types.NamespacedName) *appsv1.StatefulSetStatus
	SetStatefulSetStatus(types.NamespacedName, *appsv1.StatefulSetStatus)
}

// UnimplementedDeploymentStatus type can be used for resources that doesn't use Deployments
type UnimplementedDeploymentStatus struct{}

func (u *UnimplementedDeploymentStatus) GetDeployments() []types.NamespacedName {
	return nil
}

func (u *UnimplementedDeploymentStatus) GetDeploymentStatus(types.NamespacedName) *appsv1.DeploymentStatus {
	return nil
}

func (u *UnimplementedDeploymentStatus) SetDeploymentStatus(types.NamespacedName, *appsv1.DeploymentStatus) {
}

// UnimplementedStatefulSetStatus type can be used for resources that doesn't use StatefulSets
type UnimplementedStatefulSetStatus struct{}

func (u *UnimplementedStatefulSetStatus) GetStatefulSets() []types.NamespacedName {
	return nil
}

func (u *UnimplementedStatefulSetStatus) GetStatefulSetStatus(types.NamespacedName) *appsv1.StatefulSetStatus {
	return nil
}

func (u *UnimplementedStatefulSetStatus) SetStatefulSetStatus(types.NamespacedName, *appsv1.StatefulSetStatus) {
}
