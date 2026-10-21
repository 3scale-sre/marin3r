package resource

import (
	"github.com/3scale-sre/basereconciler/runtimeconfig"
	"k8s.io/apimachinery/pkg/runtime"
)

// GetScheme returns the appropriate runtime scheme to use for GVK inference and other operations.
// If explicit schemes are provided, it returns the first one. Otherwise, it returns the shared
// default scheme managed by the runtimeconfig package.
func GetScheme(s ...*runtime.Scheme) *runtime.Scheme {
	return runtimeconfig.SelectScheme(s...)
}
