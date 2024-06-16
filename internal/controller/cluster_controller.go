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
	kafkauiv1 "github.com/KumKeeHyun/kafka-ui-operator/api/v1"
	"github.com/KumKeeHyun/kafka-ui-operator/pkg/funcutil"
	"github.com/KumKeeHyun/kafka-ui-operator/pkg/kafkaui"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/uuid"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

const (
	finalizerName = "kafka-ui.kumkeehyun.github.com/finalizer"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	ReconcileInterval time.Duration
}

func NewClusterReconciler(client client.Client, scheme *runtime.Scheme) *ClusterReconciler {
	return &ClusterReconciler{
		Client:            client,
		Scheme:            scheme,
		ReconcileInterval: time.Second * 10,
	}
}

//+kubebuilder:rbac:groups=kafka-ui.kumkeehyun.github.com,resources=clusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kafka-ui.kumkeehyun.github.com,resources=clusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kafka-ui.kumkeehyun.github.com,resources=clusters/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kafkauiv1.Cluster{}).
		Complete(r)
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var err error
	logger := log.FromContext(ctx)

	clusters := &kafkauiv1.ClusterList{}
	if err = r.List(ctx, clusters, client.InNamespace(req.Namespace)); err != nil {
		return ctrl.Result{}, err
	}
	if err = r.addFinalizerIfRequired(ctx, logger, clusters.Items); err != nil {
		return ctrl.Result{}, err
	}
	deletedClusters, remainedClusters := funcutil.Split(clusters.Items, r.isDeletedCluster)
	if err = r.removeFinalizer(ctx, logger, deletedClusters); err != nil {
		return ctrl.Result{}, err
	}

	configMaps := &v1.ConfigMapList{}
	if err = r.Client.List(ctx, configMaps, client.InNamespace(req.Namespace)); err != nil {
		return ctrl.Result{}, err
	}
	latestConfigMap := findLatestClusterConfigMapOrEmpty(configMaps)

	if r.isNotRequireNewConfigMap(latestConfigMap, remainedClusters) {
		logger.Info("no need to update config", "Namespace", req.Namespace, "Name", req.Name)
		return ctrl.Result{}, nil
	}

	if r.isCreatedWithInReconcileInterval(latestConfigMap) {
		logger.Info("requeue after minimum interval", "Namespace", req.Namespace, "Name", req.Name)
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: r.ReconcileInterval,
		}, nil
	}

	if err = r.createNewConfigMap(ctx, req, remainedClusters); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ClusterReconciler) addFinalizerIfRequired(ctx context.Context, logger logr.Logger, clusters []kafkauiv1.Cluster) error {
	for _, cluster := range clusters {
		if r.isRequireFinalizer(cluster) {
			logger.Info("add finalizer", "Namespace", cluster.GetNamespace(), "Name", cluster.GetName())
			controllerutil.AddFinalizer(&cluster, finalizerName)
			if err := r.Client.Update(ctx, &cluster); err != nil {
				logger.Error(err, "failed to add finalizer", "Namespace", cluster.GetNamespace(), "Name", cluster.GetName())
				return err
			}
		}
	}
	return nil
}

func (r *ClusterReconciler) isRequireFinalizer(cluster kafkauiv1.Cluster) bool {
	return cluster.DeletionTimestamp.IsZero() && !controllerutil.ContainsFinalizer(&cluster, finalizerName)
}

func (r *ClusterReconciler) isDeletedCluster(cluster kafkauiv1.Cluster) bool {
	return !cluster.DeletionTimestamp.IsZero() && controllerutil.ContainsFinalizer(&cluster, finalizerName)
}

func (r *ClusterReconciler) removeFinalizer(ctx context.Context, logger logr.Logger, deletedClusters []kafkauiv1.Cluster) error {
	for _, deletedCluster := range deletedClusters {
		logger.Info("remove finalizer", "Namespace", deletedCluster.GetNamespace(), "Name", deletedCluster.GetName())
		controllerutil.RemoveFinalizer(&deletedCluster, finalizerName)
		if err := r.Update(ctx, &deletedCluster); err != nil {
			logger.Error(err, "failed to remove finalizer", "Namespace", deletedCluster.GetNamespace(), "Name", deletedCluster.GetName())
			return err
		}
	}
	return nil
}

func (r *ClusterReconciler) isNotRequireNewConfigMap(latestConfigMap v1.ConfigMap, clusters []kafkauiv1.Cluster) bool {
	kafkaProperty := &kafkaui.KafkaProperties{}
	kafkaProperty.MustUnmarshalFromYaml(latestConfigMap.Data["config.yml"])
	return kafkaProperty.MatchWithClusters(clusters)
}

func (r *ClusterReconciler) isCreatedWithInReconcileInterval(latestConfigMap v1.ConfigMap) bool {
	return latestConfigMap.CreationTimestamp.Add(r.ReconcileInterval).After(time.Now())
}

func (r *ClusterReconciler) createNewConfigMap(ctx context.Context, req ctrl.Request, clusters []kafkauiv1.Cluster) error {
	newConfigMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-" + string(uuid.NewUUID()),
			Namespace: req.Namespace,
		},
		Data: map[string]string{
			"config.yml": kafkaui.NewKafkaPropertiesFromClusters(clusters).MustMarshalToYaml(),
		},
	}
	return r.Create(ctx, newConfigMap)
}
