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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

var _ = Describe("KafkaUI Controller", Ordered, func() {
	Context("When reconciling a resource", func() {
		const (
			namespace = "default" // TODO(user):Modify as needed
		)
		var (
			err                  error
			ctx                  context.Context
			controllerReconciler *KafkaUIReconciler
		)

		BeforeAll(func() {
			ctx = context.Background()
			controllerReconciler, err = NewKafkaUIReconciler(k8sClient, k8sClient.Scheme(), namespace)
			Expect(err).To(Succeed())

			err := k8sClient.DeleteAllOf(ctx, &v1.ConfigMap{}, client.InNamespace(namespace))
			Expect(err).To(Succeed())

		})

		It("update cluster-config-volume to new ConfigMap", func() {
			By("creating resource for the Kind KafkaUI")
			resourceName := "cluster-test-01"
			resource := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: namespace,
				},
				Data: map[string]string{
					"config.yml": "test-config",
				},
			}
			Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			// wait for the resource to be cached
			time.Sleep(time.Second)

			By("Reconciling the created resource")
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      resourceName,
					Namespace: namespace,
				},
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking the updated volume")
			kafkaUIDeploy := &appsv1.Deployment{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: "kafka-ui", Namespace: namespace}, kafkaUIDeploy)).To(Succeed())
			clusterConfigVolume := findClusterConfigVolume(kafkaUIDeploy)
			Expect(clusterConfigVolume.ConfigMap).NotTo(BeNil())
			Expect(clusterConfigVolume.ConfigMap.Name).To(Equal(resourceName))
		})

		It("update app-config-volume to another ConfigMap", func() {
			By("creating resource for the Kind KafkaUI")
			resourceName := "cluster-test-02"
			resource := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: namespace,
				},
				Data: map[string]string{
					"config.yml": "test-config",
				},
			}
			Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			// wait for the resource to be cached
			time.Sleep(time.Second)

			By("Reconciling the created resource")
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      resourceName,
					Namespace: namespace,
				},
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking the updated volume")
			kafkaUIDeploy := &appsv1.Deployment{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: "kafka-ui", Namespace: namespace}, kafkaUIDeploy)).To(Succeed())
			clusterConfigVolume := findClusterConfigVolume(kafkaUIDeploy)
			Expect(clusterConfigVolume.ConfigMap).NotTo(BeNil())
			Expect(clusterConfigVolume.ConfigMap.Name).To(Equal(resourceName))
		})
	})
})
