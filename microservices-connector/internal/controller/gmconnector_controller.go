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
	"reflect"
	"strings"
	"text/template"
	"time"

	mcv1alpha3 "github.com/opea-project/GenAIInfra/microservices-connector/api/v1alpha3"
	"github.com/pkg/errors"
	yaml2 "gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
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
	Service                          = "Service"
	Deployment                       = "Deployment"
	dplymtSubfix                     = "-deployment"
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

func getManifestYaml(step string) string {
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

func reconcileResource(ctx context.Context, client client.Client, step string, ns string, svc string, svcCfg *map[string]string, retSvc *corev1.Service) error {
	fmt.Printf("get step %s config for %s@%s: %v\n", step, svc, ns, svcCfg)
	tmpltFile := getManifestYaml(step)
	if tmpltFile == "" {
		return errors.New("unexpected target")
	}
	yamlFile, err := os.ReadFile(tmpltFile)
	if err != nil {
		return fmt.Errorf("failed to read YAML file: %v", err)
	}

	var resources []string
	appliedCfg, err := patchCustomConfigToTemplates(step, svcCfg, yamlFile)
	if err != nil {
		return fmt.Errorf("failed to apply user config: %v", err)
	}
	resources = strings.Split(appliedCfg, "---")
	fmt.Printf("The raw yaml file has been split into %v yaml files\n", len(resources))

	for _, res := range resources {
		if res == "" || !strings.Contains(res, "kind:") {
			continue
		}
		createdObj, err := applyResourceToK8s(ctx, client, ns, svc, []byte(res))
		if err != nil {
			return fmt.Errorf("Failed to reconcile resource: %v\n", err)
		} else {
			fmt.Printf("Success to reconcile %s: %s\n", createdObj.GetKind(), createdObj.GetName())

			// return the service obj to get the service URL from it
			if retSvc != nil && createdObj.GetKind() == Service {
				err = scheme.Scheme.Convert(createdObj, retSvc, nil)
				if err != nil {
					return fmt.Errorf("Failed to save service: %v\n", err)
				}
			}

			if createdObj.GetKind() == Deployment && step != Router {
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
						return fmt.Errorf("Failed to save deployment: %v\n", err)
					}
					for i := range deployment.Spec.Template.Spec.Containers {
						deployment.Spec.Template.Spec.Containers[i].Env = append(
							deployment.Spec.Template.Spec.Containers[i].Env,
							newEnvVars...)
					}

					// Update the deployment using client.Client
					if err := client.Update(ctx, deployment); err != nil {
						return fmt.Errorf("Failed to update deployment: %v\n", err)
					}
				}
			}
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

func patchCustomConfigToTemplates(step string, svcCfg *map[string]string, yamlFile []byte) (string, error) {
	var userDefinedCfg RouterCfg
	if step == "router" {
		userDefinedCfg = RouterCfg{
			NoProxy:    (*svcCfg)["no_proxy"],
			HttpProxy:  (*svcCfg)["http_proxy"],
			HttpsProxy: (*svcCfg)["https_proxy"],
			GRAPH_JSON: (*svcCfg)["nodes"]}
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
	} else {
		return string(yamlFile), nil
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

	err := preProcessUserConfigmap(ctx, r.Client, req.NamespacedName.Namespace, xeon, graph)
	if err != nil {
		return reconcile.Result{Requeue: true}, errors.Wrapf(err, "Failed to pre-process the Configmap file for xeon")
	}

	var totalService uint
	var externalService uint
	var successService uint

	for nodeName, node := range graph.Spec.Nodes {
		for i, step := range node.Steps {
			fmt.Println("\nreconcile resource for node:", step.StepName)
			totalService += 1
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
				err := reconcileResource(ctx, r.Client, step.StepName, ns, svcName, &step.Executor.InternalService.Config, service)
				if err != nil {
					return reconcile.Result{Requeue: true}, errors.Wrapf(err, "Failed to reconcile service for %s", svcName)
				}
				successService += 1
				graph.Spec.Nodes[nodeName].Steps[i].ServiceURL = getServiceURL(service) + step.Executor.InternalService.Config["endpoint"]
				fmt.Printf("the service URL is: %s\n", graph.Spec.Nodes[nodeName].Steps[i].ServiceURL)

			} else {
				fmt.Println("external service is found", "name", step.ExternalService)
				graph.Spec.Nodes[nodeName].Steps[i].ServiceURL = step.ExternalService
				externalService += 1
			}
		}
		fmt.Println()
	}

	//to start a router controller
	//in case the graph changes, we need to apply the changes to router service
	//so we need to apply the router config every time
	routerService := &corev1.Service{}
	var router_ns string
	if graph.Spec.RouterConfig.NameSpace == "" {
		router_ns = req.Namespace
	} else {
		router_ns = graph.Spec.RouterConfig.NameSpace
	}

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
	//set empty service name, because we don't want to change router service name and deployment name
	err = reconcileResource(ctx, r.Client, graph.Spec.RouterConfig.Name, router_ns, graph.Spec.RouterConfig.ServiceName, &graph.Spec.RouterConfig.Config, routerService)
	if err != nil {
		return reconcile.Result{Requeue: true}, errors.Wrapf(err, "Failed to reconcile router service")
	}

	graph.Status.AccessURL = getServiceURL(routerService)
	fmt.Printf("the router service URL is: %s\n", graph.Status.AccessURL)

	graph.Status.Status = fmt.Sprintf("%d/%d/%d", successService, externalService, totalService)
	if err = r.Status().Update(context.TODO(), graph); err != nil {
		return reconcile.Result{Requeue: true}, errors.Wrapf(err, "Failed to Update CR status to %s", graph.Status.Status)
	}
	return ctrl.Result{}, nil
}

// read the configmap file from the manifests
// update the values of the fields in the configmap
// add service details to the fields
func preProcessUserConfigmap(ctx context.Context, client client.Client, ns string, hwType string, gmcGraph *mcv1alpha3.GMConnector) error {
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
	_, err = applyResourceToK8s(ctx, client, ns, "", adjustedCmBytes)
	if err != nil {
		return fmt.Errorf("failed to apply the adjusted configmap: %v", err)
	} else {
		fmt.Printf("Success to apply the adjusted configmap\n")
	}

	return nil
}

func applyResourceToK8s(ctx context.Context, c client.Client, ns string, svc string, resource []byte) (*unstructured.Unstructured, error) {
	decUnstructured := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

	obj := &unstructured.Unstructured{}
	_, _, err := decUnstructured.Decode(resource, nil, obj)
	if err != nil {
		return nil, fmt.Errorf("failed to decode YAML: %v", err)
	}

	// Set the namespace if it's specified
	if ns != "" {
		obj.SetNamespace(ns)
	}

	if svc != "" {
		if obj.GetKind() == Service {
			obj.SetName(svc)
			selectors, found, err := unstructured.NestedStringMap(obj.Object, "spec", "selector")
			if err != nil {
				return nil, fmt.Errorf("failed to get selectors: %v", err)
			}
			if found {
				selectors["app"] = svc // Set the new selector.app value
				err = unstructured.SetNestedStringMap(obj.Object, selectors, "spec", "selector")
				if err != nil {
					return nil, fmt.Errorf("failed to set new selector.app: %v", err)
				}
			}
		}
		if obj.GetKind() == Deployment {
			obj.SetName(svc + dplymtSubfix)
			// Set the labels if they're specified
			labels, found, err := unstructured.NestedStringMap(obj.Object, "spec", "selector", "matchLabels")
			if err != nil {
				return nil, fmt.Errorf("failed to get spec.selector.matchLabels: %v", err)
			}
			if found {
				labels["app"] = svc
				err = unstructured.SetNestedStringMap(obj.Object, labels, "spec", "selector", "matchLabels")
				if err != nil {
					return nil, fmt.Errorf("failed to set new spec.selector.matchLabels : %v", err)
				}
			}
			// Set the labels in template if they're specified
			labels, found, err = unstructured.NestedStringMap(obj.Object, "spec", "template", "metadata", "labels")
			if err != nil {
				return nil, fmt.Errorf("failed to get spec.template.metadata.labels: %v", err)
			}
			if found {
				labels["app"] = svc
				err = unstructured.SetNestedStringMap(obj.Object, labels, "spec", "template", "metadata", "labels")
				if err != nil {
					return nil, fmt.Errorf("failed to set spec.template.metadata.labels: %v", err)
				}
			}
		}

	}

	// Prepare the object for an update, assuming it already exists. If it doesn't, you'll need to handle that case.
	// This might involve trying an Update and, if it fails because the object doesn't exist, falling back to Create.
	// Retry updating the resource in case of transient errors.
	timeout := time.After(1 * time.Minute)
	tick := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled")
		case <-timeout:
			return nil, fmt.Errorf("timed out while trying to update or create resource")
		case <-tick.C:
			// Get the latest version of the object
			latest := &unstructured.Unstructured{}
			latest.SetGroupVersionKind(obj.GroupVersionKind())
			err = c.Get(ctx, client.ObjectKeyFromObject(obj), latest)
			if err != nil {
				if apierr.IsNotFound(err) {
					// If the object doesn't exist, create it
					err = c.Create(ctx, obj, &client.CreateOptions{})
					if err != nil {
						return nil, fmt.Errorf("failed to create resource: %v", err)
					}
				} else {
					// If there was another error, continue
					continue
				}
			} else {
				// If the object does exist, update ithui
				obj.SetResourceVersion(latest.GetResourceVersion()) // Ensure we're updating the latest version
				err = c.Update(ctx, obj, &client.UpdateOptions{})
				if err != nil {
					continue
				}
			}

			// If we reach this point, the operation was successful.
			return obj, nil
		}
	}
}

func getNsNameFromGraph(gmcGraph *mcv1alpha3.GMConnector, stepName string) (string, string) {
	var retNs string
	var retName string
	for _, router := range gmcGraph.Spec.Nodes {
		for _, step := range router.Steps {
			if step.StepName == stepName {
				// Check if InternalService is not nil
				if step.Executor.ExternalService == "" {
					// Check if NameSpace is not an empty string
					if step.Executor.InternalService.NameSpace != "" {
						retNs = step.Executor.InternalService.NameSpace
					}
					if step.Executor.InternalService.ServiceName != "" {
						retName = step.Executor.InternalService.ServiceName
					}
				}
			}
		}
	}
	return retNs, retName
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
				altNs, altSvcName := getNsNameFromGraph(gmcGraph, TeiEmbedding)
				if altNs == "" {
					altNs = ns
				}
				if altSvcName == "" {
					altSvcName = svcName
				}
				data["TEI_EMBEDDING_ENDPOINT"] = fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", altSvcName, altNs, port)
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
				altNs, altSvcName := getNsNameFromGraph(gmcGraph, TeiReranking)
				if altNs == "" {
					altNs = ns
				}
				if altSvcName == "" {
					altSvcName = svcName
				}
				data["TEI_RERANKING_ENDPOINT"] = fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", altSvcName, altNs, port)
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
				altNs, altSvcName := getNsNameFromGraph(gmcGraph, Tgi)
				if altNs == "" {
					altNs = ns
				}
				if altSvcName == "" {
					altSvcName = svcName
				}
				data["TGI_LLM_ENDPOINT"] = fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", altSvcName, altNs, port)
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
				altNs, altSvcName := getNsNameFromGraph(gmcGraph, VectorDB)
				if altNs == "" {
					altNs = ns
				}
				if altSvcName == "" {
					altSvcName = svcName
				}
				data["REDIS_URL"] = fmt.Sprintf("redis://%s.%s.svc.cluster.local:%d", altSvcName, altNs, port)
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
		svc := &corev1.Service{}
		decoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		_, _, err = decoder.Decode([]byte(res), nil, svc)
		if err != nil {
			return "", 0, err
		}
		if svc.Kind == "Service" {
			if len(svc.Spec.Ports) > 0 {
				return svc.Name, int(svc.Spec.Ports[0].Port), nil
			}
		}

	}

	return "", 0, fmt.Errorf("service name or port not found")
}

// SetupWithManager sets up the controller with the Manager.
func (r *GMConnectorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Predicate to ignore updates to status subresource
	ignoreStatusUpdatePredicate := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Cast objects to your GMConnector struct
			oldObject, ok1 := e.ObjectOld.(*mcv1alpha3.GMConnector)
			newObject, ok2 := e.ObjectNew.(*mcv1alpha3.GMConnector)
			if !ok1 || !ok2 {
				// Not the correct type, allow the event through
				return true
			}

			specChanged := !reflect.DeepEqual(oldObject.Spec, newObject.Spec)
			metadataChanged := !reflect.DeepEqual(oldObject.ObjectMeta, newObject.ObjectMeta)

			fmt.Printf("\nspec changed %t | meta changed: %t\n", specChanged, metadataChanged)

			// Compare the old and new spec, ignore metadata, status changes
			// metadata change: name, namespace, such change should create a new GMC
			// status change: depoyment status
			return specChanged
		},
		// Other funcs like CreateFunc, DeleteFunc, GenericFunc can be left as default
		// if you only want to customize the UpdateFunc behavior.
	}

	// Setup the watch with the predicate to filter events
	return ctrl.NewControllerManagedBy(mgr).
		For(&mcv1alpha3.GMConnector{}).
		WithEventFilter(ignoreStatusUpdatePredicate). // Use the predicate here
		Complete(r)
}
