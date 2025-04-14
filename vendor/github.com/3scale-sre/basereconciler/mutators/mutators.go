package mutators

import (
	"context"
	"fmt"

	"github.com/3scale-sre/basereconciler/resource"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SetDeploymentReplicas reconciles the number of replicas of a Deployment. If "enforce"
// is set to true, the value in the template is enforce, overwritting the live value. If
// the "enforce" is set to false, the live value obtained from the Kubernetes API is used.
// In general, if the Deployment uses HorizontalPodAutoscaler or any other controller modifies
// the number of replicas, enforce needs to be "false".
// Example usage:
//
//	&resource.Template[*appsv1.Deployment]{
//		TemplateBuilder: deployment(),
//		IsEnabled:       true,
//		TemplateMutations: []resource.TemplateMutationFunction{
//			mutators.SetDeploymentReplicas(!hpaExists()),
//		},
//	},
func SetDeploymentReplicas(enforce bool) resource.TemplateMutationFunction {
	return func(ctx context.Context, cl client.Client, desired client.Object) error {
		if enforce {
			// Let the value in the template
			// override the runtime value
			return nil
		}

		live := &appsv1.Deployment{}
		if err := cl.Get(ctx, client.ObjectKeyFromObject(desired), live); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			return fmt.Errorf("unable to retrieve live object: %w", err)
		}

		// override the value in the template with the
		// runtime value
		desired.(*appsv1.Deployment).Spec.Replicas = live.Spec.Replicas
		return nil
	}
}

// SetServiceLiveValues retrieves some live values of the Service spec from the Kubernetes
// API to avoid overwriting them. These values are typically set the by the kube-controller-manager
// (in some rare occasions the user might explicitly set them) and should not be modified by the
// reconciler. The fields that this function keeps in sync with the live state are:
//   - spec.clusterIP
//   - spec.ClisterIPs
//   - spec.pors[*].nodePort (when the Service type is not ClusterIP)
//
// Example usage:
//
//	&resource.Template[*corev1.Service]{
//		TemplateBuilder: service(req.Namespace, instance.Spec.ServiceAnnotations),
//		IsEnabled:       true,
//		TemplateMutations: []resource.TemplateMutationFunction{
//			mutators.SetServiceLiveValues(),
//		},
//	}
func SetServiceLiveValues() resource.TemplateMutationFunction {
	return func(ctx context.Context, cl client.Client, desired client.Object) error {

		svc := desired.(*corev1.Service)
		live := &corev1.Service{}
		if err := cl.Get(ctx, client.ObjectKeyFromObject(desired), live); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			return fmt.Errorf("unable to retrieve live object: %w", err)
		}

		// Set runtime values in the resource:
		// "/spec/clusterIP", "/spec/clusterIPs", "/spec/ports/*/nodePort"
		svc.Spec.ClusterIP = live.Spec.ClusterIP
		svc.Spec.ClusterIPs = live.Spec.ClusterIPs

		// For services that are not ClusterIP we need to populate the runtime values
		// of NodePort for each port
		if svc.Spec.Type != corev1.ServiceTypeClusterIP {
			for idx, port := range svc.Spec.Ports {
				runtimePort := findPort(port.Port, port.Protocol, live.Spec.Ports)
				if runtimePort != nil {
					svc.Spec.Ports[idx].NodePort = runtimePort.NodePort
				}
			}
		}
		return nil
	}
}

// findPort returns the Service port identified by port/protocol
func findPort(pNumber int32, pProtocol corev1.Protocol, ports []corev1.ServicePort) *corev1.ServicePort {
	// Ports within a svc are uniquely identified by
	// the "port" and "protocol" fields. This is documented in
	// k8s API reference
	for _, port := range ports {
		if pNumber == port.Port && pProtocol == port.Protocol {
			return &port
		}
	}
	// not found
	return nil
}
