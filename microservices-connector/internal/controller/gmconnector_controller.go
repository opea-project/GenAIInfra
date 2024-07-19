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
	DocSum                           = "DocSum"
	DocSumGaudi                      = "DocSumGaudi"
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
	docsum_llm_yaml                  = "/docsum_llm.yaml"
	docsum_gaudi_llm_yaml            = "/docsum_gaudi_llm.yaml"
	yaml_dir                         = "/tmp/microservices/yamls"
	Service                          = "Service"
	Deployment                       = "Deployment"
	dplymtSubfix                     = "-deployment"
	METADATA_PLATFORM                = "gmc/platform"
)

var yamlDict = map[string]string{TeiEmbedding: yaml_dir + "/tei.yaml",
	TeiEmbeddingGaudi: yaml_dir + "/tei_gaudi.yaml",
	Embedding:         yaml_dir + "/embedding-usvc.yaml",
	VectorDB:          yaml_dir + "/redis-vector-db.yaml",
	Retriever:         yaml_dir + "/retriever-usvc.yaml",
	Reranking:         yaml_dir + "/reranking-usvc.yaml",
	TeiReranking:      yaml_dir + "/teirerank.yaml",
	Tgi:               yaml_dir + "/tgi.yaml",
	TgiGaudi:          yaml_dir + "/tgi_gaudi.yaml",
	Llm:               yaml_dir + "/llm-uservice.yaml",
	DocSum:            yaml_dir + "/docsum-llm-uservice.yaml",
	Router:            yaml_dir + "/gmc-router.yaml",
}

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

func lookupManifestDir(step string) string {
	value, exist := yamlDict[step]
	if exist {
		return value
	} else {
		return ""
	}
}

// func getManifestYaml(step string) string {
// 	var tmpltFile string
// 	//TODO add validation to rule out unexpected case like both embedding and retrieving
// 	if step == Embedding {
// 		tmpltFile = yaml_dir + embedding_yaml
// 	} else if step == TeiEmbedding {
// 		tmpltFile = yaml_dir + tei_embedding_service_yaml
// 	} else if step == TeiEmbeddingGaudi {
// 		tmpltFile = yaml_dir + tei_embedding_gaudi_service_yaml
// 	} else if step == VectorDB {
// 		tmpltFile = yaml_dir + redis_vector_db_yaml
// 	} else if step == Retriever {
// 		tmpltFile = yaml_dir + retriever_yaml
// 	} else if step == Reranking {
// 		tmpltFile = yaml_dir + reranking_yaml
// 	} else if step == TeiReranking {
// 		tmpltFile = yaml_dir + tei_reranking_service_yaml
// 	} else if step == Tgi {
// 		tmpltFile = yaml_dir + tgi_service_yaml
// 	} else if step == TgiGaudi {
// 		tmpltFile = yaml_dir + tgi_gaudi_service_yaml
// 	} else if step == Llm {
// 		tmpltFile = yaml_dir + llm_yaml
// 	} else if step == DocSum {
// 		tmpltFile = yaml_dir + docsum_llm_yaml
// 	} else if step == DocSumGaudi {
// 		tmpltFile = yaml_dir + docsum_gaudi_llm_yaml
// 	} else if step == Router {
// 		tmpltFile = yaml_dir + gmc_router_yaml
// 	} else {
// 		return ""
// 	}
// 	return tmpltFile
// }

func reconcileResource(ctx context.Context, client client.Client, graphNs string, stepCfg *mcv1alpha3.Step, nodeCfg *mcv1alpha3.Router) ([]*unstructured.Unstructured, error) {
	if stepCfg == nil || nodeCfg == nil {
		return nil, errors.New("invalid svc config")
	}

	fmt.Printf("get resource config: %v\n", *stepCfg)

	var retObjs []*unstructured.Unstructured
	// by default, the svc's namespace is the same as the graph
	// unless it's specifically defined in yaml
	ns := graphNs
	if stepCfg.InternalService.NameSpace != "" {
		ns = stepCfg.InternalService.NameSpace
	}
	svc := stepCfg.InternalService.ServiceName
	svcCfg := &stepCfg.InternalService.Config

	yamlFile, err := getTemplateBytes(stepCfg.StepName)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %v", err)
	}

	resources := strings.Split(string(yamlFile), "---")
	fmt.Printf("The raw yaml file has been split into %v yaml files\n", len(resources))

	for _, res := range resources {
		if res == "" || !strings.Contains(res, "kind:") {
			continue
		}

		decUnstructured := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		obj := &unstructured.Unstructured{}
		_, _, err := decUnstructured.Decode([]byte(res), nil, obj)
		if err != nil {
			return nil, fmt.Errorf("failed to decode YAML: %v", err)
		}

		// Set the namespace according to user defined value
		if ns != "" {
			obj.SetNamespace(ns)
		}

		// set the service name according to user defined value, and related selectors/labels
		if obj.GetKind() == Service && svc != "" {
			service_obj := &corev1.Service{}
			err = scheme.Scheme.Convert(obj, service_obj, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to convert unstructured to service: %v", err)
			}
			service_obj.SetName(svc)
			service_obj.Spec.Selector["app"] = svc
			err = scheme.Scheme.Convert(service_obj, obj, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to convert unstructured to service: %v", err)
			}
		} else if obj.GetKind() == Deployment {
			deployment_obj := &appsv1.Deployment{}
			err = scheme.Scheme.Convert(obj, deployment_obj, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to convert unstructured to deployment: %v", err)
			}
			if svc != "" {
				deployment_obj.SetName(svc + dplymtSubfix)
				// Set the labels if they're specified
				deployment_obj.Spec.Selector.MatchLabels["app"] = svc
				deployment_obj.Spec.Template.Labels["app"] = svc
			}

			// append the user defined ENVs
			var newEnvVars []corev1.EnvVar
			if svcCfg != nil {
				for name, value := range *svcCfg {
					if name == "endpoint" || name == "nodes" {
						continue
					}
					if keyIsSomeEndpoint(name) {
						ds := findDownStreamService(value, stepCfg, nodeCfg)
						value, err = getDownstreamSvcEndpoint(graphNs, value, ds)
						// value = getDsEndpoint(platform, name, graphNs, ds)
						if err != nil {
							return nil, fmt.Errorf("failed to find downstream service endpoint: %v", err)
						}
					}
					itemEnvVar := corev1.EnvVar{
						Name:  name,
						Value: value,
					}
					newEnvVars = append(newEnvVars, itemEnvVar)
				}
			}
			if len(newEnvVars) > 0 {
				for i := range deployment_obj.Spec.Template.Spec.Containers {
					deployment_obj.Spec.Template.Spec.Containers[i].Env = append(
						deployment_obj.Spec.Template.Spec.Containers[i].Env,
						newEnvVars...)
				}
			}

			err = scheme.Scheme.Convert(deployment_obj, obj, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to convert unstructured to deployment: %v", err)
			}
		}

		err = applyResourceToK8s(ctx, client, obj)
		if err != nil {
			return nil, fmt.Errorf("failed to reconcile resource: %v", err)
		} else {
			fmt.Printf("Success to reconcile %s: %s\n", obj.GetKind(), obj.GetName())
			retObjs = append(retObjs, obj)
		}
	}
	return retObjs, nil
}

func keyIsSomeEndpoint(keyname string) bool {
	return keyname == "TEI_EMBEDDING_ENDPOINT" || keyname == "TEI_RERANKING_ENDPOINT" || keyname == "TGI_LLM_ENDPOINT" || keyname == "REDIS_URL"
}

func findDownStreamService(dsName string, stepCfg *mcv1alpha3.Step, nodeCfg *mcv1alpha3.Router) *mcv1alpha3.Step {
	if stepCfg == nil || nodeCfg == nil {
		return nil
	}
	for _, otherStep := range nodeCfg.Steps {
		if otherStep.InternalService.ServiceName == dsName && otherStep.InternalService.IsDownstreamService {
			return &otherStep
		}
	}
	return nil
}

func getDownstreamSvcEndpoint(graphNs string, dsName string, stepCfg *mcv1alpha3.Step) (string, error) {
	if stepCfg == nil {
		return "", errors.New(fmt.Sprintf("empty stepCfg for %s", dsName))
	}
	tmplt := lookupManifestDir(dsName)
	if tmplt == "" {
		return "", errors.New(fmt.Sprintf("failed to find yaml file for %s", dsName))
	}

	svcName, port, err := getServiceDetailsFromManifests(tmplt)
	if err == nil {
		//check GMC config if there is specific namespace for embedding
		altNs, altSvcName := getNsNameFromStep(stepCfg)
		if altNs == "" {
			altNs = graphNs
		}
		if altSvcName == "" {
			altSvcName = svcName
		}

		if dsName == VectorDB {
			return fmt.Sprintf("redis://%s.%s.svc.cluster.local:%d", altSvcName, altNs, port), nil
		} else {
			return fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", altSvcName, altNs, port), nil
		}
	} else {
		return "", errors.New(fmt.Sprintf("failed to get service details for %s: %v\n", dsName, err))
	}
}

// func getDsEndpoint(platform string, keyname string, reqNS string, ds *mcv1alpha3.Step) string {
// 	if ds == nil {
// 		return ""
// 	}
// 	var embdManifest string
// 	var rerankManifest string
// 	var tgiManifest string
// 	var redisManifest string
// 	if platform == xeon {
// 		embdManifest = yaml_dir + tei_embedding_service_yaml
// 		rerankManifest = yaml_dir + tei_reranking_service_yaml
// 		tgiManifest = yaml_dir + tgi_service_yaml
// 		redisManifest = yaml_dir + redis_vector_db_yaml
// 	} else if platform == gaudi {
// 		embdManifest = yaml_dir + tei_embedding_gaudi_service_yaml
// 		rerankManifest = yaml_dir + tei_reranking_service_yaml
// 		tgiManifest = yaml_dir + tgi_gaudi_service_yaml
// 		redisManifest = yaml_dir + redis_vector_db_yaml
// 	} else {
// 		fmt.Printf("unexpected hardware type %s", platform)
// 		return ""
// 	}

// 	var svcName string
// 	var port int
// 	var err error

// 	if keyname == "TEI_EMBEDDING_ENDPOINT" {
// 		svcName, port, err = getServiceDetailsFromManifests(embdManifest)
// 	} else if keyname == "TEI_RERANKING_ENDPOINT" {
// 		svcName, port, err = getServiceDetailsFromManifests(rerankManifest)
// 	} else if keyname == "TGI_LLM_ENDPOINT" {
// 		svcName, port, err = getServiceDetailsFromManifests(tgiManifest)
// 	} else if keyname == "REDIS_URL" {
// 		svcName, port, err = getServiceDetailsFromManifests(redisManifest)
// 	}

// 	if err == nil {
// 		//check GMC config if there is specific namespace for embedding
// 		altNs, altSvcName := getNsNameFromStep(ds)
// 		if altNs == "" {
// 			altNs = reqNS
// 		}
// 		if altSvcName == "" {
// 			altSvcName = svcName
// 		}
// 		return fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", altSvcName, altNs, port)
// 	} else {
// 		fmt.Printf("failed to get service details for %s: %v\n", embdManifest, err)
// 	}
// 	return ""

// }

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

	// platform := xeon
	// if labelValue, exists := graph.GetLabels()[METADATA_PLATFORM]; exists {
	// 	platform = labelValue
	// }

	// TO BE DELETED
	// this is deprecated when new manifests in merged
	// err := preProcessUserConfigmap(ctx, r.Client, graph.Namespace, platform)
	// if err != nil {
	// 	return reconcile.Result{Requeue: true}, errors.Wrapf(err, "Failed to pre-process the Configmap file")
	// }

	var totalService uint
	var externalService uint
	var successService uint

	for nodeName, node := range graph.Spec.Nodes {
		for i, step := range node.Steps {
			if step.NodeName != "" {
				fmt.Println("\nthis is a nested step: ", step.StepName)
				continue
			}
			fmt.Println("\nreconcile resource for node:", step.StepName)
			totalService += 1
			if step.Executor.ExternalService == "" {
				fmt.Println("trying to reconcile internal service [", step.Executor.InternalService.ServiceName, "] in namespace ", step.Executor.InternalService.NameSpace)

				// err := reconcileResource(ctx, r.Client, step.StepName, ns, svcName, &step.Executor.InternalService.Config, service)
				objs, err := reconcileResource(ctx, r.Client, graph.Namespace, &step, &node)
				if err != nil {
					return reconcile.Result{Requeue: true}, errors.Wrapf(err, "Failed to reconcile service for %s", step.Executor.InternalService.ServiceName)
				}
				if len(objs) != 0 {
					for _, obj := range objs {
						if obj.GetKind() == Service {
							service := &corev1.Service{}
							err = scheme.Scheme.Convert(obj, service, nil)
							if err != nil {
								return reconcile.Result{Requeue: true}, errors.Wrapf(err, "Failed to reconcile service")
							}
							graph.Spec.Nodes[nodeName].Steps[i].ServiceURL = getServiceURL(service) + step.InternalService.Config["endpoint"]
							fmt.Printf("the service URL is: %s\n", graph.Spec.Nodes[nodeName].Steps[i].ServiceURL)
							successService += 1
						}
					}
				}
			} else {
				fmt.Println("external service is found", "name", step.ExternalService)
				graph.Spec.Nodes[nodeName].Steps[i].ServiceURL = step.ExternalService
				externalService += 1
			}
		}
		fmt.Println()
	}

	//to start a router service
	//in case the graph changes, we need to apply the changes to router service
	//so we need to apply the router config every time
	err := reconcileRouterService(ctx, r.Client, graph)
	if err != nil {
		return reconcile.Result{Requeue: true}, errors.Wrapf(err, "Failed to reconcile router service")
	}

	graph.Status.Status = fmt.Sprintf("%d/%d/%d", successService, externalService, totalService)
	if err = r.Status().Update(context.TODO(), graph); err != nil {
		return reconcile.Result{Requeue: true}, errors.Wrapf(err, "Failed to Update CR status to %s", graph.Status.Status)
	}
	return ctrl.Result{}, nil
}

func getTemplateBytes(resourceType string) ([]byte, error) {
	tmpltFile := lookupManifestDir(resourceType)
	if tmpltFile == "" {
		return nil, errors.New("unexpected target")
	}
	yamlBytes, err := os.ReadFile(tmpltFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %v", err)
	}
	return yamlBytes, nil
}

func reconcileRouterService(ctx context.Context, client client.Client, graph *mcv1alpha3.GMConnector) error {
	routerService := &corev1.Service{}
	jsonBytes, err := json.Marshal(graph)
	if err != nil {
		// handle error
		return errors.Wrapf(err, "Failed to Marshal routes for %s", graph.Spec.RouterConfig.Name)
	}
	jsonString := string(jsonBytes)
	if graph.Spec.RouterConfig.Config == nil {
		graph.Spec.RouterConfig.Config = make(map[string]string)
	}
	graph.Spec.RouterConfig.Config["nodes"] = "'" + jsonString + "'"

	templateBytes, err := getTemplateBytes(Router)
	if err != nil {
		return errors.Wrapf(err, "Failed to get template bytes for %s", Router)
	}
	var resources []string
	appliedCfg, err := applyRouterConfigToTemplates(Router, &graph.Spec.RouterConfig.Config, templateBytes)
	if err != nil {
		return fmt.Errorf("failed to apply user config: %v", err)
	}

	resources = strings.Split(appliedCfg, "---")
	for _, res := range resources {
		if res == "" || !strings.Contains(res, "kind:") {
			continue
		}
		decUnstructured := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		obj := &unstructured.Unstructured{}
		_, _, err := decUnstructured.Decode([]byte(res), nil, obj)
		if err != nil {
			return fmt.Errorf("failed to decode YAML: %v", err)
		}

		if graph.Spec.RouterConfig.NameSpace != "" {
			obj.SetNamespace(graph.Spec.RouterConfig.NameSpace)
		} else {
			obj.SetNamespace(graph.Namespace)
		}

		err = applyResourceToK8s(ctx, client, obj)
		if err != nil {
			return fmt.Errorf("failed to reconcile resource: %v", err)
		} else {
			fmt.Printf("Success to reconcile %s: %s\n", obj.GetKind(), obj.GetName())
		}
		if obj.GetKind() == Service {
			err = scheme.Scheme.Convert(obj, routerService, nil)
			if err != nil {
				return fmt.Errorf("failed to save router service: %v", err)
			}
			graph.Status.AccessURL = getServiceURL(routerService)
			fmt.Printf("the router service URL is: %s\n", graph.Status.AccessURL)
		}
	}
	return nil
}

func applyRouterConfigToTemplates(step string, svcCfg *map[string]string, yamlFile []byte) (string, error) {
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

// TO BE DELETED
// this is deprecated when new manifests in merged
// read the configmap file from the manifests
// update the values of the fields in the configmap
// add service details to the fields
// func preProcessUserConfigmap(ctx context.Context, client client.Client, ns string, hwType string) error {
// 	var cfgFile string
// 	// var adjustFile string
// 	if hwType == xeon {
// 		cfgFile = yaml_dir + "/qna_configmap_xeon.yaml"
// 	} else if hwType == gaudi {
// 		cfgFile = yaml_dir + "/qna_configmap_gaudi.yaml"
// 	} else {
// 		return fmt.Errorf("unexpected hardware type %s", hwType)
// 	}
// 	yamlData, err := os.ReadFile(cfgFile)
// 	if err != nil {
// 		return fmt.Errorf("failed to read %s : %v", cfgFile, err)
// 	}

// 	decUnstructured := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
// 	obj := &unstructured.Unstructured{}
// 	_, _, err = decUnstructured.Decode(yamlData, nil, obj)
// 	if err != nil {
// 		return fmt.Errorf("failed to decode configmap YAML: %v", err)
// 	}
// 	obj.SetNamespace(ns)

// 	err = applyResourceToK8s(ctx, client, obj)
// 	if err != nil {
// 		return fmt.Errorf("failed to apply the adjusted configmap: %v", err)
// 	} else {
// 		fmt.Printf("Success to apply the adjusted configmap\n")
// 	}

// 	return nil
// }

func applyResourceToK8s(ctx context.Context, c client.Client, obj *unstructured.Unstructured) error {
	// Prepare the object for an update, assuming it already exists. If it doesn't, you'll need to handle that case.
	// This might involve trying an Update and, if it fails because the object doesn't exist, falling back to Create.
	// Retry updating the resource in case of transient errors.
	timeout := time.After(1 * time.Minute)
	tick := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled")
		case <-timeout:
			return fmt.Errorf("timed out while trying to update or create resource")
		case <-tick.C:
			// Get the latest version of the object
			latest := &unstructured.Unstructured{}
			latest.SetGroupVersionKind(obj.GroupVersionKind())
			err := c.Get(ctx, client.ObjectKeyFromObject(obj), latest)
			if err != nil {
				if apierr.IsNotFound(err) {
					// If the object doesn't exist, create it
					err = c.Create(ctx, obj, &client.CreateOptions{})
					if err != nil {
						return fmt.Errorf("failed to create resource: %v", err)
					}
				} else {
					// If there was another error, continue
					fmt.Printf("get object err: %v", err)
					continue
				}
			} else {
				// If the object does exist, update it
				obj.SetResourceVersion(latest.GetResourceVersion()) // Ensure we're updating the latest version
				err = c.Update(ctx, obj, &client.UpdateOptions{})
				if err != nil {
					fmt.Printf("\nupdate object err: %v", err)
					continue
				}
			}

			// If we reach this point, the operation was successful.
			return nil
		}
	}
}

func getNsNameFromStep(step *mcv1alpha3.Step) (string, string) {
	var retNs string
	var retName string

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

	return retNs, retName
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
