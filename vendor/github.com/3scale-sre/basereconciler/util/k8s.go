package util

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ObjectKey(o client.Object) types.NamespacedName {
	return types.NamespacedName{Name: o.GetName(), Namespace: o.GetNamespace()}
}

// this is an ugly function to retrieve the list of Items from a
// client.ObjectList because the interface doesn't have a GetItems
// method
func GetItems(list client.ObjectList) []client.Object {
	items := []client.Object{}
	values := reflect.ValueOf(list).Elem().FieldByName("Items")
	for i := 0; i < values.Len(); i++ {
		item := values.Index(i)
		if item.Kind() == reflect.Pointer {
			items = append(items, item.Interface().(client.Object))
		} else {
			items = append(items, item.Addr().Interface().(client.Object))
		}
	}

	return items
}

func IsBeingDeleted(o client.Object) bool {
	return !o.GetDeletionTimestamp().IsZero()
}

func NewObjectFromGVK(gvk schema.GroupVersionKind, s *runtime.Scheme) (client.Object, error) {
	o, err := s.New(gvk)
	if err != nil {
		return nil, err
	}
	new, ok := o.(client.Object)
	if !ok {
		return nil, fmt.Errorf("runtime object %T does not implement client.Object", o)
	}
	return new, nil
}

func NewObjectListFromGVK(gvk schema.GroupVersionKind, s *runtime.Scheme) (client.ObjectList, error) {
	if !strings.HasSuffix(gvk.Kind, "List") {
		gvk.Kind = gvk.Kind + "List"
	}
	o, err := s.New(gvk)
	if err != nil {
		return nil, err
	}
	new, ok := o.(client.ObjectList)
	if !ok {
		return nil, fmt.Errorf("runtime object %T does not implement client.ObjectList", o)
	}
	return new, nil
}

func ObjectReference(o client.Object, gvk schema.GroupVersionKind) *corev1.ObjectReference {
	return &corev1.ObjectReference{
		Kind:            gvk.Kind,
		Namespace:       o.GetNamespace(),
		Name:            o.GetName(),
		UID:             o.GetUID(),
		APIVersion:      gvk.GroupVersion().String(),
		ResourceVersion: o.GetResourceVersion(),
	}
}

func SetTypeMeta(o client.Object, gvk schema.GroupVersionKind) client.Object {
	o.GetObjectKind().SetGroupVersionKind(gvk)
	return o
}

// Defaulter defines functions for setting defaults on resources.
type Defaulter interface {
	client.Object
	Default()
}

func ResourceDefaulter(o Defaulter) func(context.Context, client.Client, client.Object) error {
	return func(_ context.Context, _ client.Client, o client.Object) error {
		o.(Defaulter).Default()
		return nil
	}
}
