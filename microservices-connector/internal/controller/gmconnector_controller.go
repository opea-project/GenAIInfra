/*
* Copyright (C) 2024 Intel Corporation
* SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	mcv1alpha3 "github.com/opea-project/GenAIInfra/microservices-connector/api/v1alpha3"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const yaml_dir = "/tmp/microservices/yamls"

// GMConnectorReconciler reconciles a GMConnector object
type GMConnectorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func getKubeConfig() (*rest.Config, error) {
	var config *rest.Config
	var err error

	if _, err = os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/token"); err == nil {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to create in-cluster config: %w", err)
		}
	} else {
		kubeconfigPath := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
		}
	}
	return config, nil
}

func reconcileResource(step string, ns string, svc string, svcCfg *map[string]string, retSvc *corev1.Service) error {

	var tmpltFile string

	fmt.Printf("get step %s config for %s@%s: %v\n", step, svc, ns, svcCfg)

	//TODO add validation to rule out unexpected case like both embedding and retrieving
	if step == "Configmap" {
		tmpltFile = yaml_dir + "/qna_configmap_xeon.yaml"
	} else if step == "ConfigmapGaudi" {
		tmpltFile = yaml_dir + "/qna_configmap_gaudi.yaml"
	} else if step == "Embedding" {
		tmpltFile = yaml_dir + "/embedding.yaml"
	} else if step == "TeiEmbedding" {
		tmpltFile = yaml_dir + "/tei_embedding_service.yaml"
	} else if step == "TeiEmbeddingGaudi" {
		tmpltFile = yaml_dir + "/tei_embedding_gaudi_service.yaml"
	} else if step == "VectorDB" {
		tmpltFile = yaml_dir + "/redis-vector-db.yaml"
	} else if step == "Retriever" {
		tmpltFile = yaml_dir + "/retriever.yaml"
	} else if step == "Reranking" {
		tmpltFile = yaml_dir + "/reranking.yaml"
	} else if step == "TeiReranking" {
		tmpltFile = yaml_dir + "/tei_reranking_service.yaml"
	} else if step == "Tgi" {
		tmpltFile = yaml_dir + "/tgi_service.yaml"
	} else if step == "TgiGaudi" {
		tmpltFile = yaml_dir + "/tgi_gaudi_service.yaml"
	} else if step == "Llm" {
		tmpltFile = yaml_dir + "/llm.yaml"
	} else if step == "router" {
		tmpltFile = yaml_dir + "/gmc-router.yaml"
	} else {
		return errors.New("unexpected target")
	}

	config, err := getKubeConfig()
	if err != nil {
		return err
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %v", err)
	}

	yamlFile, err := os.ReadFile(tmpltFile)
	if err != nil {
		return fmt.Errorf("failed to read YAML file: %v", err)
	}

	var resources []string
	appliedCfg, err := getCustomConfig(step, svcCfg, yamlFile)
	if err != nil {
		return fmt.Errorf("failed to apply user config: %v", err)
	}
	resources = strings.Split(appliedCfg, "---")
	fmt.Printf("The raw yaml file has been splitted into %v yaml files", len(resources))
	if len(resources) >= 2 {
		resources = resources[1:]
	}

	decUnstructured := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

	for _, res := range resources {
		if res == "" {
			continue
		}

		obj := &unstructured.Unstructured{}
		_, gvk, err := decUnstructured.Decode([]byte(res), nil, obj)
		if err != nil {
			return fmt.Errorf("failed to decode YAML: %v", err)
		}

		gvr, _ := meta.UnsafeGuessKindToResource(*gvk)
		for {
			patchBytes, err := json.Marshal(obj)
			if err != nil {
				return fmt.Errorf("failed to marshal config to json: %v", err)
			}

			createdObj, err := dynamicClient.Resource(gvr).Namespace(ns).Patch(context.TODO(), obj.GetName(), types.ApplyPatchType, patchBytes, metav1.PatchOptions{
				FieldManager: "gmc-controller",
				Force:        ptr.To(true),
			})
			if err != nil {
				fmt.Printf("Failed to reconcile resource: %v\n", err)
			} else {
				fmt.Printf("Resource %s/%s created\n", gvk.Kind, createdObj.GetName())
				if retSvc != nil && createdObj.GetKind() == "Service" {
					err = scheme.Scheme.Convert(createdObj, retSvc, nil)
					if err != nil {
						fmt.Printf("Failed to save service: %v\n", err)
					}
				}
				break
			}
			time.Sleep(time.Second * 2)
		}
	}
	return nil
}

func getServiceURL(service *corev1.Service) string {
	switch service.Spec.Type {
	case corev1.ServiceTypeClusterIP:
		// For ClusterIP, return the cluster IP and port
		if len(service.Spec.Ports) > 0 {
			return fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", service.Name, service.Namespace, service.Spec.Ports[0].Port)
		}
	case corev1.ServiceTypeNodePort:
		// For NodePort, return the node IP and node port. You need to replace <node-ip> with the actual node IP.
		if len(service.Spec.Ports) > 0 {
			return fmt.Sprintf("<node-ip>:%d", service.Spec.Ports[0].NodePort)
		}
	case corev1.ServiceTypeLoadBalancer:
		// For LoadBalancer, return the load balancer IP and port
		if len(service.Spec.Ports) > 0 && len(service.Status.LoadBalancer.Ingress) > 0 {
			return fmt.Sprintf("%s:%d", service.Status.LoadBalancer.Ingress[0].IP, service.Spec.Ports[0].Port)
		}
	case corev1.ServiceTypeExternalName:
		// For ExternalName, return the external name
		return service.Spec.ExternalName
	}
	return ""
}

func getCustomConfig(step string, svcCfg *map[string]string, yamlFile []byte) (string, error) {
	var userDefinedCfg interface{}
	if step == "Configmap" || step == "ConfigmapGaudi" {
		return string(yamlFile), nil
	} else if step == "Embedding" {
		userDefinedCfg = EmbeddingCfg{
			NoProxy:    (*svcCfg)["no_proxy"],
			HttpProxy:  (*svcCfg)["http_proxy"],
			HttpsProxy: (*svcCfg)["https_proxy"],
		}
	} else if step == "TeiEmbedding" {
		userDefinedCfg = TeiEmbeddingCfg{
			NoProxy:    (*svcCfg)["no_proxy"],
			HttpProxy:  (*svcCfg)["http_proxy"],
			HttpsProxy: (*svcCfg)["https_proxy"],
		}
	} else if step == "VectorDB" {
		userDefinedCfg = nil
	} else if step == "Retriever" {
		userDefinedCfg = RetriverCfg{
			NoProxy:    (*svcCfg)["no_proxy"],
			HttpProxy:  (*svcCfg)["http_proxy"],
			HttpsProxy: (*svcCfg)["https_proxy"],
		}
	} else if step == "Reranking" {
		userDefinedCfg = RerankingCfg{
			NoProxy:    (*svcCfg)["no_proxy"],
			HttpProxy:  (*svcCfg)["http_proxy"],
			HttpsProxy: (*svcCfg)["https_proxy"],
		}
	} else if step == "TeiReranking" {
		userDefinedCfg = TeiRerankingCfg{
			NoProxy:    (*svcCfg)["no_proxy"],
			HttpProxy:  (*svcCfg)["http_proxy"],
			HttpsProxy: (*svcCfg)["https_proxy"],
		}
	} else if step == "Tgi" {
		userDefinedCfg = TgiCfg{
			NoProxy:    (*svcCfg)["no_proxy"],
			HttpProxy:  (*svcCfg)["http_proxy"],
			HttpsProxy: (*svcCfg)["https_proxy"],
		}
	} else if step == "Llm" {
		userDefinedCfg = LlmCfg{
			NoProxy:    (*svcCfg)["no_proxy"],
			HttpProxy:  (*svcCfg)["http_proxy"],
			HttpsProxy: (*svcCfg)["https_proxy"],
		}
	} else if step == "router" {
		userDefinedCfg = RouterCfg{
			NoProxy:    (*svcCfg)["no_proxy"],
			HttpProxy:  (*svcCfg)["http_proxy"],
			HttpsProxy: (*svcCfg)["https_proxy"],
			GRAPH_JSON: (*svcCfg)["nodes"]}
	} else {
		userDefinedCfg = nil
	}

	fmt.Printf("user config %v\n", userDefinedCfg)

	tmpl, err := template.New("yamlTemplate").Parse(string(yamlFile))
	if err != nil {
		return string(yamlFile), fmt.Errorf("error parsing template: %v", err)
	}
	if userDefinedCfg != nil {
		var appliedCfg bytes.Buffer
		err = tmpl.Execute(&appliedCfg, userDefinedCfg)
		if err != nil {
			return string(yamlFile), fmt.Errorf("error executing template: %v", err)
		} else {
			fmt.Printf("applied config %s\n", appliedCfg.String())

			return appliedCfg.String(), nil
		}
	} else {
		return string(yamlFile), nil
	}

}

type EmbeddingCfg struct {
	NoProxy    string
	HttpProxy  string
	HttpsProxy string
}
type TeiEmbeddingCfg struct {
	EmbeddingModelId string
	NoProxy          string
	HttpProxy        string
	HttpsProxy       string
}

type RetriverCfg struct {
	NoProxy    string
	HttpProxy  string
	HttpsProxy string
}

type RerankingCfg struct {
	NoProxy    string
	HttpProxy  string
	HttpsProxy string
}

type TeiRerankingCfg struct {
	RerankingModelId string
	NoProxy          string
	HttpProxy        string
	HttpsProxy       string
}

type TgiCfg struct {
	NoProxy    string
	HttpProxy  string
	HttpsProxy string
}

type LlmCfg struct {
	NoProxy    string
	HttpProxy  string
	HttpsProxy string
}

type RouterCfg struct {
	NoProxy    string
	HttpProxy  string
	HttpsProxy string
	GRAPH_JSON string
}

//+kubebuilder:rbac:groups=gmc.opea.io,resources=gmconnectors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=gmc.opea.io,resources=gmconnectors/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=gmc.opea.io,resources=gmconnectors/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// the GMConnector object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *GMConnectorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	graph := &mcv1alpha3.GMConnector{}
	if err := r.Get(ctx, req.NamespacedName, graph); err != nil {
		if apierr.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}
	// get the router config
	// r.Log.Info("Reconciling connector graph", "apiVersion", graph.APIVersion, "graph", graph.Name)
	fmt.Println("Reconciling connector graph", "apiVersion", graph.APIVersion, "graph", graph.Name)
	err := reconcileResource("Configmap", req.NamespacedName.Namespace, "", nil, nil)
	if err != nil {
		return reconcile.Result{Requeue: true}, errors.Wrapf(err, "Failed to reconcile the Configmap file")
	}
	for node, router := range graph.Spec.Nodes {
		for i, step := range router.Steps {
			fmt.Println("reconcile resource for node:", step.StepName)

			if step.Executor.ExternalService == "" {
				ns := step.Executor.InternalService.NameSpace
				svcName := step.Executor.InternalService.ServiceName
				fmt.Println("trying to reconcile internal service [", svcName, "] in namespace ", ns)

				service := &corev1.Service{}
				err := reconcileResource(step.StepName, ns, svcName, &step.Executor.InternalService.Config, service)
				if err != nil {
					return reconcile.Result{Requeue: true}, errors.Wrapf(err, "Failed to reconcile service for %s", svcName)
				}

				graph.Spec.Nodes[node].Steps[i].ServiceURL = getServiceURL(service) + step.Executor.InternalService.Config["endpoint"]
				fmt.Printf("the service URL is: %s\n", graph.Spec.Nodes[node].Steps[i].ServiceURL)
			} else {
				fmt.Println("external service is found", "name", step.ExternalService)
				graph.Spec.Nodes[node].Steps[i].ServiceURL = step.ExternalService
			}
		}
		fmt.Println()
	}

	//to start a router controller
	routerService := &corev1.Service{}
	err = r.Client.Get(ctx, types.NamespacedName{Namespace: graph.Spec.RouterConfig.NameSpace, Name: graph.Spec.RouterConfig.ServiceName}, routerService)
	if err == nil {
		fmt.Println("success to get router service ", graph.Spec.RouterConfig.ServiceName)
	} else {
		jsonBytes, err := json.Marshal(graph)
		if err != nil {
			// handle error
			return reconcile.Result{Requeue: true}, errors.Wrapf(err, "Failed to Marshal routes for %s", graph.Spec.RouterConfig.Name)
		}
		jsonString := string(jsonBytes)
		if graph.Spec.RouterConfig.Config == nil {
			graph.Spec.RouterConfig.Config = make(map[string]string)
		}
		graph.Spec.RouterConfig.Config["nodes"] = "'" + jsonString + "'"
		err = reconcileResource(graph.Spec.RouterConfig.Name, graph.Spec.RouterConfig.NameSpace, graph.Spec.RouterConfig.ServiceName, &graph.Spec.RouterConfig.Config, nil)
		if err != nil {
			return reconcile.Result{Requeue: true}, errors.Wrapf(err, "Failed to reconcile router service")
		}
	}
	graph.Status.AccessURL = getServiceURL(routerService)
	graph.Status.Status = "Success"
	if err = r.Status().Update(context.TODO(), graph); err != nil {
		return reconcile.Result{Requeue: true}, errors.Wrapf(err, "Failed to Update CR status to %s", graph.Status.Status)
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GMConnectorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mcv1alpha3.GMConnector{}).
		Complete(r)
}
