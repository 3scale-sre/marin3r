package reconciler

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ObjectWithAppStatus is an interface that implements
// both client.Object and AppStatus
type ObjectWithAppStatus interface {
	client.Object
	// GetStatus must return an object that implements AppStatus. The functions returns "any"
	// to allow for implicit implementation of this interface, without ever requiring the user API
	// package to import basereconciler.
	GetStatus() any
}

// AppStatus is an interface describing a custom resource with
// an status that can be reconciled by the reconciler
type AppStatus interface {
	GetDeploymentStatus(types.NamespacedName) *appsv1.DeploymentStatus
	SetDeploymentStatus(types.NamespacedName, *appsv1.DeploymentStatus)
	GetStatefulSetStatus(types.NamespacedName) *appsv1.StatefulSetStatus
	SetStatefulSetStatus(types.NamespacedName, *appsv1.StatefulSetStatus)
}

// AppStatusWithHealth is an interface describing a custom resource with
// an status that includes health information for the workload
type AppStatusWithHealth interface {
	AppStatus
	GetHealthStatus(types.NamespacedName) string
	SetHealthStatus(types.NamespacedName, string)
	GetHealthMessage(types.NamespacedName) string
	SetHealthMessage(types.NamespacedName, string)
}

// AppStatusWithAggregatedHealth is an interface describing a custom resource with
// an status that includes aggregated health information. This is useful only when the
// custom resources managed several workloads.
type AppStatusWithAggregatedHealth interface {
	AppStatusWithHealth
	GetAggregatedHealthStatus() string
	SetAggregatedHealthStatus(string)
}

// UnimplementedDeploymentStatus type can be used for resources that doesn't use Deployments
type UnimplementedDeploymentStatus struct{}

func (u *UnimplementedDeploymentStatus) GetDeploymentStatus(types.NamespacedName) *appsv1.DeploymentStatus {
	return nil
}

func (u *UnimplementedDeploymentStatus) SetDeploymentStatus(types.NamespacedName, *appsv1.DeploymentStatus) {
}

// UnimplementedStatefulSetStatus type can be used for resources that doesn't use StatefulSets
type UnimplementedStatefulSetStatus struct{}

func (u *UnimplementedStatefulSetStatus) GetStatefulSetStatus(types.NamespacedName) *appsv1.StatefulSetStatus {
	return nil
}

func (u *UnimplementedStatefulSetStatus) SetStatefulSetStatus(types.NamespacedName, *appsv1.StatefulSetStatus) {
}

// ReconcileStatus can reconcile the status of a custom resource when the resource implements
// the ObjectWithAppStatus interface. It is specifically targeted for the status of custom
// resources that deploy Deployments/StatefulSets, as it can aggregate the status of those into the
// status of the custom resource. It also accepts functions with signature "func() bool" that can
// reconcile the status of the custom resource and detect whether status update is required or not.
func (r *Reconciler) ReconcileStatus(ctx context.Context, instance client.Object,
	deployments, statefulsets []types.NamespacedName, mutators ...func() (bool, error)) Result {
	logger := logr.FromContextOrDiscard(ctx)
	update := false

	var ok bool

	// ensure the received object implements ObjectWithAppStatus
	var obj ObjectWithAppStatus
	if obj, ok = instance.(ObjectWithAppStatus); !ok {
		return Result{Error: fmt.Errorf(
			"object '%s' with GVK '%s' does not implement ObjectWithAppStatus interface",
			instance.GetName(),
			instance.GetObjectKind().GroupVersionKind(),
		)}
	}

	// ensure the object's status implements AppStatus
	var status AppStatus
	if status, ok = (obj.GetStatus()).(AppStatus); !ok {
		return Result{Error: fmt.Errorf(
			"status for '%s' with GVK '%s' does not implement AppStatus interface",
			instance.GetName(),
			instance.GetObjectKind().GroupVersionKind(),
		)}
	}

	// check if the object's status implements AppStatusWithHealth
	var implementsHealth bool
	if _, ok := (obj.GetStatus()).(AppStatusWithHealth); ok {
		implementsHealth = true
		// set initial aggregated health if object also implements AppStatusWithAggregatedHealth
		if o, ok := (obj.GetStatus()).(AppStatusWithAggregatedHealth); ok {
			o.SetAggregatedHealthStatus(string(HealthStatusHealthy))
		}
	}

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

		// health
		if implementsHealth {
			statusWithHealth := status.(AppStatusWithHealth)

			var err error
			if update, err = setWorkloadHealth(statusWithHealth, deployment); err != nil {
				return Result{Error: err}
			}

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

		// health
		if implementsHealth {
			statusWithHealth := status.(AppStatusWithHealth)

			var err error
			if update, err = setWorkloadHealth(statusWithHealth, sts); err != nil {
				return Result{Error: err}
			}
		}
	}

	// call mutators
	for _, fn := range mutators {
		if result, err := fn(); err != nil {
			return Result{Error: err}
		} else if result {
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

func setWorkloadHealth(status AppStatusWithHealth, obj client.Object) (bool, error) {
	var observedHealth *HealthStatus
	var err error
	var update bool

	switch o := obj.(type) {
	case *appsv1.Deployment:
		observedHealth, err = GetDeploymentHealth(o)
		if err != nil {
			return false, err
		}

	case *appsv1.StatefulSet:
		observedHealth, err = GetStatefulSetHealth(o)
		if err != nil {
			return false, err
		}

	default:
		return false, fmt.Errorf("unsupported workload type %T", o)
	}

	key := client.ObjectKeyFromObject(obj)
	storedHealth := HealthStatus{
		Status:  HealthStatusCode(status.GetHealthStatus(key)),
		Message: status.GetHealthMessage(key),
	}

	if !equality.Semantic.DeepEqual(observedHealth, storedHealth) {
		status.SetHealthStatus(key, string(observedHealth.Status))
		status.SetHealthMessage(key, observedHealth.Message)
		update = true
	}

	// aggregate health if status implements AppStatusWithAggregatedHealth
	o, ok := status.(AppStatusWithAggregatedHealth)
	if ok && healthIsWorse(HealthStatusCode(o.GetAggregatedHealthStatus()),
		HealthStatusCode(status.GetHealthStatus(key))) {
		o.SetAggregatedHealthStatus(status.GetHealthStatus(key))
		update = true
	}

	return update, nil

}
