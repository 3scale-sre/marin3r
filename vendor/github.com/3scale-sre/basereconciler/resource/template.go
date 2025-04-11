package resource

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TemplateInterface represents a template that can has methods that instruct how a certain
// resource needs to be progressed to match its desired state.
type TemplateInterface interface {
	Build(ctx context.Context, cl client.Client, o client.Object) (client.Object, error)
	Enabled() bool
	GetEnsureProperties() []Property
	GetIgnoreProperties() []Property
}

// TemplateBuilderFunction is a function that returns a k8s API object (client.Object) when
// called. TemplateBuilderFunction has no access to cluster live info.
// A TemplateBuilderFunction is used to return the basic shape of a resource (a template) that can
// then be further modified before it's compared with it's live state and reconciled.
type TemplateBuilderFunction[T client.Object] func(client.Object) (T, error)

// TemplateMutationFunction represents mutation functions that require an API client, generally
// because they need to retrieve live cluster information to mutate the object.
// A TemplateMutationFunction is typically used to modify a template using live values obtained from
// a kubernetes API server.
type TemplateMutationFunction func(context.Context, client.Client, client.Object) error

// Template implements TemplateInterface.
type Template[T client.Object] struct {
	// TemplateBuilder is the function that is used as the basic
	// template for the object. It is called by Build() to create the
	// object.
	TemplateBuilder TemplateBuilderFunction[T]
	// TemplateMutations are functions that are called during Build() after
	// TemplateBuilder has been invoked, to perform mutations on the object that require
	// access to a kubernetes API server.
	TemplateMutations []TemplateMutationFunction
	// IsEnabled specifies whether the resource described by this Template should
	// exist or not.
	IsEnabled bool
	// EnsureProperties are the properties from the desired object that should be enforced
	// to the live object. The syntax is jsonpath.
	EnsureProperties []Property
	// IgnoreProperties are the properties from the live object that should not trigger
	// updates. This is used to ignore nested properties within the "EnsuredProperties". The
	// syntax is jsonpath.
	IgnoreProperties []Property
}

// NewTemplate returns a new Template struct using the passed parameters
func NewTemplate[T client.Object](tb TemplateBuilderFunction[T]) *Template[T] {
	return &Template[T]{
		TemplateBuilder: tb,
		// default to true
		IsEnabled: true,
	}
}

// NewTemplateFromObjectFunction returns a new Template using the given kubernetes
// object as the base.
func NewTemplateFromObjectFunction[T client.Object](fn func() T) *Template[T] {
	return &Template[T]{
		TemplateBuilder: func(client.Object) (T, error) { return fn(), nil },
		// default to true
		IsEnabled: true,
	}
}

// Build returns a T resource. It first executes the TemplateBuilder function and then each of the
// TemplateMutationFunction functions specified by the TemplateMutations field.
func (t *Template[T]) Build(ctx context.Context, cl client.Client, o client.Object) (client.Object, error) {
	o, err := t.TemplateBuilder(o)
	if err != nil {
		return nil, err
	}
	for _, fn := range t.TemplateMutations {
		if err := fn(ctx, cl, o); err != nil {
			return nil, err
		}
	}
	return o.DeepCopyObject().(client.Object), nil
}

// Enabled indicates if the resource should be present or not
func (t *Template[T]) Enabled() bool {
	return t.IsEnabled
}

// GetEnsureProperties returns the list of properties that should be reconciled
func (t *Template[T]) GetEnsureProperties() []Property {
	return t.EnsureProperties
}

// GetIgnoreProperties returns the list of properties that should be ignored
func (t *Template[T]) GetIgnoreProperties() []Property {
	return t.IgnoreProperties
}

func (t *Template[T]) WithMutation(fn TemplateMutationFunction) *Template[T] {
	if t.TemplateMutations == nil {
		t.TemplateMutations = []TemplateMutationFunction{fn}
	} else {
		t.TemplateMutations = append(t.TemplateMutations, fn)
	}
	return t
}

func (t *Template[T]) WithMutations(fns []TemplateMutationFunction) *Template[T] {
	for _, fn := range fns {
		t.WithMutation(fn)
	}
	return t
}

func (t *Template[T]) WithEnabled(enabled bool) *Template[T] {
	t.IsEnabled = enabled
	return t
}

func (t *Template[T]) WithEnsureProperties(ensure []Property) *Template[T] {
	t.EnsureProperties = ensure
	return t
}

func (t *Template[T]) WithIgnoreProperties(ignore []Property) *Template[T] {
	t.IgnoreProperties = ignore
	return t
}

// Apply chains template functions to make them composable
func (t *Template[T]) Apply(mutation TemplateBuilderFunction[T]) *Template[T] {

	fn := t.TemplateBuilder
	t.TemplateBuilder = func(in client.Object) (T, error) {
		o, err := fn(in)
		if err != nil {
			return o, err
		}
		return mutation(o)
	}

	return t
}
