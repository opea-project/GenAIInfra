/*
* Copyright (C) 2024 Intel Corporation
* SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mcv1alpha3 "github.com/opea-project/GenAIInfra/microservices-connector/api/v1alpha3"
)

var _ = Describe("GMConnector Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		gmconnector := &mcv1alpha3.GMConnector{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind GMConnector")
			err := k8sClient.Get(ctx, typeNamespacedName, gmconnector)
			if err != nil && errors.IsNotFound(err) {
				resource := &mcv1alpha3.GMConnector{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "gmc.opea.io/v1alpha3",
						Kind:       "GMConnector",
					},

					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
						UID:       "1f9a258c-b7d2-4bb3-9fac-ddf1b4369d24",
					},
					Spec: mcv1alpha3.GMConnectorSpec{
						RouterConfig: mcv1alpha3.RouterConfig{
							Name:        "router",
							ServiceName: "router-service",
							Config: map[string]string{
								"endpoint": "/",
							},
						},
						Nodes: map[string]mcv1alpha3.Router{
							"root": {
								RouterType: "Sequence",
								Steps: []mcv1alpha3.Step{
									{
										StepName: Embedding,
										Data:     "$response",
										Executor: mcv1alpha3.Executor{
											InternalService: mcv1alpha3.GMCTarget{
												NameSpace:   "default",
												ServiceName: "embedding-service",
												Config: map[string]string{
													"endpoint":               "/v1/embeddings",
													"TEI_EMBEDDING_ENDPOINT": "tei-embedding-service",
												},
											},
										},
									},
									{
										StepName: TeiEmbedding,
										Executor: mcv1alpha3.Executor{
											InternalService: mcv1alpha3.GMCTarget{
												NameSpace:   "default",
												ServiceName: "tei-embedding-service",
												Config: map[string]string{
													"endpoint": "/v1/tei-embeddings",
													"MODEL_ID": "somemodel",
												},
												IsDownstreamService: true,
											},
										},
									},
									{
										StepName: VectorDB,
										Executor: mcv1alpha3.Executor{
											InternalService: mcv1alpha3.GMCTarget{
												NameSpace:   "default",
												ServiceName: "vector-service",
												Config: map[string]string{
													"endpoint": "/v1/vec",
												},
												IsDownstreamService: true,
											},
										},
									},
									{
										StepName: Retriever,
										Executor: mcv1alpha3.Executor{
											InternalService: mcv1alpha3.GMCTarget{
												NameSpace:   "default",
												ServiceName: "retriever-service",
												Config: map[string]string{
													"endpoint":               "/v1/retrv",
													"REDIS_URL":              "vector-service",
													"TEI_EMBEDDING_ENDPOINT": "tei-embedding-service",
												},
											},
										},
									},
									{
										StepName: Reranking,
										Executor: mcv1alpha3.Executor{
											InternalService: mcv1alpha3.GMCTarget{
												NameSpace:   "default",
												ServiceName: "rerank-service",
												Config: map[string]string{
													"endpoint":               "/v1/reranking",
													"TEI_RERANKING_ENDPOINT": "tei-reranking-svc",
												},
											},
										},
									},
									{
										StepName: TeiReranking,
										Executor: mcv1alpha3.Executor{
											InternalService: mcv1alpha3.GMCTarget{
												NameSpace:   "default",
												ServiceName: "tei-reranking-svc",
												Config: map[string]string{
													"endpoint": "/rernk",
												},
												IsDownstreamService: true,
											},
										},
									},
									{
										StepName: Tgi,
										Executor: mcv1alpha3.Executor{
											InternalService: mcv1alpha3.GMCTarget{
												NameSpace:   "default",
												ServiceName: "tgi-service-name",
												Config: map[string]string{
													"endpoint": "/generate",
												},
												IsDownstreamService: true,
											},
										},
									},
									{
										StepName: Llm,
										Executor: mcv1alpha3.Executor{
											InternalService: mcv1alpha3.GMCTarget{
												NameSpace:   "default",
												ServiceName: "llm-service",
												Config: map[string]string{
													"endpoint":         "/v1/llm",
													"TGI_LLM_ENDPOINT": "tgi-service-name",
												},
											},
										},
									},
								},
							},
						},
					},
				}
				// Expect(k8sClient.Create(ctx, resource)).To(Succeed())
				Eventually(func() error {
					return k8sClient.Create(ctx, resource)
				}, 60*time.Second, 1*time.Second).Should(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			Eventually(func() error {
				resource := &mcv1alpha3.GMConnector{}
				return k8sClient.Get(ctx, typeNamespacedName, resource)
			}, 60*time.Second, 1*time.Second).Should(Succeed())

			By("Cleanup the specific resource instance GMConnector")
			resource := &mcv1alpha3.GMConnector{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &GMConnectorReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})

			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
		})
	})
})

func TestGetServiceURL(t *testing.T) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "default",
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Port: 8080,
				},
			},
		},
	}

	expectedURL := "http://test-service.default.svc.cluster.local:8080"
	actualURL := getServiceURL(service)

	if actualURL != expectedURL {
		t.Errorf("Expected URL: %s, but got: %s", expectedURL, actualURL)
	}
}
