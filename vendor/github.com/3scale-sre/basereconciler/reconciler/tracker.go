package reconciler

import (
	"reflect"
	"sync"

	"github.com/3scale-sre/basereconciler/util"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// ReconcilerWithTypeTracker is a reconciler with a TypeTracker
// The type tracker is used by the "resource pruner" and "dynamic watches"
// features
type ReconcilerWithTypeTracker interface {
	reconcile.Reconciler
	BuildTypeTracker(ctrl controller.Controller)
}

// SetupWithDynamicTypeWatches is a helper to build a controller that can watch resource
// types dynamically. It is typically used within the "SetupWithManager" function.
// Example usage:
//
//	func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
//		return reconciler.SetupWithDynamicTypeWatches(r,
//			ctrl.NewControllerManagedBy(mgr).
//				For(&v1alpha1.Test{}).
//	            // add any other watches here
//				Watches(...}.Watches(...),
//		)
//	}
func SetupWithDynamicTypeWatches(r ReconcilerWithTypeTracker, bldr *builder.Builder) error {
	ctrl, err := bldr.Build(r)
	if err != nil {
		return err
	}
	r.BuildTypeTracker(ctrl)
	return nil
}

type typeTracker struct {
	seenTypes []schema.GroupVersionKind
	ctrl      controller.Controller
	mu        sync.Mutex
}

func (tt *typeTracker) trackType(gvk schema.GroupVersionKind) {
	if !util.ContainsBy(tt.seenTypes, func(x schema.GroupVersionKind) bool {
		return reflect.DeepEqual(x, gvk)
	}) {
		tt.mu.Lock()
		defer tt.mu.Unlock()
		tt.seenTypes = append(tt.seenTypes, gvk)
	}
}

func (tt *typeTracker) isTypeTracked(gvk schema.GroupVersionKind) bool {
	return util.ContainsBy(tt.seenTypes, func(x schema.GroupVersionKind) bool {
		return reflect.DeepEqual(x, gvk)
	})
}

func (r *Reconciler) watchOwned(gvk schema.GroupVersionKind, owner client.Object) error {
	o, err := util.NewObjectFromGVK(gvk, r.Scheme)
	if err != nil {
		return err
	}
	r.typeTracker.mu.Lock()
	defer r.typeTracker.mu.Unlock()
	err = r.typeTracker.ctrl.Watch(source.Kind(
		r.mgr.GetCache(), o,
		handler.EnqueueRequestForOwner(r.mgr.GetScheme(), r.mgr.GetRESTMapper(), owner, handler.OnlyControllerOwner()),
	))
	if err != nil {
		return err
	}
	return nil
}

// BuildTypeTracker passes the controller to the reconciler so watches
// can be added dynamically
func (r *Reconciler) BuildTypeTracker(ctrl controller.Controller) {
	r.typeTracker = typeTracker{
		seenTypes: []schema.GroupVersionKind{},
		ctrl:      ctrl,
	}
}
