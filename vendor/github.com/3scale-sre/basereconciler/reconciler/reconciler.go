package reconciler

import (
	"context"
	"fmt"
	"time"

	"github.com/3scale-sre/basereconciler/config"
	"github.com/3scale-sre/basereconciler/resource"
	"github.com/3scale-sre/basereconciler/util"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type action string

const (
	ContinueAction         action = "Continue"
	ReturnAction           action = "Return"
	ReturnAndRequeueAction action = "ReturnAndRequeue"
)

type Result struct {
	Action       action
	RequeueAfter time.Duration
	Error        error
}

func (result Result) ShouldReturn() bool {
	return result.Action == ReturnAction || result.Action == ReturnAndRequeueAction || result.Error != nil
}

func (result Result) Values() (ctrl.Result, error) {

	return ctrl.Result{
			Requeue:      func() bool { return result.Action == ReturnAndRequeueAction }(),
			RequeueAfter: result.RequeueAfter,
		},
		result.Error
}

type lifecycleOptions struct {
	initializationLogic         []initializationFunction
	inMemoryinitializationLogic []inMemoryinitializationFunction
	finalizer                   *string
	finalizationLogic           []finalizationFunction
}

func newLifecycleOptions() *lifecycleOptions {
	return &lifecycleOptions{finalizationLogic: []finalizationFunction{}}
}

// lifecycleOption is an interface that defines options that can be passed to
// the reconciler's ManageResourceLifecycle() function
type lifecycleOption interface {
	applyToLifecycleOptions(*lifecycleOptions)
}

type finalizer string

func (f finalizer) applyToLifecycleOptions(opts *lifecycleOptions) {
	opts.finalizer = ptr.To(string(f))
	opts.initializationLogic = append(opts.initializationLogic, f.initFinalizer)
}

// WithFinalizer can be used to provide a finalizer string that the resource will be initialized with
// For finalization logic to be run before objet deletion, a finalizar must be passed.
func WithFinalizer(f string) finalizer {
	return finalizer(f)
}

func (f finalizer) initFinalizer(ctx context.Context, c client.Client, o client.Object) error {
	if !controllerutil.ContainsFinalizer(o, string(f)) {
		controllerutil.AddFinalizer(o, string(f))
	}
	return nil
}

type finalizationFunction func(context.Context, client.Client) error

func (fn finalizationFunction) applyToLifecycleOptions(opts *lifecycleOptions) {
	opts.finalizationLogic = append(opts.finalizationLogic, fn)
}

// WithFinalizationFunc can be used to provide functions that will be run on object finalization.
// A Finalizer must be set for these functions to be called.
func WithFinalizationFunc(fn func(context.Context, client.Client) error) finalizationFunction {
	return fn
}

type initializationFunction func(context.Context, client.Client, client.Object) error

func (fn initializationFunction) applyToLifecycleOptions(opts *lifecycleOptions) {
	opts.initializationLogic = append(opts.initializationLogic, fn)
}

// WithInitializationFunc can be used to provide functions that run resource initialization, like for example
// applying defaults or labels to the resource.
func WithInitializationFunc(fn func(context.Context, client.Client, client.Object) error) initializationFunction {
	return fn
}

type inMemoryinitializationFunction func(context.Context, client.Client, client.Object) error

func (fn inMemoryinitializationFunction) applyToLifecycleOptions(opts *lifecycleOptions) {
	opts.inMemoryinitializationLogic = append(opts.inMemoryinitializationLogic, fn)
}

// WithInitializationFunc can be used to provide functions that run resource initialization, like for example
// applying defaults or labels to the resource.
func WithInMemoryInitializationFunc(
	fn func(context.Context, client.Client, client.Object) error) inMemoryinitializationFunction {
	return fn
}

// Reconciler computes a list of resources that it needs to keep in place
type Reconciler struct {
	client.Client
	Log         logr.Logger
	Scheme      *runtime.Scheme
	typeTracker typeTracker
	mgr         manager.Manager
}

// NewFromManager returns a new Reconciler from a controller-runtime manager.Manager
func NewFromManager(mgr manager.Manager) *Reconciler {
	return &Reconciler{Client: mgr.GetClient(), Scheme: mgr.GetScheme(), Log: logr.Discard(), mgr: mgr}
}

// WithLogger sets the Reconciler logger
func (r *Reconciler) WithLogger(logger logr.Logger) *Reconciler {
	r.Log = logger
	return r
}

// Logger returns the Reconciler logger and a copy of the context that also includes the logger inside
// to pass it around easily.
func (r *Reconciler) Logger(ctx context.Context, keysAndValues ...interface{}) (context.Context, logr.Logger) {
	var logger logr.Logger
	if !r.Log.IsZero() {
		// get the logger configured in the Reconciler
		logger = r.Log.WithValues(keysAndValues...)
	} else {
		// try to get a logger from the context
		logger = logr.FromContextOrDiscard(ctx).WithValues(keysAndValues...)
	}
	return logr.NewContext(ctx, logger), logger
}

// ManageResourceLifecycle manages the lifecycle of the resource, from initialization to
// finalization and deletion.
// The behaviour can be modified depending on the options passed to the function:
//   - WithInitializationFunc(...): pass a function with initialization logic for the custom resource.
//     The function will be executed and if changes to the custom resource are detected the resource will
//     be updated. It can be used to set default values on the custom resource. Can be used more than once.
//   - WithInMemoryInitializationFunc(...): pass a function with initialization logic to the custom resource.
//     If the custom resource is modified in nay way, the changes won't be persisted in the API server and will
//     only have effect within the reconcile loop. Can be used more than once.
//   - WithFinalizer(...): passes a string that will be configured as a resource finalizar, ensuring that the
//     custom resource has the finalizer in place, updating it if required.
//   - WithFinalizationFunc(...): pass finalization functions that will be
//     run when the custom resource is being deleted. Only works ifa finalizer is also passed, otherwise
//     the custom resource will be immediately deleted and the functions won't run. Can be used more than once.
func (r *Reconciler) ManageResourceLifecycle(ctx context.Context, req reconcile.Request, obj client.Object,
	opts ...lifecycleOption) Result {

	options := newLifecycleOptions()
	for _, o := range opts {
		o.applyToLifecycleOptions(options)
	}

	ctx, logger := r.Logger(ctx)
	err := r.Client.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, obj)
	if err != nil {
		if errors.IsNotFound(err) {
			// Return and don't requeue
			return Result{Action: ReturnAction}
		}
		return Result{Error: err}
	}

	if util.IsBeingDeleted(obj) {

		// finalizer logic is only triggered if the controller
		// sets a finalizer and the finalizer is still present in the
		// resource
		if options.finalizer != nil && controllerutil.ContainsFinalizer(obj, *options.finalizer) {

			err := r.finalize(ctx, options.finalizationLogic)
			if err != nil {
				logger.Error(err, "unable to delete instance")
				return Result{Error: err}
			}
			controllerutil.RemoveFinalizer(obj, *options.finalizer)
			err = r.Client.Update(ctx, obj)
			if err != nil {
				logger.Error(err, "unable to update instance")
				return Result{Error: err}
			}

		}
		// object being deleted, return without doing anything
		// and stop the reconcile loop
		return Result{Action: ReturnAction}
	}

	ok, err := r.isInitialized(ctx, obj, options.initializationLogic)
	if err != nil {
		return Result{Error: err}
	}
	if !ok {
		err := r.Client.Update(ctx, obj)
		if err != nil {
			logger.Error(err, "unable to initialize instance")
			return Result{Error: err}
		}
		return Result{Action: ReturnAndRequeueAction}
	}

	if err := r.inMemoryInitialization(ctx, obj, options.inMemoryinitializationLogic); err != nil {
		return Result{Error: err}
	}

	return Result{Action: ContinueAction}
}

// isInitialized can be used to check if instance is correctly initialized.
// Returns false if it isn't and an update is required.
func (r *Reconciler) isInitialized(ctx context.Context, obj client.Object, fns []initializationFunction) (bool, error) {
	orig := obj.DeepCopyObject()
	for _, fn := range fns {
		err := fn(ctx, r.Client, obj)
		if err != nil {
			return false, err
		}
	}

	if !equality.Semantic.DeepEqual(orig, obj) {
		return false, nil
	}

	return true, nil
}

// inMemoryInitialization can be used to perform initializarion on the resource that is not
// persisted in the API storage. This can be used to perform initialization on the resource without
// writing it to the API to avoid surfacing it uo to the user. This approach is a bit more
// gitops friendly as it avoids modifying the resource, but it doesn't provide any information
// to the user on the initialization being used for reconciliation.
func (r *Reconciler) inMemoryInitialization(
	ctx context.Context, obj client.Object, fns []inMemoryinitializationFunction) error {
	for _, fn := range fns {
		err := fn(ctx, r.Client, obj)
		if err != nil {
			return err
		}
	}
	return nil
}

// finalize contains finalization logic for the Reconciler
func (r *Reconciler) finalize(ctx context.Context, fns []finalizationFunction) error {
	// Call any cleanup functions passed
	for _, fn := range fns {
		err := fn(ctx, r.Client)
		if err != nil {
			return err
		}
	}
	return nil
}

// ReconcileOwnedResources handles generalized resource reconcile logic for a controller:
//
//   - Takes a list of templates and calls resource.CreateOrUpdate on each one of them. The templates
//     need to implement the resource.TemplateInterface interface. Users can take advantage of the generic
//     resource.Template[T] struct that the resource package provides, which already implements the
//     resource.TemplateInterface.
//   - Each template is added to the list of managed resources if resource.CreateOrUpdate returns with no error
//   - If the resource pruner is enabled any resource owned by the custom resource not present in the list of managed
//     resources is deleted. The resource pruner must be enabled in the global config (see package config) and also not
//     explicitly disabled in the resource by the '<annotations-domain>/prune: true/false' annotation.
func (r *Reconciler) ReconcileOwnedResources(
	ctx context.Context, owner client.Object, list []resource.TemplateInterface) Result {

	managedResources := []corev1.ObjectReference{}
	requeue := false

	for _, template := range list {
		ref, err := resource.CreateOrUpdate(ctx, r.Client, r.Scheme, owner, template)
		if err != nil {
			return Result{Error: fmt.Errorf("unable to CreateOrUpdate resource: %w", err)}
		}
		if ref != nil {
			managedResources = append(managedResources, *ref)
			gvk := schema.FromAPIVersionAndKind(ref.APIVersion, ref.Kind)

			// if the resource type is not yet tracked
			if !r.typeTracker.isTypeTracked(gvk) {

				// watch this gvk if feature is enabled
				if config.AreDynamicWatchesEnabled() {
					if err := r.watchOwned(gvk, owner); err != nil {
						return Result{Error: fmt.Errorf("unable to watch GVK %s: %w", gvk, err)}
					}
				}

				// add this GVK to the type tracker
				r.typeTracker.trackType(gvk)
				requeue = true
			}
		}
	}

	if isPrunerEnabled(owner) {
		if err := r.pruneOrphaned(ctx, owner, managedResources); err != nil {

			return Result{Error: fmt.Errorf("unable to prune orphaned resources: %w", err)}
		}
	}

	if requeue {
		return Result{Action: ReturnAndRequeueAction}
	} else {
		return Result{Action: ContinueAction}
	}
}

// FilteredEventHandler returns an EventHandler for the specific client.ObjectList
// passed as parameter. It will produce reconcile requests for any client.Object of the
// given type that returns true when passed to the filter function. If the filter function
// is "nil" all the listed object will receive a reconcile request.
// The filter function receives both the object that generated the event and the object that
// might need to be reconciled in response to that event. Depending on whether it returns true
// or false the reconciler request will be generated or not.
//
// In the following example, a watch for Secret resources which match the name "secret" is added
// to the reconciler. The watch will generate reconmcile requests for v1alpha1.Test resources
// any time a Secret with name "secret" is created/uddated/deleted
//
//	func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
//		return ctrl.NewControllerManagedBy(mgr).
//			For(&v1alpha1.Test{}).
//			Watches(&source.Kind{Type: &corev1.Secret{TypeMeta: metav1.TypeMeta{Kind: "Secret"}}},
//				r.FilteredEventHandler(
//					&v1alpha1.TestList{},
//					func(event, o client.Object) bool {
//						return event.GetName() == "secret"
//					},
//					r.Log)).
//			Complete(r)
//	}
func (r *Reconciler) FilteredEventHandler(ol client.ObjectList,
	filter func(event client.Object, o client.Object) bool, logger logr.Logger) handler.EventHandler {

	return handler.EnqueueRequestsFromMapFunc(
		func(ctx context.Context, event client.Object) []reconcile.Request {
			if err := r.Client.List(ctx, ol); err != nil {
				logger.Error(err, "unable to retrieve the list of resources")
				return []reconcile.Request{}
			}
			items := util.GetItems(ol)
			if len(items) == 0 {
				return []reconcile.Request{}
			}

			req := make([]reconcile.Request, 0, len(items))
			for _, item := range items {
				if filter != nil {
					if !filter(event, item) {
						continue
					}
				}
				req = append(req, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(item)})
			}
			return req
		},
	)
}
