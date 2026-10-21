package runtimeconfig

import (
	"sync"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

var (
	schemeMu      sync.RWMutex
	defaultScheme *runtime.Scheme = clientgoscheme.Scheme
	schemeFrozen  bool
)

// DefaultScheme returns the current shared runtime scheme used across the basereconciler
// packages. It is safe for concurrent use.
func DefaultScheme() *runtime.Scheme {
	schemeMu.RLock()
	defer schemeMu.RUnlock()
	return defaultScheme
}

// SetDefaultScheme overrides the shared runtime scheme. Once set, the scheme becomes frozen
// and subsequent calls to EnsureDefaultScheme will not replace it. Passing a nil scheme is a no-op.
func SetDefaultScheme(s *runtime.Scheme) {
	if s == nil {
		return
	}
	schemeMu.Lock()
	defaultScheme = s
	schemeFrozen = true
	schemeMu.Unlock()
}

// EnsureDefaultScheme sets the shared runtime scheme only if it has not been customized yet.
// This provides a first-one-wins behaviour that is useful when wiring a reconciler from a
// controller-runtime manager. Passing a nil scheme is a no-op.
func EnsureDefaultScheme(s *runtime.Scheme) {
	if s == nil {
		return
	}
	schemeMu.Lock()
	if !schemeFrozen {
		defaultScheme = s
		schemeFrozen = true
	}
	schemeMu.Unlock()
}

// SelectScheme returns the first non-nil scheme from the provided list. If none are supplied,
// it falls back to the shared default scheme.
func SelectScheme(explicit ...*runtime.Scheme) *runtime.Scheme {
	if len(explicit) > 0 && explicit[0] != nil {
		return explicit[0]
	}
	return DefaultScheme()
}
