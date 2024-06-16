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
	"github.com/KumKeeHyun/kafka-ui-operator/pkg/funcutil"
	"github.com/KumKeeHyun/kafka-ui-operator/pkg/kafkaui"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kafkauiv1 "github.com/KumKeeHyun/kafka-ui-operator/api/v1"
)

// RoleReconciler reconciles a Role object
type RoleReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	ReconcileInterval time.Duration
}

func NewRoleReconciler(client client.Client, scheme *runtime.Scheme) *RoleReconciler {
	return &RoleReconciler{
		Client:            client,
		Scheme:            scheme,
		ReconcileInterval: time.Second * 10,
	}
}

//+kubebuilder:rbac:groups=kafka-ui.kumkeehyun.github.com,resources=roles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kafka-ui.kumkeehyun.github.com,resources=roles/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kafka-ui.kumkeehyun.github.com,resources=roles/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create

// SetupWithManager sets up the controller with the Manager.
func (r *RoleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kafkauiv1.Role{}).
		Complete(r)
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *RoleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var err error
	logger := log.FromContext(ctx)

	roles := &kafkauiv1.RoleList{}
	if err = r.List(ctx, roles, client.InNamespace(req.Namespace)); err != nil {
		return ctrl.Result{}, err
	}
	if err = r.addFinalizerIfRequired(ctx, logger, roles.Items); err != nil {
		return ctrl.Result{}, err
	}
	deletedRoles, remainedRoles := funcutil.Split(roles.Items, func(role kafkauiv1.Role) bool { return isDeletedObject(&role) })
	if err = r.removeFinalizer(ctx, logger, deletedRoles); err != nil {
		return ctrl.Result{}, err
	}

	configMaps := &v1.ConfigMapList{}
	if err = r.Client.List(ctx, configMaps, client.InNamespace(req.Namespace)); err != nil {
		return ctrl.Result{}, err
	}
	latestConfigMap := findLatestRoleConfigMapOrEmpty(configMaps)

	if r.isNotRequireNewConfigMap(latestConfigMap, remainedRoles) {
		logger.Info("no need to update config", "Namespace", req.Namespace, "Name", req.Name)
		return ctrl.Result{}, nil
	}

	if isCreatedWithInReconcileInterval(latestConfigMap, r.ReconcileInterval) {
		logger.Info("requeue after minimum interval", "Namespace", req.Namespace, "Name", req.Name)
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: r.ReconcileInterval,
		}, nil
	}

	if err = r.createNewConfigMap(ctx, req, remainedRoles); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *RoleReconciler) addFinalizerIfRequired(ctx context.Context, logger logr.Logger, roles []kafkauiv1.Role) error {
	for _, role := range roles {
		if isObjectRequireFinalizer(&role) {
			logger.Info("add finalizer", "Namespace", role.GetNamespace(), "Name", role.GetName())
			controllerutil.AddFinalizer(&role, finalizerName)
			if err := r.Client.Update(ctx, &role); err != nil {
				logger.Error(err, "failed to add finalizer", "Namespace", role.GetNamespace(), "Name", role.GetName())
				return err
			}
		}
	}
	return nil
}

func (r *RoleReconciler) removeFinalizer(ctx context.Context, logger logr.Logger, deletedRoles []kafkauiv1.Role) error {
	for _, deletedRole := range deletedRoles {
		logger.Info("remove finalizer", "Namespace", deletedRole.GetNamespace(), "Name", deletedRole.GetName())
		controllerutil.RemoveFinalizer(&deletedRole, finalizerName)
		if err := r.Update(ctx, &deletedRole); err != nil {
			logger.Error(err, "failed to remove finalizer", "Namespace", deletedRole.GetNamespace(), "Name", deletedRole.GetName())
			return err
		}
	}
	return nil
}

func (r *RoleReconciler) isNotRequireNewConfigMap(latestConfigMap v1.ConfigMap, remainedRoles []kafkauiv1.Role) bool {
	rbacProperties := &kafkaui.RBACProperties{}
	rbacProperties.MustUnmarshalFromYaml(latestConfigMap.Data["config.yml"])
	return rbacProperties.MatchWithRoles(remainedRoles)
}

func (r *RoleReconciler) createNewConfigMap(ctx context.Context, req ctrl.Request, roles []kafkauiv1.Role) error {
	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "role-" + string(uuid.NewUUID()),
			Namespace: req.Namespace,
		},
		Data: map[string]string{
			"config.yml": kafkaui.NewRBACPropertiesFromRoles(roles).MustMarshalToYaml(),
		},
	}
	return r.Create(ctx, configMap)
}

func isObjectRequireFinalizer(object client.Object) bool {
	return object.GetDeletionTimestamp().IsZero() && !controllerutil.ContainsFinalizer(object, finalizerName)
}

func isDeletedObject(object client.Object) bool {
	return !object.GetDeletionTimestamp().IsZero() && controllerutil.ContainsFinalizer(object, finalizerName)
}

func isCreatedWithInReconcileInterval(latestConfigMap v1.ConfigMap, reconcileInterval time.Duration) bool {
	return latestConfigMap.CreationTimestamp.Add(reconcileInterval).After(time.Now())
}
