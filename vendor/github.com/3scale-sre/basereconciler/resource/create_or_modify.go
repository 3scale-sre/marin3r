package resource

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/3scale-sre/basereconciler/util"
	"github.com/go-logr/logr"
	"github.com/nsf/jsondiff"
	"github.com/ohler55/ojg/jp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// ModifyOp represents the type of modification operation to use when reconciling resources.
type ModifyOp string

const (
	// ModifyOpUpdate uses full object updates (client.Update) for reconciliation.
	// This replaces the entire object and requires up-to-date resource versions.
	ModifyOpUpdate ModifyOp = "Update"

	// ModifyOpPatch uses patch operations (client.Patch with client.MergeFrom) for reconciliation.
	// This sends only changed fields and avoids conflicts from out-of-date resource versions.
	ModifyOpPatch ModifyOp = "Patch"
)

// CreateOrModify implements intelligent resource reconciliation with configurable modification operations.
// It consolidates the logic from both CreateOrUpdate and CreateOrPatch, with the operation type
// determined by the template's configuration rather than being explicitly passed.
//
// Core workflow:
//  1. Build the desired state from the template (including mutations)
//  2. Get the current live state from the Kubernetes API
//  3. Normalize both states using configuration rules (ensure/ignore properties)
//  4. Compare normalized states to detect if changes are required
//  5. If changes needed, apply via Update or Patch operation based on template.GetReconcileOptions().ModifyOp
//  6. Return object reference
//
// Parameters:
//   - ctx: the context. The logger is expected to be within the context, otherwise the function won't
//     produce any logs.
//   - cl: the kubernetes API client
//   - scheme: the kubernetes API scheme
//   - owner: the object that owns the resource. Used to set the OwnerReference in the resource
//   - template: the struct that describes how the resource needs to be reconciled. It must implement
//     the TemplateInterface interface. The template's configuration determines:
//   - Which properties to ensure/ignore (from GetReconcileOptions())
//   - The modification operation type (Update or Patch) to use
//   - The GVK of the resource (from GetGVK())
func CreateOrModify(ctx context.Context, cl client.Client, scheme *runtime.Scheme,
	owner client.Object, template TemplateInterface) (*corev1.ObjectReference, error) {

	// STEP 1: Build the desired state from the template
	// This calls the template's Build() method which:
	// - Executes the TemplateBuilder function to create the base object
	// - Applies any TemplateMutationFunction functions to modify the object using live cluster data
	desired, err := template.Build(ctx, cl, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to build template: %w", err)
	}

	// Extract metadata for logging and API operations
	key := client.ObjectKeyFromObject(desired)
	gvk, err := apiutil.GVKForObject(desired, scheme)
	if err != nil {
		return nil, err
	}
	logger := logr.FromContextOrDiscard(ctx).WithValues("gvk", gvk, "resource", desired.GetName())

	// STEP 2: Get the current live state from the Kubernetes API
	// Create an empty object of the same type as desired to retrieve the live state
	live, err := util.NewObjectFromGVK(gvk, scheme)
	if err != nil {
		return nil, wrapError("unable to create object from GVK", key, gvk, err)
	}
	err = cl.Get(ctx, key, live)
	if err != nil {
		if errors.IsNotFound(err) {
			// Resource doesn't exist in the cluster
			if template.Enabled() {
				// Template is enabled, so create the resource
				// Note: Creation always uses Create() operation regardless of op parameter
				if err := controllerutil.SetControllerReference(owner, desired, scheme); err != nil {
					return nil, wrapError("unable to set controller reference", key, gvk, err)
				}
				err = cl.Create(ctx, util.SetTypeMeta(desired, gvk))
				if err != nil {
					return nil, wrapError("unable to create resource", key, gvk, err)
				}
				logger.Info("resource created")
				return util.ObjectReference(desired, gvk), nil

			} else {
				// Template is disabled and resource doesn't exist, nothing to do
				return nil, nil
			}
		}
		return nil, wrapError("unable to get resource", key, gvk, err)
	}

	// STEP 3: Handle disabled templates
	// If template is disabled but resource exists, delete it
	// Note: Deletion always uses Delete() operation regardless of op parameter
	if !template.Enabled() {
		err := cl.Delete(ctx, live)
		if err != nil {
			return nil, wrapError("unable to delete object", key, gvk, err)
		}
		logger.Info("resource deleted")
		return nil, nil
	}

	// STEP 4: Determine reconciliation configuration
	// Get the list of properties to ensure and ignore during reconciliation
	ensure := template.GetReconcileOptions().EnsureProperties
	ignore := template.GetReconcileOptions().IgnoreProperties

	if err != nil {
		return nil, wrapError("unable to retrieve config for resource reconciler", key, gvk, err)
	}

	// STEP 5: Normalize both desired and live states for comparison
	// Normalization creates comparable versions of the objects that:
	// - Include only the properties specified in 'ensure' list
	// - Exclude properties specified in 'ignore' list
	// This allows for precise control over what gets compared and reconciled
	normalizedDesired, err := normalize(desired, ensure, ignore, gvk, scheme)
	if err != nil {
		return nil, wrapError("unable to normalize desired", key, gvk, err)
	}

	normalizedLive, err := normalize(live, ensure, ignore, gvk, scheme)
	if err != nil {
		return nil, wrapError("unable to normalize live", key, gvk, err)
	}

	// STEP 6: Compare normalized states to detect if reconciliation is needed
	if !equality.Semantic.DeepEqual(normalizedLive, normalizedDesired) {
		// STEP 7: Apply changes using the specified operation type
		u_live, err := reconcilePropertiesToUnstructured(live, ensure, normalizedDesired, gvk, logger)
		if err != nil {
			return nil, wrapError("unable to reconcile properties for modify", key, gvk, err)
		}

		op := template.GetReconcileOptions().ModifyOp
		switch op {
		case ModifyOpUpdate:
			logger.V(1).Info("resource update required", "diff", printfDiff(normalizedLive, normalizedDesired))

			// Perform full object update
			if err := cl.Update(ctx, client.Object(&unstructured.Unstructured{Object: u_live})); err != nil {
				return nil, wrapError("unable to update resource", key, gvk, err)
			}
			logger.Info("Resource updated")

		case ModifyOpPatch:
			logger.V(1).Info("resource patch required", "diff", printfDiff(normalizedLive, normalizedDesired))

			// Use the patch approach from CreateOrPatch
			// Create a patch object using the current live state as the base
			patch := client.MergeFrom(live.DeepCopyObject().(client.Object))

			// Convert the modified unstructured object back to the live object
			if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u_live, live); err != nil {
				return nil, wrapError("unable to convert unstructured back to live object", key, gvk, err)
			}

			// Apply the patch (client.MergeFrom automatically creates an efficient patch)
			if err := cl.Patch(ctx, live, patch); err != nil {
				return nil, wrapError("unable to patch resource", key, gvk, err)
			}
			logger.Info("Resource patched")

		}
	}

	return util.ObjectReference(live, gvk), nil
}

func reconcilePropertiesToUnstructured(live client.Object, ensure []Property, normalizedDesired client.Object,
	gvk schema.GroupVersionKind, logger logr.Logger) (map[string]interface{}, error) {

	// Convert both objects to unstructured for property manipulation
	u_live, err := runtime.DefaultUnstructuredConverter.ToUnstructured(util.SetTypeMeta(live, gvk))
	if err != nil {
		return nil, fmt.Errorf("unable to convert live object to unstructured: %w", err)
	}

	u_normalizedDesired, err := runtime.DefaultUnstructuredConverter.ToUnstructured(normalizedDesired)
	if err != nil {
		return nil, fmt.Errorf("unable to convert normalized desired to unstructured: %w", err)
	}

	// Apply each ensure property using the same reconciliation logic as CreateOrUpdate
	for _, property := range ensure {
		if err := property.reconcile(u_live, u_normalizedDesired, logger); err != nil {
			return nil, fmt.Errorf("unable to reconcile property %s: %w", property, err)
		}
	}

	return u_live, nil
}

// normalize creates a "normalized" version of a Kubernetes object for comparison purposes.
// This function is critical for the reconciliation process as it:
//  1. Extracts only the properties specified in the 'ensure' list
//  2. Removes properties specified in the 'ignore' list
//  3. Returns a new object containing only the relevant fields for comparison
//
// This normalization allows for precise control over what gets compared during reconciliation,
// enabling users to:
//   - Focus only on specific fields (metadata.labels, spec.replicas, etc.)
//   - Ignore fields managed by other controllers (status, revision annotations, etc.)
//   - Avoid unnecessary updates when irrelevant fields change
//
// Parameters:
//   - o: the source object to normalize
//   - ensure: list of JSONPath properties to include in the normalized object
//   - ignore: list of JSONPath properties to exclude from the normalized object
//   - gvk: GroupVersionKind for creating the output object
//   - s: runtime scheme for object conversion
func normalize(o client.Object, ensure, ignore []Property,
	gvk schema.GroupVersionKind, s *runtime.Scheme) (client.Object, error) {

	// Convert the input object to unstructured format for JSONPath manipulation
	in, err := runtime.DefaultUnstructuredConverter.ToUnstructured(o)
	if err != nil {
		return nil, err
	}
	u_normalized := map[string]any{}

	// STEP 1: Extract only the properties specified in the 'ensure' list
	// For each ensure property, use JSONPath to extract the value and add it to the normalized object
	for _, p := range ensure {
		expr, err := jp.ParseString(p.jsonPath())
		if err != nil {
			return nil, fmt.Errorf("unable to parse JSONPath '%s': %w", p.jsonPath(), err)
		}
		val := expr.Get(in)

		switch len(val) {
		case 0:
			// Property doesn't exist in source, skip it
			continue
		case 1:
			// Property exists, add it to the normalized object
			if err := expr.Set(u_normalized, val[0]); err != nil {
				return nil, fmt.Errorf("usable to add value '%v' in JSONPath '%s'", val[0], p.jsonPath())
			}
		default:
			// JSONPath returned multiple values, which is not supported for ensure properties
			return nil, fmt.Errorf("multi-valued JSONPath (%s) not supported for 'ensure' properties", p.jsonPath())
		}

	}

	// STEP 2: Remove properties specified in the 'ignore' list
	// This allows ignoring specific nested properties within broader 'ensure' properties
	// Example: ensure "spec" but ignore "spec.clusterIP"
	for _, p := range ignore {
		expr, err := jp.ParseString(p.jsonPath())
		if err != nil {
			return nil, fmt.Errorf("unable to parse JSONPath '%s': %w", p.jsonPath(), err)
		}
		if err = expr.Del(u_normalized); err != nil {
			return nil, fmt.Errorf("unable to parse delete JSONPath '%s' from unstructured: %w", p.jsonPath(), err)
		}
	}

	// STEP 3: Convert the normalized unstructured data back to a typed object
	normalized, err := util.NewObjectFromGVK(gvk, s)
	if err != nil {
		return nil, err
	}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u_normalized, normalized); err != nil {
		return nil, err
	}

	return normalized, nil
}

// printfDiff generates a human-readable diff between two objects for logging purposes.
// This helps operators understand exactly what changes were detected during reconciliation.
func printfDiff(a, b client.Object) string {
	ajson, err := json.Marshal(a)
	if err != nil {
		return fmt.Errorf("unable to log differences: %w", err).Error()
	}
	bjson, err := json.Marshal(b)
	if err != nil {
		return fmt.Errorf("unable to log differences: %w", err).Error()
	}
	opts := jsondiff.DefaultJSONOptions()
	opts.SkipMatches = true
	opts.Indent = "\t"
	_, diff := jsondiff.Compare(ajson, bjson, &opts)
	return diff
}

// wrapError provides consistent error formatting with context about the resource being processed.
func wrapError(msg string, key types.NamespacedName, gvk schema.GroupVersionKind, err error) error {
	return fmt.Errorf("%s %s/%s/%s: %w", msg, gvk.Kind, key.Name, key.Namespace, err)
}
