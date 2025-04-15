package reconciler

import (
	"context"
	"fmt"
	"strconv"

	"github.com/3scale-sre/basereconciler/config"
	"github.com/3scale-sre/basereconciler/util"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

func (r *Reconciler) pruneOrphaned(ctx context.Context, owner client.Object, managed []corev1.ObjectReference) error {
	logger := logr.FromContextOrDiscard(ctx)

	ownerGVK, err := apiutil.GVKForObject(owner, r.Scheme)
	if err != nil {
		return fmt.Errorf("unable to get GVK for owner: %w", err)
	}

	for _, gvk := range r.typeTracker.seenTypes {

		objectList, err := util.NewObjectListFromGVK(gvk, r.Scheme)
		if err != nil {
			return fmt.Errorf("unable to get list type for '%s': %w", gvk.String(), err)
		}
		err = r.Client.List(ctx, objectList, client.InNamespace(owner.GetNamespace()))
		if err != nil {
			return err
		}

		for _, obj := range util.GetItems(objectList) {

			owned := util.ContainsBy(obj.GetOwnerReferences(), func(ref metav1.OwnerReference) bool {
				return ref.Kind == ownerGVK.Kind &&
					ref.Name == owner.GetName() &&
					ref.APIVersion == ownerGVK.GroupVersion().String()
			})
			managed := util.ContainsBy(managed, func(ref corev1.ObjectReference) bool {
				return ref.Name == obj.GetName() &&
					ref.Namespace == obj.GetNamespace() &&
					ref.Kind == gvk.Kind &&
					ref.APIVersion == gvk.GroupVersion().String()
			})

			// if isOwned(owner, obj) && !util.IsBeingDeleted(obj) && !isManaged(util.ObjectKey(obj), gvk.Kind, managed) {
			if owned && !util.IsBeingDeleted(obj) && !managed {
				err := r.Client.Delete(ctx, obj)
				if err != nil {
					return err
				}
				logger.Info("resource deleted", "kind", gvk.Kind, "resource", obj.GetName())
			}
		}
	}
	return nil
}

func isPrunerEnabled(owner client.Object) bool {
	// prune is active by default
	prune := true

	// get the per resource switch (annotation)
	if value, ok := owner.GetAnnotations()[fmt.Sprintf("%s/prune", config.GetAnnotationsDomain())]; ok {
		var err error
		prune, err = strconv.ParseBool(value)
		if err != nil {
			prune = true
		}
	}
	return prune && config.IsResourcePrunerEnabled()
}
