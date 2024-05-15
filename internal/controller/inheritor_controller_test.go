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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	labelsv1 "github.com/theisferre/label-inheritance-operator/api/v1"
)

var _ = Describe("Inheritor Controller", func() {

	BeforeEach(func() {

	})

	// After each run, delete existing Inheritor resources
	AfterEach(func() {
		inheritorList := &labelsv1.InheritorList{}
		err := k8sClient.List(ctx, inheritorList)
		Expect(err).NotTo(HaveOccurred())
		for _, inheritor := range inheritorList.Items {
			Expect(k8sClient.Delete(ctx, &inheritor)).Should(Succeed())
		}

		podList := &corev1.PodList{}
		err = k8sClient.List(ctx, podList)
		Expect(err).NotTo(HaveOccurred())
		for _, pod := range podList.Items {
			Expect(k8sClient.Delete(ctx, &pod)).Should(Succeed())
		}

		configMapList := &corev1.ConfigMapList{}
		err = k8sClient.List(ctx, configMapList)
		Expect(err).NotTo(HaveOccurred())
		for _, configMap := range configMapList.Items {
			Expect(k8sClient.Delete(ctx, &configMap)).Should(Succeed())
		}
	})

	Context("Update Pod labels", func() {

		It("should update the labels of the pod", func() {

			key := types.NamespacedName{
				Name:      "test-inheritor",
				Namespace: "default",
			}

			spec := labelsv1.InheritorSpec{
				Selectors: []labelsv1.Selector{
					{
						NamespaceSelector: metav1.LabelSelector{
							MatchLabels: map[string]string{
								"kubernetes.io/metadata.name": key.Namespace,
							},
						},
						IncludeLabels: []string{"app"},
					},
				},
			}

			inheritor := &labelsv1.Inheritor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: spec,
			}

			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: key.Namespace,
					Labels: map[string]string{
						"app": "new-label",
					},
				},
			}

			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
					Labels: map[string]string{
						"foo": "bar",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test-container",
							Image: "nginx",
						},
					},
				},
			}

			By("Creating the namespace")
			Expect(k8sClient.Update(ctx, ns)).Should(Succeed())

			By("Creating the pod")
			Expect(k8sClient.Create(ctx, pod)).Should(Succeed())
			time.Sleep(5 * time.Second)

			By("Creating the inheritor")
			Expect(k8sClient.Create(ctx, inheritor)).Should(Succeed())
			time.Sleep(5 * time.Second)

			By("Checking the pod labels")
			Eventually(func() map[string]string {
				pod := &corev1.Pod{}
				err := k8sClient.Get(ctx, key, pod)
				Expect(err).NotTo(HaveOccurred())
				return pod.Labels
			}, time.Second*10).Should(HaveKey("app"))

			Eventually(func() map[string]string {
				pod := &corev1.Pod{}
				err := k8sClient.Get(ctx, key, pod)
				Expect(err).NotTo(HaveOccurred())
				return pod.Labels
			}, time.Second*10).Should(HaveKeyWithValue("app", "new-label"))

		})
	})

	Context("Update Configmap Labels", func() {

		It("should update the labels of the configmap", func() {

			key := types.NamespacedName{
				Name:      "test-inheritor",
				Namespace: "default",
			}

			spec := labelsv1.InheritorSpec{
				Selectors: []labelsv1.Selector{
					{
						NamespaceSelector: metav1.LabelSelector{
							MatchLabels: map[string]string{
								"kubernetes.io/metadata.name": key.Namespace,
							},
						},
						IncludeLabels: []string{"app"},
					},
				},
			}

			inheritor := &labelsv1.Inheritor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: spec,
			}

			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: key.Namespace,
					Labels: map[string]string{
						"app": "new-label",
					},
				},
			}

			cm := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
					Labels: map[string]string{
						"foo": "bar",
					},
				},
				Data: map[string]string{
					"test": "test",
				},
			}

			By("Creating the namespace")
			Expect(k8sClient.Update(ctx, ns)).Should(Succeed())

			By("Creating the configmap")
			Expect(k8sClient.Create(ctx, cm)).Should(Succeed())
			time.Sleep(5 * time.Second)

			By("Creating the inheritor")
			Expect(k8sClient.Create(ctx, inheritor)).Should(Succeed())
			time.Sleep(5 * time.Second)

			By("Checking the configmap labels")
			Eventually(func() map[string]string {
				cm := &corev1.ConfigMap{}
				err := k8sClient.Get(ctx, key, cm)
				Expect(err).NotTo(HaveOccurred())
				return cm.Labels
			}, time.Second*10).Should(HaveKey("app"))

			Eventually(func() map[string]string {
				cm := &corev1.ConfigMap{}
				err := k8sClient.Get(ctx, key, cm)
				Expect(err).NotTo(HaveOccurred())
				return cm.Labels
			}, time.Second*10).Should(HaveKeyWithValue("app", "new-label"))
		})
	})

	Context("Update pod in other namespace", func() {
		It("Should update pods in other namespace", func() {

			keyOther := types.NamespacedName{
				Name:      "abc",
				Namespace: "other-namespace",
			}

			spec := labelsv1.InheritorSpec{
				Selectors: []labelsv1.Selector{
					{
						NamespaceSelector: metav1.LabelSelector{
							MatchLabels: map[string]string{
								"kubernetes.io/metadata.name": keyOther.Namespace,
							},
						},
						IncludeLabels: []string{"app"},
					},
				},
			}

			key := types.NamespacedName{
				Name:      "test-inheritor",
				Namespace: "default",
			}

			inheritor := &labelsv1.Inheritor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: spec,
			}

			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: keyOther.Namespace,
					Labels: map[string]string{
						"app": "new-label",
					},
				},
			}

			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      keyOther.Name,
					Namespace: keyOther.Namespace,
					Labels: map[string]string{
						"foo": "bar",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test-container",
							Image: "nginx",
						},
					},
				},
			}

			By("Creating the namespace")
			Expect(k8sClient.Create(ctx, ns)).Should(Succeed())

			By("Creating the pod")
			Expect(k8sClient.Create(ctx, pod)).Should(Succeed())
			time.Sleep(5 * time.Second)

			By("Creating the inheritor")
			Expect(k8sClient.Create(ctx, inheritor)).Should(Succeed())
			time.Sleep(5 * time.Second)

			By("Checking the pod labels")
			Eventually(func() map[string]string {
				pod := &corev1.Pod{}
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      keyOther.Name,
					Namespace: keyOther.Namespace,
				}, pod)
				Expect(err).NotTo(HaveOccurred())
				return pod.Labels
			}, time.Second*10).Should(HaveKey("app"))

			Eventually(func() map[string]string {
				pod := &corev1.Pod{}
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      keyOther.Name,
					Namespace: keyOther.Namespace,
				}, pod)
				Expect(err).NotTo(HaveOccurred())
				return pod.Labels
			}, time.Second*10).Should(HaveKeyWithValue("app", "new-label"))
		})
	})

	Context("Overwrite Pod labels", func() {

		It("should overwrite the labels of the pod", func() {

			key := types.NamespacedName{
				Name:      "test-inheritor",
				Namespace: "default",
			}

			spec := labelsv1.InheritorSpec{
				Selectors: []labelsv1.Selector{
					{
						NamespaceSelector: metav1.LabelSelector{
							MatchLabels: map[string]string{
								"kubernetes.io/metadata.name": key.Namespace,
							},
						},
						IncludeLabels: []string{"app", "foo"},
					},
				},
			}

			inheritor := &labelsv1.Inheritor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: spec,
			}

			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: key.Namespace,
					Labels: map[string]string{
						"app": "new-label",
						"foo": "new-foo",
					},
				},
			}

			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
					Labels: map[string]string{
						"app": "old-label",
						"foo": "old-foo",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test-container",
							Image: "nginx",
						},
					},
				},
			}

			By("Creating the namespace")
			Expect(k8sClient.Update(ctx, ns)).Should(Succeed())

			By("Creating the pod")
			Expect(k8sClient.Create(ctx, pod)).Should(Succeed())
			time.Sleep(5 * time.Second)

			By("Creating the inheritor")
			Expect(k8sClient.Create(ctx, inheritor)).Should(Succeed())
			time.Sleep(5 * time.Second)

			By("Checking the pod labels")

			for _, label := range spec.Selectors[0].IncludeLabels {
				Eventually(func() map[string]string {
					pod := &corev1.Pod{}
					err := k8sClient.Get(ctx, key, pod)
					Expect(err).NotTo(HaveOccurred())
					return pod.Labels
				}, time.Second*10).Should(HaveKeyWithValue(label, ns.Labels[label]))
			}
		})
	})

})

// Context("When reconciling a resource", func() {
// 	const resourceName = "test-resource"

// 	ctx := context.Background()

// 	typeNamespacedName := types.NamespacedName{
// 		Name:      resourceName,
// 		Namespace: "default", // TODO(user):Modify as needed
// 	}
// 	inheritor := &labelsv1.Inheritor{}

// 	BeforeEach(func() {
// 		By("creating the custom resource for the Kind Inheritor")
// 		err := k8sClient.Get(ctx, typeNamespacedName, inheritor)
// 		if err != nil && errors.IsNotFound(err) {
// 			resource := &labelsv1.Inheritor{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:      resourceName,
// 					Namespace: "default",
// 				},
// 				// TODO(user): Specify other spec details if needed.
// 			}
// 			Expect(k8sClient.Create(ctx, resource)).To(Succeed())
// 		}
// 	})

// 	AfterEach(func() {
// 		// TODO(user): Cleanup logic after each test, like removing the resource instance.
// 		resource := &labelsv1.Inheritor{}
// 		err := k8sClient.Get(ctx, typeNamespacedName, resource)
// 		Expect(err).NotTo(HaveOccurred())

// 		By("Cleanup the specific resource instance Inheritor")
// 		Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
// 	})
// 	It("should successfully reconcile the resource", func() {
// 		By("Reconciling the created resource")
// 		controllerReconciler := &InheritorReconciler{
// 			Client: k8sClient,
// 			Scheme: k8sClient.Scheme(),
// 		}

// 		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
// 			NamespacedName: typeNamespacedName,
// 		})
// 		Expect(err).NotTo(HaveOccurred())
// 		// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
// 		// Example: If you expect a certain status condition after reconciliation, verify it here.
// 	})
// })
