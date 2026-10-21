package resource

import (
	"context"
	"fmt"
	"reflect"

	"github.com/3scale-sre/basereconciler/config"
	"github.com/3scale-sre/basereconciler/runtimeconfig"
	"github.com/3scale-sre/basereconciler/util"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

// ReconcileOptions defines how a resource should be reconciled during the CreateOrModify process.
// It specifies which properties to manage, which to ignore, and what type of modification
// operation to use when applying changes.
type ReconcileOptions struct {
	// EnsureProperties are the JSONPath properties that should be reconciled from desired to live state.
	// Examples: "metadata.labels", "spec.replicas", "spec.template.spec.containers[0].image"
	EnsureProperties []Property

	// IgnoreProperties are JSONPath properties that should be excluded from reconciliation.
	// Examples: "spec.clusterIP", "metadata.annotations['deployment.kubernetes.io/revision']"
	IgnoreProperties []Property

	// ModifyOp specifies the type of modification operation to use (Update or Patch).
	ModifyOp ModifyOp
}

// TemplateInterface represents a template that defines how a Kubernetes resource should be
// reconciled. It provides methods to:
//  1. Build the desired state of the resource (including applying mutations)
//  2. Determine if the resource should exist (enabled/disabled)
//  3. Configure which properties should be ensured or ignored during reconciliation
//
// This interface allows for flexible resource management where templates can be
// dynamically enabled/disabled and have fine-grained control over reconciliation behavior.
type TemplateInterface interface {
	// Build creates the desired state of the resource by executing the template builder
	// and applying any configured mutations. Mutations have access to the Kubernetes API
	// client to retrieve live cluster state if needed.
	Build(ctx context.Context, cl client.Client, o client.Object) (client.Object, error)

	// Enabled determines whether this resource should exist in the cluster.
	// When false, existing resources will be deleted.
	Enabled() bool

	// GetReconcileOptions returns the reconcile options for this template.
	// This includes the properties to ensure/ignore and the modification operation type.
	GetReconcileOptions() ReconcileOptions

	// GetGVK returns the GVK of the resource this template describes.
	// If the template's GVK field is already set, it returns that value.
	// Otherwise, it infers the GVK from the generic type T using the provided scheme,
	// or the package default scheme if no scheme is provided.
	GetGVK(s ...*runtime.Scheme) schema.GroupVersionKind
}

// TemplateBuilderFunction is a function that returns a Kubernetes API object (client.Object)
// when called. This function creates the base/skeleton of the desired resource.
//
// Key characteristics:
//   - Has no access to cluster live information (use mutations for that)
//   - Should return the basic shape/template of the resource
//   - Can access the owner object if needed for generating names, namespaces, etc.
//   - Is called first during the Build() process
//
// Example:
//
//	func() *corev1.Service {
//	  return &corev1.Service{
//	    ObjectMeta: metav1.ObjectMeta{Name: "my-service", Namespace: "default"},
//	    Spec: corev1.ServiceSpec{...},
//	  }
//	}
type TemplateBuilderFunction[T client.Object] func(client.Object) (T, error)

// TemplateMutationFunction represents mutation functions that require an API client,
// typically because they need to retrieve live cluster information to mutate the object.
//
// Use cases for mutations:
//   - Retrieve existing Service.spec.clusterIP to avoid conflicts
//   - Copy secrets or configmaps from other namespaces
//   - Fetch current replica counts for gradual scaling
//   - Apply live values that shouldn't be overridden by the template
//
// Mutations are applied in order after the TemplateBuilder has created the base object.
// Each mutation can modify any field of the object.
//
// Example:
//
//	func(ctx context.Context, cl client.Client, obj client.Object) error {
//	  svc := obj.(*corev1.Service)
//	  // Retrieve and preserve the existing clusterIP
//	  existing := &corev1.Service{}
//	  if err := cl.Get(ctx, client.ObjectKeyFromObject(svc), existing); err == nil {
//	    svc.Spec.ClusterIP = existing.Spec.ClusterIP
//	  }
//	  return nil
//	}
type TemplateMutationFunction func(context.Context, client.Client, client.Object) error

// Template implements TemplateInterface and provides a flexible way to define
// how Kubernetes resources should be built and reconciled.
//
// Key features:
//   - Automatic GVK inference from the generic type parameter T
//   - Configurable modification operation (Update or Patch)
//   - Fine-grained property-based reconciliation control
//   - Two-phase construction with builder + mutations
//
// The Template follows a two-phase construction process:
//  1. TemplateBuilder creates the base object (no API access)
//  2. TemplateMutations modify the object using live cluster data (with API access)
//
// This separation allows for clean separation of concerns:
//   - Static configuration in the builder
//   - Dynamic/live-dependent configuration in mutations
type Template[T client.Object] struct {
	// GVK is the GVK of the resource this template describes.
	GVK schema.GroupVersionKind

	// TemplateBuilder is the function that creates the base template for the object.
	// It is called first during Build() to create the foundation of the desired state.
	TemplateBuilder TemplateBuilderFunction[T]

	// TemplateMutations are functions that are called during Build() after
	// TemplateBuilder has been invoked. They can modify the object using live
	// cluster information accessible through the Kubernetes API client.
	// Mutations are applied in the order they are defined.
	TemplateMutations []TemplateMutationFunction

	// IsEnabled specifies whether the resource described by this Template should
	// exist in the cluster. When false, the resource will be deleted if it exists.
	IsEnabled bool

	// EnsureProperties are the JSONPath properties from the desired object that should
	// be enforced on the live object. If empty, global configuration is used instead.
	// Examples: ["metadata.labels", "spec.replicas", "spec.template.spec.containers"]
	EnsureProperties []Property

	// IgnoreProperties are JSONPath properties that should not trigger updates,
	// even if they are within an EnsureProperty path. This allows fine-grained control
	// over which fields are managed by this controller vs other controllers.
	// Examples: ["spec.clusterIP", "metadata.annotations['deployment.kubernetes.io/revision']"]
	IgnoreProperties []Property

	// ModifyOp is the modification operation to use when reconciling the resource.
	// Defaults to ModifyOpUpdate if not explicitly set. Use WithModifyOp() to configure.
	ModifyOp ModifyOp
}

// NewTemplate returns a new Template struct using the passed TemplateBuilderFunction.
// The template is enabled by default and automatically infers its GVK from the generic type T.
//
// Scheme handling:
//   - If no scheme is provided, uses the shared default scheme managed by runtimeconfig
//   - If a scheme is provided, uses that specific scheme for GVK inference
//   - This allows for both convenient defaults and explicit overrides when needed
func NewTemplate[T client.Object](tb TemplateBuilderFunction[T], scheme ...*runtime.Scheme) *Template[T] {
	return &Template[T]{
		TemplateBuilder: tb,
		// default to true - resources should exist unless explicitly disabled
		IsEnabled: true,
		GVK:       inferGVKFromType[T](runtimeconfig.SelectScheme(scheme...)),
	}
}

// NewTemplateFromObjectFunction creates a Template from a simple function that returns
// an object. This is a convenience constructor for cases where you don't need access
// to the owner object in the builder function. The template automatically infers its GVK
// from the generic type T.
//
// Scheme handling:
//   - If no scheme is provided, uses the shared default scheme managed by runtimeconfig
//   - If a scheme is provided, uses that specific scheme for GVK inference
//   - This allows for both convenient defaults and explicit overrides when needed
//
// Example:
//
//	NewTemplateFromObjectFunction(func() *corev1.Service {
//	  return &corev1.Service{...}
//	})
func NewTemplateFromObjectFunction[T client.Object](fn func() T, scheme ...*runtime.Scheme) *Template[T] {
	return &Template[T]{
		TemplateBuilder: func(client.Object) (T, error) { return fn(), nil },
		// default to true - resources should exist unless explicitly disabled
		IsEnabled: true,
		GVK:       inferGVKFromType[T](runtimeconfig.SelectScheme(scheme...)),
	}
}

// Build returns the final desired state of the resource by executing the two-phase
// construction process:
//  1. Execute the TemplateBuilder function to create the base object
//  2. Apply each TemplateMutationFunction in order to modify the object
//  3. Return a deep copy to prevent accidental modifications
//
// This process allows templates to be both static (via builder) and dynamic (via mutations).
func (t *Template[T]) Build(ctx context.Context, cl client.Client, o client.Object) (client.Object, error) {
	// Phase 1: Build the base object using the template builder
	o, err := t.TemplateBuilder(o)
	if err != nil {
		return nil, err
	}

	// Phase 2: Apply all mutations in order
	// Each mutation can access the Kubernetes API to retrieve live cluster information
	// This code isn't safe to run if the client is nil
	if cl != nil {
		for _, fn := range t.TemplateMutations {
			if err := fn(ctx, cl, o); err != nil {
				return nil, err
			}
		}
	}

	// Return a deep copy to prevent accidental modifications by callers
	return o.DeepCopyObject().(client.Object), nil
}

// GetReconcileOptions returns the reconcile options for this template.
// If empty, the reconciler will use global default configuration for this resource type.
func (t *Template[T]) GetReconcileOptions() ReconcileOptions {
	opts := ReconcileOptions{}

	// Check if template has explicit modify op
	if t.ModifyOp != "" {
		opts.ModifyOp = t.ModifyOp
	} else {
		// TODO: allow a user to configure this globally like with the ensure/ignore properties
		// default to update to keep backward compatibility
		opts.ModifyOp = ModifyOpUpdate
	}

	// Check if template has explicit configuration
	if len(t.EnsureProperties) > 0 {
		opts.EnsureProperties = t.EnsureProperties
		opts.IgnoreProperties = t.IgnoreProperties
	} else {
		// No explicit config, use global default configuration for this GVK
		cfg, err := config.GetDefaultReconcileConfigForGVK(t.GetGVK())
		if err != nil {
			return ReconcileOptions{}
		}
		opts.EnsureProperties = util.ConvertStringSlice[string, Property](cfg.EnsureProperties)
		opts.IgnoreProperties = util.ConvertStringSlice[string, Property](cfg.IgnoreProperties)
	}

	return opts
}

// Enabled indicates if the resource should be present in the cluster or not.
// When false, existing resources will be deleted during reconciliation.
func (t *Template[T]) Enabled() bool {
	return t.IsEnabled
}

// GetGVK returns the GVK of the resource this template describes.
func (t *Template[T]) GetGVK(s ...*runtime.Scheme) schema.GroupVersionKind {
	if t.GVK != (schema.GroupVersionKind{}) {
		return t.GVK
	}
	gvk := inferGVKFromType[T](runtimeconfig.SelectScheme(s...))
	return gvk
}

// WithMutation adds a single TemplateMutationFunction to the template.
// Mutations are applied in the order they are added.
// Returns the template for method chaining.
func (t *Template[T]) WithMutation(fn TemplateMutationFunction) *Template[T] {
	if t.TemplateMutations == nil {
		t.TemplateMutations = []TemplateMutationFunction{fn}
	} else {
		t.TemplateMutations = append(t.TemplateMutations, fn)
	}
	return t
}

// WithMutations adds multiple TemplateMutationFunctions to the template.
// Each mutation is added in the order provided.
// Returns the template for method chaining.
func (t *Template[T]) WithMutations(fns []TemplateMutationFunction) *Template[T] {
	for _, fn := range fns {
		t.WithMutation(fn)
	}
	return t
}

// WithEnabled sets whether the resource should exist in the cluster.
// Returns the template for method chaining.
func (t *Template[T]) WithEnabled(enabled bool) *Template[T] {
	t.IsEnabled = enabled
	return t
}

// WithEnsureProperties sets the list of JSONPath properties that should be reconciled.
// This overrides any global default configuration for this template.
// Returns the template for method chaining.
func (t *Template[T]) WithEnsureProperties(ensure []Property) *Template[T] {
	t.EnsureProperties = ensure
	return t
}

// WithIgnoreProperties sets the list of JSONPath properties that should be ignored
// during reconciliation. Returns the template for method chaining.
func (t *Template[T]) WithIgnoreProperties(ignore []Property) *Template[T] {
	t.IgnoreProperties = ignore
	return t
}

// WithModifyOp sets the modification operation type (Update or Patch) to use when
// reconciling this resource. Returns the template for method chaining.
func (t *Template[T]) WithModifyOp(op ModifyOp) *Template[T] {
	t.ModifyOp = op
	return t
}

// Apply chains template functions to make them composable. This allows you to
// create reusable template transformations that can be applied to existing templates.
//
// Example:
//
//	baseTemplate := NewTemplate(createBaseService)
//	prodTemplate := baseTemplate.Apply(addProductionLabels).Apply(setHighAvailability)
//
// The transformation function receives the output of the current TemplateBuilder and
// can modify it before returning. Note: this is different from TemplateMutationFunction
// which operates on live cluster data - this operates purely on the template level.
func (t *Template[T]) Apply(transformation TemplateBuilderFunction[T]) *Template[T] {

	fn := t.TemplateBuilder
	t.TemplateBuilder = func(in client.Object) (T, error) {
		// First execute the existing template builder
		o, err := fn(in)
		if err != nil {
			return o, err
		}
		// Then apply the transformation to augment the result
		return transformation(o)
	}

	return t
}

func inferGVKFromType[T client.Object](scheme *runtime.Scheme) schema.GroupVersionKind {
	// Create a zero value of the generic type
	var zero T
	objType := reflect.TypeOf(zero)

	// Handle pointer types
	if objType.Kind() == reflect.Ptr {
		objType = objType.Elem()
	}

	// Create a new instance
	obj := reflect.New(objType).Interface().(client.Object)

	// Use controller-runtime's built-in GVK inference
	gvk, err := apiutil.GVKForObject(obj, scheme)
	if err != nil {
		panic(fmt.Errorf("failed to infer GVK for type %T: %w", (*T)(nil), err))
	}
	return gvk
}

// ExtractGVK is a convenience function that filters and extracts Template[T] instances
// from a slice of TemplateInterface objects. It performs type assertion to identify
// templates that match the specified generic type T and returns pointers to those templates.
//
// This function is useful when working with heterogeneous collections of templates
// and you need to operate on only those templates that manage a specific resource type.
//
// Type Safety:
//   - Only templates that are actually *Template[T] instances are included
//   - Invalid type assertions are silently skipped (no panics)
//   - Returns pointers to the original templates, allowing mutations
//
// Parameters:
//   - templates: A slice of TemplateInterface objects to filter
//
// Returns:
//   - A slice of *Template[T] pointers containing only the templates that match type T
//   - Empty slice if no templates match the specified type
//
// Example:
//
//	var allTemplates []TemplateInterface = []TemplateInterface{
//	    serviceTemplate,  // *Template[*corev1.Service]
//	    podTemplate,      // *Template[*corev1.Pod]
//	    configTemplate,   // *Template[*corev1.ConfigMap]
//	}
//
//	// Extract only Service templates
//	serviceTemplates := ExtractGVK[*corev1.Service](allTemplates)
//	// Returns: []*Template[*corev1.Service] with only serviceTemplate
//
//	// Extract only Pod templates
//	podTemplates := ExtractGVK[*corev1.Pod](allTemplates)
//	// Returns: []*Template[*corev1.Pod] with only podTemplate
//
//	// Now you can mutate the extracted templates
//	for _, template := range serviceTemplates {
//	    template.WithEnsureProperties([]Property{"spec.ports"})
//	}
func ExtractGVK[T client.Object](templates []TemplateInterface) []*Template[T] {
	out := make([]*Template[T], 0, len(templates))
	for _, template := range templates {
		if realizedTemplate, ok := template.(*Template[T]); ok {
			out = append(out, realizedTemplate)
		}
	}
	return out
}
