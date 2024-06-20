/*
* Copyright (C) 2024 Intel Corporation
* SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"context"
	"strings"
	"testing"

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
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
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
										StepName: "Embedding",
										Data:     "$response",
										Executor: mcv1alpha3.Executor{
											InternalService: mcv1alpha3.GMCTarget{
												NameSpace:   "default",
												ServiceName: "embedding-service",
												Config: map[string]string{
													"endpoint": "/v1/embeddings",
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
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
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
	actualURL := GetServiceURL(service)

	if actualURL != expectedURL {
		t.Errorf("Expected URL: %s, but got: %s", expectedURL, actualURL)
	}
}

func TestApplyCustomConfig_ExpectedCases(t *testing.T) {
	svcCfg := &map[string]string{
		"tei_endpoint": "http://tei.example.com",
	}
	yamlFile := []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  key1: value1
  key2: value2
`)

	expectedCfg := `apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  key1: value1
  key2: value2
`

	step := Router
	actualCfg, err := patchCustomConfigToTemplates(step, svcCfg, yamlFile)
	if err != nil {
		t.Errorf("applyCustomConfig() returned an error: %v", err)
	}

	if strings.TrimSpace(actualCfg) != strings.TrimSpace(expectedCfg) {
		t.Errorf("Expected config:\n%v\n\nBut got:\n%v", expectedCfg, actualCfg)
	}

}

func TestApplyCustomConfig_EmptyStep(t *testing.T) {
	step := ""
	svcCfg := &map[string]string{
		"tei_endpoint": "http://tei.example.com",
	}
	yamlFile := []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  key1: value1
  key2: value2
`)
	expectedCfg := string(yamlFile)
	actualCfg, err := patchCustomConfigToTemplates(step, svcCfg, yamlFile)
	if err != nil {
		t.Errorf("applyCustomConfig() returned an error: %v", err)
	}
	if strings.TrimSpace(actualCfg) != strings.TrimSpace(expectedCfg) {
		t.Errorf("Expected config:\n%v\n\nBut got:\n%v", expectedCfg, actualCfg)
	}
}
