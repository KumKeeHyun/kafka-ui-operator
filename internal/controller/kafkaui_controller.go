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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// KafkaUIReconciler reconciles a KafkaUI object
type KafkaUIReconciler struct {
	client.Client
	Scheme               *runtime.Scheme
	deployNamespacedName types.NamespacedName
}

func NewKafkaUIReconciler(client client.Client, scheme *runtime.Scheme, namespace string) (*KafkaUIReconciler, error) {
	kafkauiDeploy := newKafkaUIDeployment(namespace)
	if err := client.Create(context.Background(), kafkauiDeploy); err != nil && !errors.IsAlreadyExists(err) {
		return nil, err
	}
	return &KafkaUIReconciler{
		Client: client,
		Scheme: scheme,
		deployNamespacedName: types.NamespacedName{
			Namespace: namespace,
			Name:      "kafka-ui",
		},
	}, nil
}

func newKafkaUIDeployment(namespace string) *appsv1.Deployment {
	kafkauiDeploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kafka-ui",
			Namespace: namespace,
			Labels: map[string]string{
				"app/name":                  "kafka-ui",
				"app.kubernetes.io/part-of": "kafka-ui-operator",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas:             pointer.Int32(2),
			RevisionHistoryLimit: pointer.Int32(3),
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app/name":                  "kafka-ui",
					"app.kubernetes.io/part-of": "kafka-ui-operator",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kafka-ui",
					Namespace: namespace,
					Labels: map[string]string{
						"app/name":                  "kafka-ui",
						"app.kubernetes.io/part-of": "kafka-ui-operator",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Image:           "ghcr.io/kafbat/kafka-ui:v1.0.0",
							ImagePullPolicy: v1.PullIfNotPresent,
							Name:            "kafka-ui",
							Ports:           []v1.ContainerPort{{ContainerPort: 8080}},
							Env: []v1.EnvVar{
								{
									Name:  "SPRING_CONFIG_ADDITIONAL-LOCATION",
									Value: "optional:/etc/kafka-ui/cluster-config/config.yml,optional:/etc/kafka-ui/role-config/config.yml,optional:/etc/kafka-ui/auth-config/config.yml",
								},
							},
							LivenessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									HTTPGet: &v1.HTTPGetAction{
										Path: "/actuator/health",
										Port: intstr.FromInt32(8080),
									},
								},
								InitialDelaySeconds: 5,
								TimeoutSeconds:      3,
								PeriodSeconds:       5,
								SuccessThreshold:    1,
								FailureThreshold:    3,
							},
							ReadinessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									HTTPGet: &v1.HTTPGetAction{
										Path: "/actuator/health",
										Port: intstr.FromInt32(8080),
									},
								},
								InitialDelaySeconds: 5,
								TimeoutSeconds:      5,
								PeriodSeconds:       5,
								SuccessThreshold:    1,
								FailureThreshold:    5,
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "cluster-config-volume",
									ReadOnly:  true,
									MountPath: "/etc/kafka-ui/cluster-config",
								},
								{
									Name:      "role-config-volume",
									ReadOnly:  true,
									MountPath: "/etc/kafka-ui/role-config",
								},
								{
									Name:      "auth-config-volume",
									ReadOnly:  true,
									MountPath: "/etc/kafka-ui/auth-config",
								},
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: "cluster-config-volume",
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: "role-config-volume",
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: "auth-config-volume",
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									SecretName: "kafka-ui-auth-secret",
									Items: []v1.KeyToPath{
										{
											Key:  "kafka-ui-auth.yaml",
											Path: "config.yml",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return kafkauiDeploy
}

//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch

// SetupWithManager sets up the controller with the Manager.
func (r *KafkaUIReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.ConfigMap{}).
		Complete(r)
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *KafkaUIReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var err error
	logger := log.FromContext(ctx)

	// TODO(user): your logic here

	kafkaUIDeploy := &appsv1.Deployment{}
	err = r.Get(ctx, r.deployNamespacedName, kafkaUIDeploy)
	if err != nil {
		return ctrl.Result{}, err
	}

	configMaps := &v1.ConfigMapList{}
	err = r.Client.List(ctx, configMaps, client.InNamespace(r.deployNamespacedName.Namespace))
	if err != nil {
		return ctrl.Result{}, err
	}
	if len(configMaps.Items) < 1 {
		return ctrl.Result{}, nil
	}

	requireUpdateDeploy := false

	latestClusterConfigMap := findLatestClusterConfigMapOrEmpty(configMaps)
	currentClusterVolume := findClusterConfigVolume(kafkaUIDeploy)
	if currentClusterVolume != nil && latestClusterConfigMap.Name != "" {
		if currentClusterVolume.ConfigMap == nil || currentClusterVolume.ConfigMap.Name != latestClusterConfigMap.Name {
			logger.Info("replace cluster config volume", "Namespace", latestClusterConfigMap.GetNamespace(), "Name", latestClusterConfigMap.GetName())
			replaceClusterConfigVolume(kafkaUIDeploy, latestClusterConfigMap)
			requireUpdateDeploy = true
		}
	}

	latestRoleConfigMap := findLatestRoleConfigMapOrEmpty(configMaps)
	currentRoleVolume := findRoleConfigVolume(kafkaUIDeploy)
	if currentRoleVolume != nil && latestRoleConfigMap.Name != "" {
		if currentRoleVolume.ConfigMap == nil || currentRoleVolume.ConfigMap.Name != latestRoleConfigMap.Name {
			logger.Info("replace role config volume", "Namespace", latestRoleConfigMap.GetNamespace(), "Name", latestRoleConfigMap.GetName())
			replaceRoleConfigVolume(kafkaUIDeploy, latestRoleConfigMap)
			requireUpdateDeploy = true
		}
	}

	if requireUpdateDeploy {
		logger.Info("update kafka-ui deploy", "Namespace", kafkaUIDeploy.GetNamespace(), "Name", kafkaUIDeploy.GetName())
		err = r.Update(ctx, kafkaUIDeploy)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func findClusterConfigVolume(deployment *appsv1.Deployment) *v1.Volume {
	for _, volume := range deployment.Spec.Template.Spec.Volumes {
		if volume.Name == "cluster-config-volume" {
			return &volume
		}
	}
	return nil
}

func replaceClusterConfigVolume(deployment *appsv1.Deployment, newConfigMap v1.ConfigMap) {
	for i, volume := range deployment.Spec.Template.Spec.Volumes {
		if volume.Name == "cluster-config-volume" {
			deployment.Spec.Template.Spec.Volumes[i] = v1.Volume{
				Name: "cluster-config-volume",
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{Name: newConfigMap.Name},
					},
				},
			}
			return
		}
	}
}

func findRoleConfigVolume(deployment *appsv1.Deployment) *v1.Volume {
	for _, volume := range deployment.Spec.Template.Spec.Volumes {
		if volume.Name == "role-config-volume" {
			return &volume
		}
	}
	return nil
}

func replaceRoleConfigVolume(deployment *appsv1.Deployment, newConfigMap v1.ConfigMap) {
	for i, volume := range deployment.Spec.Template.Spec.Volumes {
		if volume.Name == "role-config-volume" {
			deployment.Spec.Template.Spec.Volumes[i] = v1.Volume{
				Name: "role-config-volume",
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{Name: newConfigMap.Name},
					},
				},
			}
			return
		}
	}
}
