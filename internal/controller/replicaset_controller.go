/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"strings"
)

// ReplicaSetReconciler reconciles a ReplicaSet object
type ReplicaSetReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps,resources=replicasets,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *ReplicaSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	rs := &appsv1.ReplicaSet{}
	err := r.Get(ctx, req.NamespacedName, rs)
	logger.Info("Reconcile ReplicaSet", "Namespace", req.Namespace, "Name", req.Name)
	if client.IgnoreNotFound(err) != nil {
		return ctrl.Result{}, err
	}

	volumes := rs.Spec.Template.Spec.Volumes
	for _, volume := range volumes {
		if volume.ConfigMap == nil {
			continue
		}
		cm := &v1.ConfigMap{}
		err := r.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: volume.ConfigMap.Name}, cm)
		if client.IgnoreNotFound(err) != nil {
			return ctrl.Result{}, err
		}
		err = controllerutil.SetOwnerReference(rs, cm, r.Scheme)
		if err != nil {
			return ctrl.Result{}, err
		}
		err = r.Update(ctx, cm)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReplicaSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	kafkaUICreatedFilter := predicate.Funcs{CreateFunc: func(e event.CreateEvent) bool {
		if strings.HasPrefix(e.Object.GetName(), "kafka-ui") {
			return true
		}
		return false
	}}
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.ReplicaSet{}, builder.WithPredicates(kafkaUICreatedFilter)).
		Complete(r)
}
