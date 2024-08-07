/*
* Copyright (C) 2024 Intel Corporation
* SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	mcv1alpha3 "github.com/opea-project/GenAIInfra/microservices-connector/api/v1alpha3"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment

func TestControllers(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	err := os.MkdirAll(yaml_dir, os.ModePerm)
	Expect(err).NotTo(HaveOccurred())

	// templateDir := "../../../manifests/ChatQnA"

	// files := []string{
	// 	templateDir + tei_reranking_service_yaml,
	// 	templateDir + embedding_yaml,
	// 	templateDir + tei_embedding_service_yaml,
	// 	templateDir + tei_embedding_gaudi_service_yaml,
	// 	templateDir + tgi_service_yaml,
	// 	templateDir + tei_reranking_service_yaml,
	// 	templateDir + tgi_gaudi_service_yaml,
	// 	templateDir + llm_yaml,
	// 	templateDir + redis_vector_db_yaml,
	// 	templateDir + retriever_yaml,
	// 	templateDir + reranking_yaml,
	// 	templateDir + "/qna_configmap_xeon.yaml",
	// 	templateDir + "/qna_configmap_gaudi.yaml",
	// 	"../../config/gmcrouter/gmc-router.yaml",
	// }

	templateDir := "../../config/manifests"

	files := []string{
		templateDir + "/tei.yaml",
		templateDir + "/tei_gaudi.yaml",
		templateDir + "/embedding-usvc.yaml",
		templateDir + "/redis-vector-db.yaml",
		templateDir + "/retriever-usvc.yaml",
		templateDir + "/reranking-usvc.yaml",
		templateDir + "/teirerank.yaml",
		templateDir + "/tgi.yaml",
		templateDir + "/tgi_gaudi.yaml",
		templateDir + "/llm-uservice.yaml",
		templateDir + "/docsum-llm-uservice.yaml",
		"../../config/gmcrouter/gmc-router.yaml",
	}
	for _, file := range files {
		cmd := exec.Command("cp", file, yaml_dir)
		// cmd := exec.Command("ls", file)
		err = cmd.Run()
		Expect(err).NotTo(HaveOccurred())
	}

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,

		// The BinaryAssetsDirectory is only required if you want to run the tests directly
		// without call the makefile target test. If not informed it will look for the
		// default path defined in controller-runtime which is /usr/local/kubebuilder/.
		// Note that you must have the required binaries setup under the bin directory to perform
		// the tests directly. When we run make test it will be setup and used automatically.
		BinaryAssetsDirectory: filepath.Join("..", "..", "bin", "k8s",
			fmt.Sprintf("1.29.0-%s-%s", runtime.GOOS, runtime.GOARCH)),
	}

	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = mcv1alpha3.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
