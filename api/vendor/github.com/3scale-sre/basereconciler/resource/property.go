package resource

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/ohler55/ojg/jp"
	"k8s.io/apimachinery/pkg/api/equality"
)

type propertyDelta int

const (
	missingInBoth                   propertyDelta = 0
	missingFromDesiredPresentInLive propertyDelta = 1
	presentInDesiredMissingFromLive propertyDelta = 2
	presentInBoth                   propertyDelta = 3
)

// Property represents a json path to a field in the resource that can
// be either reconciled to ensure it mathes the desired value or can be ignored
// to avoid reconciling certain fields in the rource we are not interested in.
type Property string

func (p Property) jsonPath() string { return string(p) }

func (p Property) reconcile(u_live, u_desired map[string]any, logger logr.Logger) error {
	expr, err := jp.ParseString(p.jsonPath())
	if err != nil {
		return fmt.Errorf("unable to parse JSONPath '%s': %w", p.jsonPath(), err)
	}

	desiredVal := expr.Get(u_desired)
	liveVal := expr.Get(u_live)
	if len(desiredVal) > 1 || len(liveVal) > 1 {
		return fmt.Errorf("multi-valued JSONPath (%s) not supported when reconciling properties", p.jsonPath())
	}

	switch delta(len(desiredVal), len(liveVal)) {

	case missingInBoth:
		// nothing to do
		return nil

	case missingFromDesiredPresentInLive:
		// delete property from u_live
		if err := expr.Del(u_live); err != nil {
			return fmt.Errorf("usable to delete JSONPath '%s'", p.jsonPath())
		}
		return nil

	case presentInDesiredMissingFromLive:
		// add property to u_live
		if err := expr.Set(u_live, desiredVal[0]); err != nil {
			return fmt.Errorf("usable to add value '%v' in JSONPath '%s'", desiredVal[0], p.jsonPath())
		}
		return nil

	case presentInBoth:
		// replace property in u_live if values differ
		if !equality.Semantic.DeepEqual(desiredVal[0], liveVal[0]) {
			if err := expr.Set(u_live, desiredVal[0]); err != nil {
				return fmt.Errorf("usable to replace value '%v' in JSONPath '%s'", desiredVal[0], p.jsonPath())
			}
			return nil
		}

	}

	return nil
}

func delta(a, b int) propertyDelta {
	return propertyDelta(a<<1 + b)
}

func (p Property) ignore(m map[string]any) error {
	expr, err := jp.ParseString(p.jsonPath())
	if err != nil {
		return fmt.Errorf("unable to parse JSONPath '%s': %w", p.jsonPath(), err)
	}
	if err = expr.Del(m); err != nil {
		return fmt.Errorf("unable to parse delete JSONPath '%s' from unstructured: %w", p.jsonPath(), err)
	}
	return nil
}
