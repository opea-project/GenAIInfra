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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1 "k8s.io/api/apps/v1"
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

			mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
				Scheme: k8sClient.Scheme(),
				// Client: k8sClient,
			})
			Expect(err).NotTo(HaveOccurred())

			// Create a new GMConnectorReconciler
			reconciler := &GMConnectorReconciler{
				Client: k8sClient,
				// Log:    ctrl.Log.WithName("controllers").WithName("GMConnector"),
				Scheme: mgr.GetScheme(),
			}

			// Call the SetupWithManager function
			err = reconciler.SetupWithManager(mgr)
			Expect(err).NotTo(HaveOccurred())

			err = k8sClient.Get(ctx, typeNamespacedName, gmconnector)
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
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance
			resource := &mcv1alpha3.GMConnector{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance GMConnector")
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
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "embedding-service",
				Namespace: "default",
			}, &corev1.Service{})).To(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "embedding-service-deployment",
				Namespace: "default",
			}, &appsv1.Deployment{})).To(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "embedding-usvc-config",
				Namespace: "default",
			}, &corev1.ConfigMap{})).To(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "tei-embedding-service",
				Namespace: "default",
			}, &corev1.Service{})).To(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "tei-embedding-service-deployment",
				Namespace: "default",
			}, &appsv1.Deployment{})).To(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "tei-config",
				Namespace: "default",
			}, &corev1.ConfigMap{})).To(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "vector-service",
				Namespace: "default",
			}, &corev1.Service{})).To(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "vector-service-deployment",
				Namespace: "default",
			}, &appsv1.Deployment{})).To(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "retriever-service",
				Namespace: "default",
			}, &corev1.Service{})).To(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "retriever-service-deployment",
				Namespace: "default",
			}, &appsv1.Deployment{})).To(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "retriever-usvc-config",
				Namespace: "default",
			}, &corev1.ConfigMap{})).To(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "rerank-service",
				Namespace: "default",
			}, &corev1.Service{})).To(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "rerank-service-deployment",
				Namespace: "default",
			}, &appsv1.Deployment{})).To(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "reranking-usvc-config",
				Namespace: "default",
			}, &corev1.ConfigMap{})).To(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "tei-reranking-svc",
				Namespace: "default",
			}, &corev1.Service{})).To(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "tei-reranking-svc-deployment",
				Namespace: "default",
			}, &appsv1.Deployment{})).To(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "teirerank-config",
				Namespace: "default",
			}, &corev1.ConfigMap{})).To(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "tgi-service-name",
				Namespace: "default",
			}, &corev1.Service{})).To(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "tgi-service-name-deployment",
				Namespace: "default",
			}, &appsv1.Deployment{})).To(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "tgi-config",
				Namespace: "default",
			}, &corev1.ConfigMap{})).To(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "llm-service",
				Namespace: "default",
			}, &corev1.Service{})).To(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "llm-service-deployment",
				Namespace: "default",
			}, &appsv1.Deployment{})).To(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "llm-uservice-config",
				Namespace: "default",
			}, &corev1.ConfigMap{})).To(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "router-service",
				Namespace: "default",
			}, &corev1.Service{})).To(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "router-service-deployment",
				Namespace: "default",
			}, &appsv1.Deployment{})).To(Succeed())

			pipeline := &mcv1alpha3.GMConnector{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, pipeline)).To(Succeed())
			Expect(pipeline.Status.Status).To(Equal("0/0/9"))
			Expect(len(pipeline.Status.Annotations)).To(Equal(25))

		})

		It("should successfully reconcile the deployment for status update", func() {
			controllerReconciler := &GMConnectorReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}
			By("Reconciling the existed resource")
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			embedDp := &appsv1.Deployment{}
			embedDpMeta := types.NamespacedName{
				Name:      "embedding-service-deployment",
				Namespace: "default",
			}
			Expect(k8sClient.Get(ctx, embedDpMeta, embedDp)).To(Succeed())
			embedDp.Status.AvailableReplicas = int32(1)
			embedDp.Status.Replicas = embedDp.Status.AvailableReplicas
			embedDp.Status.ReadyReplicas = embedDp.Status.AvailableReplicas
			Expect(*embedDp.Spec.Replicas).To(Equal(int32(1)))
			Expect(embedDp.OwnerReferences[0].Name).To(Equal(resourceName))
			Expect(embedDp.OwnerReferences[0].Kind).To(Equal("GMConnector"))
			err = k8sClient.Status().Update(ctx, embedDp)
			Expect(err).NotTo(HaveOccurred())
			embedDp2 := &appsv1.Deployment{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "embedding-service-deployment",
				Namespace: "default",
			}, embedDp2)).To(Succeed())
			Expect(embedDp2.Status.AvailableReplicas).To(Equal(int32(1)))
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: embedDpMeta,
			})
			Expect(err).NotTo(HaveOccurred())
			pipeline := &mcv1alpha3.GMConnector{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, pipeline)).To(Succeed())
			Expect(pipeline.Status.Status).To(Equal("1/0/9"))
		})

		It("should successfully reconcile the deployment for removing step", func() {
			controllerReconciler := &GMConnectorReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}
			By("Reconciling the existed resource")
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			err = k8sClient.Get(ctx, typeNamespacedName, gmconnector)
			Expect(err).NotTo(HaveOccurred())

			resource := &mcv1alpha3.GMConnector{
				TypeMeta:   gmconnector.TypeMeta,
				ObjectMeta: gmconnector.ObjectMeta,
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
							},
						},
					},
				},
			}
			Expect(k8sClient.Update(ctx, resource)).To(Succeed())
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "embedding-service",
				Namespace: "default",
			}, &corev1.Service{})).NotTo(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "embedding-service-deployment",
				Namespace: "default",
			}, &appsv1.Deployment{})).NotTo(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "embedding-usvc-config",
				Namespace: "default",
			}, &corev1.ConfigMap{})).NotTo(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "tei-embedding-service",
				Namespace: "default",
			}, &corev1.Service{})).To(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "tei-embedding-service-deployment",
				Namespace: "default",
			}, &appsv1.Deployment{})).To(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "tei-config",
				Namespace: "default",
			}, &corev1.ConfigMap{})).To(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "vector-service",
				Namespace: "default",
			}, &corev1.Service{})).To(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "vector-service-deployment",
				Namespace: "default",
			}, &appsv1.Deployment{})).To(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "retriever-service",
				Namespace: "default",
			}, &corev1.Service{})).NotTo(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "retriever-service-deployment",
				Namespace: "default",
			}, &appsv1.Deployment{})).NotTo(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "retriever-usvc-config",
				Namespace: "default",
			}, &corev1.ConfigMap{})).NotTo(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "rerank-service",
				Namespace: "default",
			}, &corev1.Service{})).NotTo(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "rerank-service-deployment",
				Namespace: "default",
			}, &appsv1.Deployment{})).NotTo(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "reranking-usvc-config",
				Namespace: "default",
			}, &corev1.ConfigMap{})).NotTo(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "tei-reranking-svc",
				Namespace: "default",
			}, &corev1.Service{})).To(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "tei-reranking-svc-deployment",
				Namespace: "default",
			}, &appsv1.Deployment{})).To(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "teirerank-config",
				Namespace: "default",
			}, &corev1.ConfigMap{})).To(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "tgi-service-name",
				Namespace: "default",
			}, &corev1.Service{})).To(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "tgi-service-name-deployment",
				Namespace: "default",
			}, &appsv1.Deployment{})).To(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "tgi-config",
				Namespace: "default",
			}, &corev1.ConfigMap{})).To(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "llm-service",
				Namespace: "default",
			}, &corev1.Service{})).NotTo(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "llm-service-deployment",
				Namespace: "default",
			}, &appsv1.Deployment{})).NotTo(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "llm-uservice-config",
				Namespace: "default",
			}, &corev1.ConfigMap{})).NotTo(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "router-service",
				Namespace: "default",
			}, &corev1.Service{})).To(Succeed())

			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "router-service-deployment",
				Namespace: "default",
			}, &appsv1.Deployment{})).To(Succeed())

			pipeline := &mcv1alpha3.GMConnector{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, pipeline)).To(Succeed())
			Expect(pipeline.Status.Status).To(Equal("0/0/5"))
			Expect(len(pipeline.Status.Annotations)).To(Equal(13))
		})
	})
})

var _ = Describe("Predicate Functions", func() {
	var (
		oldGMConnector *mcv1alpha3.GMConnector
		newGMConnector *mcv1alpha3.GMConnector
		oldDeployment  *appsv1.Deployment
		newDeployment  *appsv1.Deployment
		updateEvent    event.UpdateEvent
	)

	BeforeEach(func() {
		oldGMConnector = &mcv1alpha3.GMConnector{
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
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-resource2",
				Namespace: "default",
				UID:       "1f9a258c-b7d2-4bb3-9fac-ddf1b4369d25",
			},
		}
		newGMConnector = oldGMConnector.DeepCopy()
		oldDeployment = &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "embedding-service-deployment",
				Namespace: "default",
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: "gmc.opea.io/v1alpha3",
						Kind:       "GMConnector",
						Name:       "test-resource2",
						UID:        "1f9a258c-b7d2-4bb3-9fac-ddf1b4369d25",
					},
				},
			},
			Status: appsv1.DeploymentStatus{
				AvailableReplicas: 1,
				Conditions: []appsv1.DeploymentCondition{
					{
						Type:   appsv1.DeploymentAvailable,
						Status: corev1.ConditionTrue,
					},
				},
				ReadyReplicas:   1,
				Replicas:        1,
				UpdatedReplicas: 1,
			},
		}
		newDeployment = oldDeployment.DeepCopy()
	})

	Describe("isGMConnectorSpecOrMetadataChanged", func() {
		It("should return true if the spec has changed", func() {
			newGMConnector.Spec.RouterConfig.Name = "newRouter"
			updateEvent = event.UpdateEvent{
				ObjectOld: oldGMConnector,
				ObjectNew: newGMConnector,
			}
			Expect(isGMCSpecOrMetadataChanged(updateEvent)).To(BeTrue())
		})

		It("should return true if the metadata has changed", func() {
			newGMConnector.ObjectMeta.Name = "new-name"
			updateEvent = event.UpdateEvent{
				ObjectOld: oldGMConnector,
				ObjectNew: newGMConnector,
			}
			Expect(isGMCSpecOrMetadataChanged(updateEvent)).To(BeTrue())
		})
		It("should return false if neither spec nor metadata has changed", func() {
			updateEvent = event.UpdateEvent{
				ObjectOld: oldGMConnector,
				ObjectNew: newGMConnector,
			}
			Expect(isGMCSpecOrMetadataChanged(updateEvent)).To(BeFalse())
		})
	})

	Describe("isDeploymentStatusChanged", func() {
		It("should return true if the status has changed", func() {
			newDeployment.Status.Conditions[0].Status = corev1.ConditionFalse
			updateEvent = event.UpdateEvent{
				ObjectOld: oldDeployment,
				ObjectNew: newDeployment,
			}
			Expect(isDeploymentStatusChanged(updateEvent)).To(BeTrue())
		})

		It("should return false if the status has not changed", func() {
			updateEvent = event.UpdateEvent{
				ObjectOld: oldDeployment,
				ObjectNew: newDeployment,
			}
			Expect(isDeploymentStatusChanged(updateEvent)).To(BeFalse())
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
func TestIsMetadataChanged(t *testing.T) {
	oldObject := &metav1.ObjectMeta{
		Name:      "fido",
		Namespace: "sprint",
		Labels: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
		Generation: 1,
	}

	newObject := &metav1.ObjectMeta{
		Name:      "dido",
		Namespace: "sprint",
		Labels: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
		Generation: 1,
	}

	changed := isMetadataChanged(nil, newObject)
	if !changed {
		t.Errorf("Expected metadata changes to be detected, but got false")
	}

	changed = isMetadataChanged(oldObject, nil)
	if !changed {
		t.Errorf("Expected metadata changes to be detected, but got false")
	}

	//check name
	changed = isMetadataChanged(oldObject, newObject)
	if !changed {
		t.Errorf("Expected metadata changes to be detected, but got false")
	}

	//check name space
	newObject = &metav1.ObjectMeta{
		Name:      "fido",
		Namespace: "coca-cola",
		Labels: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}
	changed = isMetadataChanged(oldObject, newObject)
	if !changed {
		t.Errorf("Expected metadata changes to not be detected, but got true")
	}

	// check label
	newObject = &metav1.ObjectMeta{
		Name:      "fido",
		Namespace: "sprint",
		Labels: map[string]string{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		},
		DeletionTimestamp: &metav1.Time{
			Time: time.Now(),
		},
	}
	changed = isMetadataChanged(oldObject, newObject)
	if !changed {
		t.Errorf("Expected metadata changes to not be detected, but got true")
	}

	newObject.Labels = map[string]string{
		"key1": "value1",
		"key2": "value4",
	}
	changed = isMetadataChanged(oldObject, newObject)
	if !changed {
		t.Errorf("Expected metadata changes to not be detected, but got true")
	}

	newObject.Labels = map[string]string{
		"key1": "value1",
	}
	changed = isMetadataChanged(oldObject, newObject)
	if !changed {
		t.Errorf("Expected metadata changes to not be detected, but got true")
	}

	// check deletion timestamp
	newObject = &metav1.ObjectMeta{
		Name:      "fido",
		Namespace: "sprint",
		Labels: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
		DeletionTimestamp: &metav1.Time{
			Time: time.Now(),
		},
	}
	changed = isMetadataChanged(oldObject, newObject)
	if !changed {
		t.Errorf("Expected metadata changes to not be detected, but got true")
	}

	// check annotation
	newObject = &metav1.ObjectMeta{
		Name:      "fido",
		Namespace: "sprint",
		Labels: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
		Annotations: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}
	changed = isMetadataChanged(oldObject, newObject)
	if changed {
		t.Errorf("Expected metadata changes to not be detected, but got true")
	}

	// check generation
	newObject = &metav1.ObjectMeta{
		Name:      "fido",
		Namespace: "sprint",
		Labels: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
		Generation: 2,
	}
	changed = isMetadataChanged(oldObject, newObject)
	if changed {
		t.Errorf("Expected metadata changes to not be detected, but got true")
	}
}
