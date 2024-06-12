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
	yaml2 "gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
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

const (
	Configmap                        = "Configmap"
	ConfigmapGaudi                   = "ConfigmapGaudi"
	Embedding                        = "Embedding"
	TeiEmbedding                     = "TeiEmbedding"
	TeiEmbeddingGaudi                = "TeiEmbeddingGaudi"
	VectorDB                         = "VectorDB"
	Retriever                        = "Retriever"
	Reranking                        = "Reranking"
	TeiReranking                     = "TeiReranking"
	Tgi                              = "Tgi"
	TgiGaudi                         = "TgiGaudi"
	Llm                              = "Llm"
	Router                           = "router"
	xeon                             = "xeon"
	gaudi                            = "gaudi"
	tei_reranking_service_yaml       = "/tei_reranking_service.yaml"
	embedding_yaml                   = "/embedding.yaml"
	tei_embedding_service_yaml       = "/tei_embedding_service.yaml"
	tei_embedding_gaudi_service_yaml = "/tei_embedding_gaudi_service.yaml"
	tgi_service_yaml                 = "/tgi_service.yaml"
	tgi_gaudi_service_yaml           = "/tgi_gaudi_service.yaml"
	llm_yaml                         = "/llm.yaml"
	gmc_router_yaml                  = "/gmc-router.yaml"
	redis_vector_db_yaml             = "/redis-vector-db.yaml"
	retriever_yaml                   = "/retriever.yaml"
	reranking_yaml                   = "/reranking.yaml"
	yaml_dir                         = "/tmp/microservices/yamls"
)

// GMConnectorReconciler reconciles a GMConnector object
type GMConnectorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

type RouterCfg struct {
	NoProxy    string
	HttpProxy  string
	HttpsProxy string
	GRAPH_JSON string
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

func getStepYamlTemplate(step string) string {
	var tmpltFile string
	//TODO add validation to rule out unexpected case like both embedding and retrieving
	if step == Embedding {
		tmpltFile = yaml_dir + embedding_yaml
	} else if step == TeiEmbedding {
		tmpltFile = yaml_dir + tei_embedding_service_yaml
	} else if step == TeiEmbeddingGaudi {
		tmpltFile = yaml_dir + tei_embedding_gaudi_service_yaml
	} else if step == VectorDB {
		tmpltFile = yaml_dir + redis_vector_db_yaml
	} else if step == Retriever {
		tmpltFile = yaml_dir + retriever_yaml
	} else if step == Reranking {
		tmpltFile = yaml_dir + reranking_yaml
	} else if step == TeiReranking {
		tmpltFile = yaml_dir + tei_reranking_service_yaml
	} else if step == Tgi {
		tmpltFile = yaml_dir + tgi_service_yaml
	} else if step == TgiGaudi {
		tmpltFile = yaml_dir + tgi_gaudi_service_yaml
	} else if step == Llm {
		tmpltFile = yaml_dir + llm_yaml
	} else if step == Router {
		tmpltFile = yaml_dir + gmc_router_yaml
	} else {
		return ""
	}
	return tmpltFile
}

func reconcileResource(ctx context.Context, dynamicClient *dynamic.DynamicClient, step string, ns string, svc string, svcCfg *map[string]string, retSvc *corev1.Service) error {
	fmt.Printf("get step %s config for %s@%s: %v\n", step, svc, ns, svcCfg)
	tmpltFile := getStepYamlTemplate(step)
	if tmpltFile == "" {
		return errors.New("unexpected target")
	}
	yamlFile, err := os.ReadFile(tmpltFile)
	if err != nil {
		return fmt.Errorf("failed to read YAML file: %v", err)
	}

	var resources []string
	appliedCfg, err := applyCustomConfig(step, svcCfg, yamlFile)
	if err != nil {
		return fmt.Errorf("failed to apply user config: %v", err)
	}
	resources = strings.Split(appliedCfg, "---")
	fmt.Printf("The raw yaml file has been split into %v yaml files\n", len(resources))

	for _, res := range resources {
		if res == "" || !strings.Contains(res, "kind:") {
			continue
		}
		// if create failed, wait 2s and retry, this will be removed when the monitor task is implemented
		for {
			createdObj, err := applyResourceToK8s(ctx, dynamicClient, ns, []byte(res))
			if err != nil {
				fmt.Printf("Failed to reconcile resource: %v\n", err)
			} else {
				fmt.Printf("Success to reconcile %s: %s\n", createdObj.GetKind(), createdObj.GetName())

				// return the service obj to get the serivce URL from it
				if retSvc != nil && createdObj.GetKind() == "Service" {
					err = scheme.Scheme.Convert(createdObj, retSvc, nil)
					if err != nil {
						fmt.Printf("Failed to save service: %v\n", err)
					}
				}

				if createdObj.GetKind() == "Deployment" {
					var newEnvVars []corev1.EnvVar
					if svcCfg != nil {
						for name, value := range *svcCfg {
							if name == "endpoint" {
								continue
							}
							itemEnvVar := corev1.EnvVar{
								Name:  name,
								Value: value,
							}
							newEnvVars = append(newEnvVars, itemEnvVar)
						}
					}
					if len(newEnvVars) > 0 {
						deployment := &appsv1.Deployment{}
						err = runtime.DefaultUnstructuredConverter.FromUnstructured(createdObj.UnstructuredContent(), deployment)
						if err != nil {
							fmt.Printf("Failed to save deployment: %v\n", err)
						}
						for i := range deployment.Spec.Template.Spec.Containers {
							deployment.Spec.Template.Spec.Containers[i].Env = append(
								deployment.Spec.Template.Spec.Containers[i].Env,
								newEnvVars...)
						}
						modifiedObj, derr := runtime.DefaultUnstructuredConverter.ToUnstructured(deployment)
						if derr != nil {
							fmt.Printf("Failed to marshal updated deployment: %v", derr)
						}

						// Remove managedFields from the unstructured object
						if _, ok := modifiedObj["metadata"].(map[string]interface{}); ok {
							delete(modifiedObj["metadata"].(map[string]interface{}), "managedFields")
						}

						modifiedUnstructured := &unstructured.Unstructured{Object: modifiedObj}
						modifiedBytes, merr := json.Marshal(modifiedUnstructured)
						if merr != nil {
							fmt.Printf("Failed to marshal updated deployment: %v", merr)
						}
						gvr := appsv1.SchemeGroupVersion.WithResource("deployments")
						_, merr = dynamicClient.Resource(gvr).Namespace(ns).Patch(context.TODO(),
							createdObj.GetName(),
							types.ApplyPatchType,
							modifiedBytes,
							metav1.PatchOptions{
								FieldManager: "my-controller",
								Force:        ptr.To(true),
							})
						if merr != nil {
							fmt.Printf("Failed to patch deployment: %v", merr)
						}
					}
				}
				break
			}
			time.Sleep(2 * time.Second)
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

func applyCustomConfig(step string, svcCfg *map[string]string, yamlFile []byte) (string, error) {
	var userDefinedCfg RouterCfg
	if step == "router" {
		userDefinedCfg = RouterCfg{
			NoProxy:    (*svcCfg)["no_proxy"],
			HttpProxy:  (*svcCfg)["http_proxy"],
			HttpsProxy: (*svcCfg)["https_proxy"],
			GRAPH_JSON: (*svcCfg)["nodes"]}
	} else {
		return string(yamlFile), nil
	}

	fmt.Printf("user config %v\n", userDefinedCfg)

	tmpl, err := template.New("yamlTemplate").Parse(string(yamlFile))
	if err != nil {
		return string(yamlFile), fmt.Errorf("error parsing template: %v", err)
	}

	var appliedCfg bytes.Buffer
	err = tmpl.Execute(&appliedCfg, userDefinedCfg)
	if err != nil {
		return string(yamlFile), fmt.Errorf("error executing template: %v", err)
	} else {
		// fmt.Printf("applied config %s\n", appliedCfg.String())
		return appliedCfg.String(), nil
	}
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

	config, err := getKubeConfig()
	if err != nil {
		return reconcile.Result{}, err
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = preProcessUserConfigmap(ctx, dynamicClient, req.NamespacedName.Namespace, xeon, graph)
	if err != nil {
		return reconcile.Result{Requeue: true}, errors.Wrapf(err, "Failed to pre-process the Configmap file for xeon")
	}

	// TODO
	// we need add a config if the hardware is gaudi
	// no matter guadi or xeon, the manifest read configmap by name "qna-config"
	// err = preProcessUserConfigmap(ctx, dynamicClient, req.NamespacedName.Namespace, gaudi, graph)
	// if err != nil {
	// 	return reconcile.Result{Requeue: true}, errors.Wrapf(err, "Failed to pre-process the Configmap file for gaudi")
	// }

	for node, router := range graph.Spec.Nodes {
		for i, step := range router.Steps {
			fmt.Println("\nreconcile resource for node:", step.StepName)

			if step.Executor.ExternalService == "" {
				var ns string
				if step.Executor.InternalService.NameSpace == "" {
					ns = req.Namespace
				} else {
					ns = step.Executor.InternalService.NameSpace
				}
				svcName := step.Executor.InternalService.ServiceName
				fmt.Println("trying to reconcile internal service [", svcName, "] in namespace ", ns)

				service := &corev1.Service{}
				err := reconcileResource(ctx, dynamicClient, step.StepName, ns, svcName, &step.Executor.InternalService.Config, service)
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
	var router_ns string
	if graph.Spec.RouterConfig.NameSpace == "" {
		router_ns = req.Namespace
	} else {
		router_ns = graph.Spec.RouterConfig.NameSpace
	}
	err = r.Client.Get(ctx, types.NamespacedName{Namespace: router_ns, Name: graph.Spec.RouterConfig.ServiceName}, routerService)
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
		err = reconcileResource(ctx, dynamicClient, graph.Spec.RouterConfig.Name, router_ns, graph.Spec.RouterConfig.ServiceName, &graph.Spec.RouterConfig.Config, nil)
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

// read the configmap file from the manifests
// update the values of the fields in the configmap
// add service details to the fields
func preProcessUserConfigmap(ctx context.Context, dynamicClient *dynamic.DynamicClient, ns string, hwType string, gmcGraph *mcv1alpha3.GMConnector) error {
	var cfgFile string
	// var adjustFile string
	if hwType == xeon {
		cfgFile = yaml_dir + "/qna_configmap_xeon.yaml"
	} else if hwType == gaudi {
		cfgFile = yaml_dir + "/qna_configmap_gaudi.yaml"
	} else {
		return fmt.Errorf("unexpected hardware type %s", hwType)
	}
	yamlData, err := os.ReadFile(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to read %s : %v", cfgFile, err)
	}

	fmt.Printf("adjust config: %s\n", cfgFile)

	// Unmarshal the YAML data into a map
	var yamlMap map[string]interface{}
	err = yaml2.Unmarshal(yamlData, &yamlMap)
	if err != nil {
		return fmt.Errorf("failed to unmarshal YAML data: %v", err)
	}

	//adjust the values of the fields defined in manifest configmap file
	adjustConfigmap(ns, hwType, &yamlMap, gmcGraph)

	//NOTE: the filesystem could be read-only, DONOT write it

	adjustedCmBytes, err := yaml2.Marshal(yamlMap)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML data: %v", err)
	}
	_, err = applyResourceToK8s(ctx, dynamicClient, ns, adjustedCmBytes)
	if err != nil {
		return fmt.Errorf("failed to apply the adjusted configmap: %v", err)
	} else {
		fmt.Printf("Success to apply the adjusted configmap\n")
	}

	return nil
}

func applyResourceToK8s(ctx context.Context, dynamicClient *dynamic.DynamicClient, ns string, resource []byte) (*unstructured.Unstructured, error) {
	decUnstructured := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

	obj := &unstructured.Unstructured{}
	_, gvk, err := decUnstructured.Decode(resource, nil, obj)
	if err != nil {
		return nil, fmt.Errorf("failed to decode YAML: %v", err)
	}
	gvr, _ := meta.UnsafeGuessKindToResource(*gvk)
	patchBytes, err := json.Marshal(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config to json: %v", err)
	}

	return dynamicClient.Resource(gvr).Namespace(ns).Patch(ctx, obj.GetName(), types.ApplyPatchType, patchBytes, metav1.PatchOptions{
		FieldManager: "gmc-controller",
		Force:        ptr.To(true),
	})
}

func getNsFromGraph(gmcGraph *mcv1alpha3.GMConnector, stepName string) string {
	for _, router := range gmcGraph.Spec.Nodes {
		for _, step := range router.Steps {
			if step.StepName == stepName {
				// Check if InternalService is not nil
				if step.Executor.ExternalService == "" {
					// Check if NameSpace is not an empty string
					if step.Executor.InternalService.NameSpace != "" {
						return step.Executor.InternalService.NameSpace
					}
				}
			}
		}
	}
	return ""
}

func adjustConfigmap(ns string, hwType string, yamlMap *map[string]interface{}, gmcGraph *mcv1alpha3.GMConnector) {
	var embdManifest string
	var rerankManifest string
	var tgiManifest string
	var redisManifest string
	if hwType == xeon {
		embdManifest = yaml_dir + tei_embedding_service_yaml
		rerankManifest = yaml_dir + tei_reranking_service_yaml
		tgiManifest = yaml_dir + tgi_service_yaml
		redisManifest = yaml_dir + redis_vector_db_yaml
	} else if hwType == gaudi {
		embdManifest = yaml_dir + tei_embedding_gaudi_service_yaml
		rerankManifest = yaml_dir + tei_reranking_service_yaml
		tgiManifest = yaml_dir + tgi_gaudi_service_yaml
		redisManifest = yaml_dir + redis_vector_db_yaml
	} else {
		fmt.Printf("unexpected hardware type %s", hwType)
		return
	}
	if data, ok := (*yamlMap)["data"].(map[interface{}]interface{}); ok {
		// Update the value of "TEI_EMBEDDING_ENDPOINT" field
		if _, ok := data["TEI_EMBEDDING_ENDPOINT"].(string); ok {
			svcName, port, err := getServiceDetailsFromManifests(embdManifest)
			if err == nil {
				//check GMC config if there is specific namespace for embedding
				altNs := getNsFromGraph(gmcGraph, TeiEmbedding)
				if altNs != "" {
					data["TEI_EMBEDDING_ENDPOINT"] = fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", svcName, altNs, port)
				} else {
					data["TEI_EMBEDDING_ENDPOINT"] = fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", svcName, ns, port)
				}
			} else {
				fmt.Printf("failed to get service details for %s: %v\n", embdManifest, err)
			}
		} else {
			fmt.Printf("failed to get data for TEI_EMBEDDING_ENDPOINT\n")
		}
		// Update the value of "TEI_RERANKING_ENDPOINT" field
		if _, ok = data["TEI_RERANKING_ENDPOINT"].(string); ok {
			svcName, port, err := getServiceDetailsFromManifests(rerankManifest)
			if err == nil {
				//check GMC config if there is specific namespace for reranking
				altNs := getNsFromGraph(gmcGraph, TeiReranking)
				if altNs != "" {
					data["TEI_RERANKING_ENDPOINT"] = fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", svcName, altNs, port)
				} else {
					data["TEI_RERANKING_ENDPOINT"] = fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", svcName, ns, port)
				}
			} else {
				fmt.Printf("failed to get service details for %s: %v\n", rerankManifest, err)
			}
		} else {
			fmt.Printf("failed to get data for TEI_RERANKING_ENDPOINT\n")
		}
		// Update the value of "TGI_LLM_ENDPOINT" field
		if _, ok = data["TGI_LLM_ENDPOINT"].(string); ok {
			svcName, port, err := getServiceDetailsFromManifests(tgiManifest)
			if err == nil {
				//check GMC config if there is specific namespace for tgillm
				altNs := getNsFromGraph(gmcGraph, Tgi)
				if altNs != "" {
					data["TGI_LLM_ENDPOINT"] = fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", svcName, altNs, port)
				} else {
					data["TGI_LLM_ENDPOINT"] = fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", svcName, ns, port)
				}
			} else {
				fmt.Printf("failed to get service details for %s: %v\n", tgiManifest, err)
			}
		} else {
			fmt.Printf("failed to get data for TGI_LLM_ENDPOINT\n")
		}
		// Update the value of "REDIS_URL" field
		if _, ok = data["REDIS_URL"].(string); ok {
			svcName, port, err := getServiceDetailsFromManifests(redisManifest)
			if err == nil {
				//check GMC config if there is specific namespace for tgillm
				altNs := getNsFromGraph(gmcGraph, Tgi)
				if altNs == "" {
					altNs = ns
				}
				data["REDIS_URL"] = fmt.Sprintf("redis://%s.%s.svc.cluster.local:%d", svcName, altNs, port)
			} else {
				fmt.Printf("failed to get service details for %s: %v\n", redisManifest, err)
			}
		} else {
			fmt.Printf("failed to get data for REDIS_URL\n")
		}
	} else {
		fmt.Printf("failed to interpret data %v\n", data)
	}
}

type YAMLContent struct {
	Service struct {
		Kind     string `yaml:"kind"`
		Metadata struct {
			Name string `yaml:"name"`
		} `yaml:"metadata"`
		Spec struct {
			Ports []struct {
				Name string `yaml:"name"`
				Port int    `yaml:"port"`
			} `yaml:"ports"`
		} `yaml:"spec"`
	} `yaml:"services"`
}

func getServiceDetailsFromManifests(filePath string) (string, int, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", 0, err
	}
	resources := strings.Split(string(data), "---")

	for _, res := range resources {
		if res == "" || !strings.Contains(res, "kind: Service") {
			continue
		}
		var content YAMLContent
		err = yaml2.Unmarshal([]byte(res), &content.Service)
		if err != nil {
			return "", 0, err
		}
		if content.Service.Kind == "Service" {
			if len(content.Service.Spec.Ports) > 0 {
				return content.Service.Metadata.Name, content.Service.Spec.Ports[0].Port, nil
			}
		}

	}

	return "", 0, fmt.Errorf("service name or port not found")
}

// SetupWithManager sets up the controller with the Manager.
func (r *GMConnectorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mcv1alpha3.GMConnector{}).
		Complete(r)
}
