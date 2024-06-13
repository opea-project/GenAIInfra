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
								"no_proxy":    ".codegen.svc.cluster.local",
								"http_proxy":  "insert-your-http-proxy-here",
								"https_proxy": "insert-your-https-proxy-here",
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
													"no_proxy":    ".codegen.svc.cluster.local",
													"http_proxy":  "insert-your-http-proxy-here",
													"https_proxy": "insert-your-https-proxy-here",
													"endpoint":    "/v1/embeddings",
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

// func TestGetKubeConfig(t *testing.T) {
// 	config, err := getKubeConfig()
// 	if err != nil {
// 		t.Errorf("getKubeConfig() error = %v", err)
// 		return
// 	}
// 	if config == nil {
// 		t.Error("Expected kube config to be not nil")
// 	}
// 	// Add more assertions based on what getKubeConfig() is supposed to do
// }

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

func TestGetCustomConfig_ExpectedCases(t *testing.T) {
	step := "Embedding"
	svcCfg := &map[string]string{
		"no_proxy":     "localhost",
		"http_proxy":   "http://proxy.example.com",
		"https_proxy":  "https://proxy.example.com",
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

	actualCfg, err := getCustomConfig(step, svcCfg, yamlFile)
	if err != nil {
		t.Errorf("getCustomConfig() returned an error: %v", err)
	}

	if strings.TrimSpace(actualCfg) != strings.TrimSpace(expectedCfg) {
		t.Errorf("Expected config:\n%v\n\nBut got:\n%v", expectedCfg, actualCfg)
	}

	step = TeiEmbedding
	actualCfg, err = getCustomConfig(step, svcCfg, yamlFile)
	if err != nil {
		t.Errorf("getCustomConfig() returned an error: %v", err)
	}

	if strings.TrimSpace(actualCfg) != strings.TrimSpace(expectedCfg) {
		t.Errorf("Expected config:\n%v\n\nBut got:\n%v", expectedCfg, actualCfg)
	}

	step = VectorDB
	actualCfg, err = getCustomConfig(step, svcCfg, yamlFile)
	if err != nil {
		t.Errorf("getCustomConfig() returned an error: %v", err)
	}

	if strings.TrimSpace(actualCfg) != strings.TrimSpace(expectedCfg) {
		t.Errorf("Expected config:\n%v\n\nBut got:\n%v", expectedCfg, actualCfg)
	}

	step = Retriever
	actualCfg, err = getCustomConfig(step, svcCfg, yamlFile)
	if err != nil {
		t.Errorf("getCustomConfig() returned an error: %v", err)
	}

	if strings.TrimSpace(actualCfg) != strings.TrimSpace(expectedCfg) {
		t.Errorf("Expected config:\n%v\n\nBut got:\n%v", expectedCfg, actualCfg)
	}

	step = Reranking
	actualCfg, err = getCustomConfig(step, svcCfg, yamlFile)
	if err != nil {
		t.Errorf("getCustomConfig() returned an error: %v", err)
	}

	if strings.TrimSpace(actualCfg) != strings.TrimSpace(expectedCfg) {
		t.Errorf("Expected config:\n%v\n\nBut got:\n%v", expectedCfg, actualCfg)
	}

	step = TeiReranking
	actualCfg, err = getCustomConfig(step, svcCfg, yamlFile)
	if err != nil {
		t.Errorf("getCustomConfig() returned an error: %v", err)
	}

	if strings.TrimSpace(actualCfg) != strings.TrimSpace(expectedCfg) {
		t.Errorf("Expected config:\n%v\n\nBut got:\n%v", expectedCfg, actualCfg)
	}

	step = Tgi
	actualCfg, err = getCustomConfig(step, svcCfg, yamlFile)
	if err != nil {
		t.Errorf("getCustomConfig() returned an error: %v", err)
	}

	if strings.TrimSpace(actualCfg) != strings.TrimSpace(expectedCfg) {
		t.Errorf("Expected config:\n%v\n\nBut got:\n%v", expectedCfg, actualCfg)
	}

	step = Llm
	actualCfg, err = getCustomConfig(step, svcCfg, yamlFile)
	if err != nil {
		t.Errorf("getCustomConfig() returned an error: %v", err)
	}

	if strings.TrimSpace(actualCfg) != strings.TrimSpace(expectedCfg) {
		t.Errorf("Expected config:\n%v\n\nBut got:\n%v", expectedCfg, actualCfg)
	}

	step = Router
	actualCfg, err = getCustomConfig(step, svcCfg, yamlFile)
	if err != nil {
		t.Errorf("getCustomConfig() returned an error: %v", err)
	}

	if strings.TrimSpace(actualCfg) != strings.TrimSpace(expectedCfg) {
		t.Errorf("Expected config:\n%v\n\nBut got:\n%v", expectedCfg, actualCfg)
	}

}
