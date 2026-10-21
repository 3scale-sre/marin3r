package resource

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/ohler55/ojg/jp"
	"k8s.io/apimachinery/pkg/api/equality"
)

// propertyDelta represents the different states a property can be in when comparing
// desired vs live objects. This is used to determine what reconciliation action to take.
type propertyDelta int

const (
	// missingInBoth: Property doesn't exist in either desired or live object - no action needed
	missingInBoth propertyDelta = 0
	// missingFromDesiredPresentInLive: Property exists in live but not in desired - delete from live
	missingFromDesiredPresentInLive propertyDelta = 1
	// presentInDesiredMissingFromLive: Property exists in desired but not in live - add to live
	presentInDesiredMissingFromLive propertyDelta = 2
	// presentInBoth: Property exists in both - compare values and update if different
	presentInBoth propertyDelta = 3
)

// Property represents a JSONPath to a field in a Kubernetes resource that can be:
// 1. Reconciled (ensured) - the field value from desired state is enforced on live state
// 2. Ignored - the field is excluded from comparison/reconciliation even if it's within an ensured path
//
// Examples:
//   - "metadata.labels" - ensures all labels match desired state
//   - "spec.replicas" - ensures replica count matches desired
//   - "spec.template.spec.containers[0].image" - ensures specific container image
//   - "metadata.annotations['deployment.kubernetes.io/revision']" - can be ignored to let k8s manage it
type Property string

// jsonPath returns the string representation of the property for JSONPath operations.
func (p Property) jsonPath() string { return string(p) }

// reconcile performs the actual reconciliation of a single property between live and desired states.
// This is where the magic happens - it handles all the different scenarios that can occur
// when reconciling a specific field:
//
// The function uses JSONPath to:
//  1. Extract the property value from both live and desired unstructured objects
//  2. Determine what action is needed based on presence/absence in each object
//  3. Apply the appropriate change (add, remove, or update) to the live object
//
// This granular approach allows precise control over which fields get reconciled and how.
func (p Property) reconcile(u_live, u_desired map[string]any, _ logr.Logger) error {
	// Parse the JSONPath expression for this property
	expr, err := jp.ParseString(p.jsonPath())
	if err != nil {
		return fmt.Errorf("unable to parse JSONPath '%s': %w", p.jsonPath(), err)
	}

	// Extract values from both desired and live objects using JSONPath
	desiredVal := expr.Get(u_desired)
	liveVal := expr.Get(u_live)

	// Validate that JSONPath queries return at most one value
	// Multi-valued results are not supported for property reconciliation
	if len(desiredVal) > 1 || len(liveVal) > 1 {
		return fmt.Errorf("multi-valued JSONPath (%s) not supported when reconciling properties", p.jsonPath())
	}

	// Determine what reconciliation action is needed based on property presence
	switch delta(len(desiredVal), len(liveVal)) {

	case missingInBoth:
		// Property doesn't exist in either object - nothing to reconcile
		return nil

	case missingFromDesiredPresentInLive:
		// Property exists in live but not in desired state
		// This means we want to remove the property from the live object
		// Example: A label that was removed from the desired state
		if err := expr.Del(u_live); err != nil {
			return fmt.Errorf("usable to delete JSONPath '%s'", p.jsonPath())
		}
		return nil

	case presentInDesiredMissingFromLive:
		// Property exists in desired but not in live state
		// This means we want to add the property to the live object
		// Example: A new label that was added to the desired state
		if err := expr.Set(u_live, desiredVal[0]); err != nil {
			return fmt.Errorf("usable to add value '%v' in JSONPath '%s'", desiredVal[0], p.jsonPath())
		}
		return nil

	case presentInBoth:
		// Property exists in both objects - compare values for differences
		// Only update if the values are actually different (semantic comparison)
		if !equality.Semantic.DeepEqual(desiredVal[0], liveVal[0]) {
			// Values differ, update live object with desired value
			if err := expr.Set(u_live, desiredVal[0]); err != nil {
				return fmt.Errorf("usable to replace value '%v' in JSONPath '%s'", desiredVal[0], p.jsonPath())
			}
			return nil
		}
		// Values are the same, no update needed

	}

	return nil
}

// delta calculates the propertyDelta enum value based on the count of values
// found in desired and live objects. This elegant bit manipulation converts
// the presence/absence pattern into a single enum value:
//
// desired=0, live=0 -> 0<<1 + 0 = 0 (missingInBoth)
// desired=0, live=1 -> 0<<1 + 1 = 1 (missingFromDesiredPresentInLive)
// desired=1, live=0 -> 1<<1 + 0 = 2 (presentInDesiredMissingFromLive)
// desired=1, live=1 -> 1<<1 + 1 = 3 (presentInBoth)
func delta(a, b int) propertyDelta {
	return propertyDelta(a<<1 + b)
}
