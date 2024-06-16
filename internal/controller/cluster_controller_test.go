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
	"fmt"
	kafkauiv1 "github.com/KumKeeHyun/kafka-ui-operator/api/v1"
	"github.com/KumKeeHyun/kafka-ui-operator/pkg/kafkaui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

var _ = Describe("Cluster Controller", Ordered, func() {
	Context("When reconciling a resource", func() {
		const (
			namespace = "default" // TODO(user):Modify as needed
		)
		var (
			ctx                  context.Context
			controllerReconciler *ClusterReconciler
		)

		BeforeAll(func() {
			ctx = context.Background()
			controllerReconciler = &ClusterReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}
			err := k8sClient.DeleteAllOf(ctx, &kafkauiv1.Cluster{}, client.InNamespace(namespace))
			Expect(err).To(Succeed())

			err = k8sClient.DeleteAllOf(ctx, &v1.ConfigMap{}, client.InNamespace(namespace))
			Expect(err).To(Succeed())
		})

		It("create ConfigMap for initial Cluster resource", func() {
			By("creating resource for the Kind Cluster")
			resourceName := "test-cluster-00"
			resource := &kafkauiv1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: namespace,
				},
				Spec: kafkauiv1.ClusterSpec{
					ClusterProperties: kafkauiv1.ClusterProperties{
						Name:             resourceName,
						BootstrapServers: "localhost:9092",
					},
				},
			}
			Expect(k8sClient.Create(ctx, resource)).To(Succeed())

			By("Reconciling the created resource")
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      resourceName,
					Namespace: namespace,
				},
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking the created ConfigMap")
			configMaps := &v1.ConfigMapList{}
			Expect(k8sClient.List(ctx, configMaps, client.InNamespace(namespace))).To(Succeed())
			Expect(configMaps.Items).NotTo(BeEmpty())
			latestConfigMap := findLatestClusterConfigMapOrEmpty(configMaps)

			kafkaProperty := &kafkaui.KafkaProperties{}
			kafkaProperty.MustUnmarshalFromYaml(latestConfigMap.Data["config.yml"])

			clusters := &kafkauiv1.ClusterList{}
			Expect(k8sClient.List(ctx, clusters, client.InNamespace(namespace))).To(Succeed())
			Expect(clusters.Items).To(HaveLen(1))

			Expect(kafkaProperty.MatchWithClusters(clusters.Items)).To(BeTrue())
			for _, cluster := range clusters.Items {
				Expect(controllerutil.ContainsFinalizer(&cluster, finalizerName)).To(BeTrue())
			}
		})

		It("create ConfigMap for additional Cluster resource", func() {
			By("creating resource for the Kind Cluster")
			for i := 1; i < 5; i++ {
				clusterName := fmt.Sprintf("test-cluster-%02d", i)
				resource := &kafkauiv1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      clusterName,
						Namespace: namespace,
					},
					Spec: kafkauiv1.ClusterSpec{
						ClusterProperties: kafkauiv1.ClusterProperties{
							Name:             clusterName,
							BootstrapServers: fmt.Sprintf("localhost:%d", 9092+i),
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
			// wait for the resource to be created
			time.Sleep(time.Second)

			By("Reconciling the created resource")
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-cluster-01",
					Namespace: namespace,
				},
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking the created ConfigMap")
			configMaps := &v1.ConfigMapList{}
			Expect(k8sClient.List(ctx, configMaps, client.InNamespace(namespace))).To(Succeed())
			Expect(configMaps.Items).NotTo(BeEmpty())
			latestConfigMap := findLatestClusterConfigMapOrEmpty(configMaps)

			kafkaProperty := &kafkaui.KafkaProperties{}
			kafkaProperty.MustUnmarshalFromYaml(latestConfigMap.Data["config.yml"])

			clusters := &kafkauiv1.ClusterList{}
			Expect(k8sClient.List(ctx, clusters, client.InNamespace(namespace))).To(Succeed())
			Expect(clusters.Items).To(HaveLen(5))

			Expect(kafkaProperty.MatchWithClusters(clusters.Items)).To(BeTrue())
			for _, cluster := range clusters.Items {
				Expect(controllerutil.ContainsFinalizer(&cluster, finalizerName)).To(BeTrue())
			}
		})

		It("create ConfigMap for deleted Cluster resource", func() {
			By("deleting resource for the Kind Cluster")
			resourceName := "test-cluster-00"
			resource := &kafkauiv1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: namespace,
				},
			}
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			// wait for the resource to be created
			time.Sleep(time.Second)

			By("Reconciling the deleted resource")
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      resourceName,
					Namespace: namespace,
				},
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking the created ConfigMap")
			configMaps := &v1.ConfigMapList{}
			Expect(k8sClient.List(ctx, configMaps, client.InNamespace(namespace))).To(Succeed())
			Expect(configMaps.Items).NotTo(BeEmpty())
			latestConfigMap := findLatestClusterConfigMapOrEmpty(configMaps)

			kafkaProperty := &kafkaui.KafkaProperties{}
			kafkaProperty.MustUnmarshalFromYaml(latestConfigMap.Data["config.yml"])

			clusters := &kafkauiv1.ClusterList{}
			Expect(k8sClient.List(ctx, clusters, client.InNamespace(namespace))).To(Succeed())
			Expect(clusters.Items).To(HaveLen(4))

			Expect(kafkaProperty.MatchWithClusters(clusters.Items)).To(BeTrue())
			for _, cluster := range clusters.Items {
				Expect(controllerutil.ContainsFinalizer(&cluster, finalizerName)).To(BeTrue())
			}
		})

		AfterAll(func() {
			err := k8sClient.DeleteAllOf(ctx, &kafkauiv1.Cluster{}, client.InNamespace(namespace))
			Expect(err).To(Succeed())
			err = k8sClient.DeleteAllOf(ctx, &v1.ConfigMap{}, client.InNamespace(namespace))
			Expect(err).To(Succeed())
		})

	})
})
