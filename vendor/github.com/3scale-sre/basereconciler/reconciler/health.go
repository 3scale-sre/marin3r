package reconciler

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
)

// Holds health assessment results
type HealthStatus struct {
	Status  HealthStatusCode `json:"status,omitempty"`
	Message string           `json:"message,omitempty"`
}

// Represents resource health status
type HealthStatusCode string

const (
	// Indicates that health assessment failed and actual health status is unknown
	HealthStatusUnknown HealthStatusCode = "Unknown"
	// Progressing health status means that resource is not healthy but still have a chance to reach healthy state
	HealthStatusProgressing HealthStatusCode = "Progressing"
	// Resource is 100% healthy
	HealthStatusHealthy HealthStatusCode = "Healthy"
	// Assigned to resources that are suspended or paused. The typical example is a
	// [suspended](https://kubernetes.io/docs/tasks/job/automated-tasks-with-cron-jobs/#suspend) CronJob.
	HealthStatusSuspended HealthStatusCode = "Suspended"
	// Degrade status is used if resource status indicates failure or resource could not reach healthy state
	// within some timeout.
	HealthStatusDegraded HealthStatusCode = "Degraded"
	// Indicates that resource is missing in the cluster.
	HealthStatusMissing HealthStatusCode = "Missing"
)

func GetDeploymentHealth(deployment *appsv1.Deployment) (*HealthStatus, error) {
	if deployment.Spec.Paused {
		return &HealthStatus{
			Status:  HealthStatusSuspended,
			Message: "Deployment is paused",
		}, nil
	}
	//nolint:lll
	// Borrowed at kubernetes/kubectl/rollout_status.go https://github.com/kubernetes/kubernetes/blob/5232ad4a00ec93942d0b2c6359ee6cd1201b46bc/pkg/kubectl/rollout_status.go#L80
	if deployment.Generation <= deployment.Status.ObservedGeneration {
		cond := getAppsv1DeploymentCondition(deployment.Status, appsv1.DeploymentProgressing)
		switch {
		case cond != nil && cond.Reason == "ProgressDeadlineExceeded":
			return &HealthStatus{
				Status:  HealthStatusDegraded,
				Message: fmt.Sprintf("Deployment %q exceeded its progress deadline", deployment.Name),
			}, nil
		case deployment.Spec.Replicas != nil && deployment.Status.UpdatedReplicas < *deployment.Spec.Replicas:
			return &HealthStatus{
				Status: HealthStatusProgressing,
				Message: fmt.Sprintf("Waiting for rollout to finish: %d out of %d new replicas have been updated...",
					deployment.Status.UpdatedReplicas, *deployment.Spec.Replicas),
			}, nil
		case deployment.Status.Replicas > deployment.Status.UpdatedReplicas:
			return &HealthStatus{
				Status: HealthStatusProgressing,
				Message: fmt.Sprintf("Waiting for rollout to finish: %d old replicas are pending termination...",
					deployment.Status.Replicas-deployment.Status.UpdatedReplicas),
			}, nil
		case deployment.Status.AvailableReplicas < deployment.Status.UpdatedReplicas:
			return &HealthStatus{
				Status: HealthStatusProgressing,
				Message: fmt.Sprintf("Waiting for rollout to finish: %d of %d updated replicas are available...",
					deployment.Status.AvailableReplicas, deployment.Status.UpdatedReplicas),
			}, nil
		}
	} else {
		return &HealthStatus{
			Status:  HealthStatusProgressing,
			Message: "Waiting for rollout to finish: observed deployment generation less than desired generation",
		}, nil
	}

	return &HealthStatus{
		Status: HealthStatusHealthy,
	}, nil
}

func getAppsv1DeploymentCondition(status appsv1.DeploymentStatus,
	condType appsv1.DeploymentConditionType) *appsv1.DeploymentCondition {
	for i := range status.Conditions {
		c := status.Conditions[i]
		if c.Type == condType {
			return &c
		}
	}
	return nil
}

func GetStatefulSetHealth(sts *appsv1.StatefulSet) (*HealthStatus, error) {
	//nolint:lll
	// Borrowed at kubernetes/kubectl/rollout_status.go https://github.com/kubernetes/kubernetes/blob/5232ad4a00ec93942d0b2c6359ee6cd1201b46bc/pkg/kubectl/rollout_status.go#L131
	if sts.Status.ObservedGeneration == 0 || sts.Generation > sts.Status.ObservedGeneration {
		return &HealthStatus{
			Status:  HealthStatusProgressing,
			Message: "Waiting for statefulset spec update to be observed...",
		}, nil
	}
	if sts.Spec.Replicas != nil && sts.Status.ReadyReplicas < *sts.Spec.Replicas {
		return &HealthStatus{
			Status:  HealthStatusProgressing,
			Message: fmt.Sprintf("Waiting for %d pods to be ready...", *sts.Spec.Replicas-sts.Status.ReadyReplicas),
		}, nil
	}
	if sts.Spec.UpdateStrategy.Type == appsv1.RollingUpdateStatefulSetStrategyType &&
		sts.Spec.UpdateStrategy.RollingUpdate != nil {
		if sts.Spec.Replicas != nil && sts.Spec.UpdateStrategy.RollingUpdate.Partition != nil {
			if sts.Status.UpdatedReplicas < (*sts.Spec.Replicas - *sts.Spec.UpdateStrategy.RollingUpdate.Partition) {
				return &HealthStatus{
					Status: HealthStatusProgressing,
					//nolint:lll
					Message: fmt.Sprintf("Waiting for partitioned roll out to finish: %d out of %d new pods have been updated...",
						sts.Status.UpdatedReplicas, (*sts.Spec.Replicas - *sts.Spec.UpdateStrategy.RollingUpdate.Partition)),
				}, nil
			}
		}
		return &HealthStatus{
			Status: HealthStatusHealthy,
			Message: fmt.Sprintf("partitioned roll out complete: %d new pods have been updated...",
				sts.Status.UpdatedReplicas),
		}, nil
	}
	if sts.Spec.UpdateStrategy.Type == appsv1.OnDeleteStatefulSetStrategyType {
		return &HealthStatus{
			Status:  HealthStatusHealthy,
			Message: fmt.Sprintf("statefulset has %d ready pods", sts.Status.ReadyReplicas),
		}, nil
	}
	if sts.Status.UpdateRevision != sts.Status.CurrentRevision {
		return &HealthStatus{
			Status: HealthStatusProgressing,
			Message: fmt.Sprintf("waiting for statefulset rolling update to complete %d pods at revision %s...",
				sts.Status.UpdatedReplicas, sts.Status.UpdateRevision),
		}, nil
	}
	return &HealthStatus{
		Status: HealthStatusHealthy,
		Message: fmt.Sprintf("statefulset rolling update complete %d pods at revision %s...",
			sts.Status.CurrentReplicas, sts.Status.CurrentRevision),
	}, nil
}

// healthOrder is a list of health codes in order of most healthy to least healthy
var healthOrder = []HealthStatusCode{
	HealthStatusHealthy,
	HealthStatusSuspended,
	HealthStatusProgressing,
	HealthStatusMissing,
	HealthStatusDegraded,
	HealthStatusUnknown,
}

// isWorse returns whether or not the new health status code is a worse condition than the current
func healthIsWorse(current, new HealthStatusCode) bool {
	currentIndex := 0
	newIndex := 0
	for i, code := range healthOrder {
		if current == code {
			currentIndex = i
		}
		if new == code {
			newIndex = i
		}
	}
	return newIndex > currentIndex
}
