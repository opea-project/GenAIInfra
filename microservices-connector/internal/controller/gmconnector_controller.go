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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	Configmap                = "Configmap"
	ConfigmapGaudi           = "ConfigmapGaudi"
	Embedding                = "Embedding"
	TeiEmbedding             = "TeiEmbedding"
	TeiEmbeddingGaudi        = "TeiEmbeddingGaudi"
	VectorDB                 = "VectorDB"
	Retriever                = "Retriever"
	Reranking                = "Reranking"
	TeiReranking             = "TeiReranking"
	Tgi                      = "Tgi"
	TgiGaudi                 = "TgiGaudi"
	Llm                      = "Llm"
	DocSum                   = "DocSum"
	Router                   = "router"
	DataPrep                 = "DataPrep"
	xeon                     = "xeon"
	gaudi                    = "gaudi"
	WebRetriever             = "WebRetriever"
	yaml_dir                 = "/tmp/microservices/yamls/"
	Service                  = "Service"
	Deployment               = "Deployment"
	dplymtSubfix             = "-deployment"
	METADATA_PLATFORM        = "gmc/platform"
	DefaultRouterServiceName = "router-service"
	ASR                      = "Asr"
	TTS                      = "Tts"
	SpeechT5                 = "SpeechT5"
	SpeechT5Gaudi            = "SpeechT5Gaudi"
	Whisper                  = "Whisper"
	WhisperGaudi             = "WhisperGaudi"
	gmcFinalizer             = "gmcFinalizer"
)

var yamlDict = map[string]string{
	TeiEmbedding:      yaml_dir + "tei.yaml",
	TeiEmbeddingGaudi: yaml_dir + "tei_gaudi.yaml",
	Embedding:         yaml_dir + "embedding-usvc.yaml",
	VectorDB:          yaml_dir + "redis-vector-db.yaml",
	Retriever:         yaml_dir + "retriever-usvc.yaml",
	Reranking:         yaml_dir + "reranking-usvc.yaml",
	TeiReranking:      yaml_dir + "teirerank.yaml",
	Tgi:               yaml_dir + "tgi.yaml",
	TgiGaudi:          yaml_dir + "tgi_gaudi.yaml",
	Llm:               yaml_dir + "llm-uservice.yaml",
	DocSum:            yaml_dir + "docsum-llm-uservice.yaml",
	Router:            yaml_dir + "gmc-router.yaml",
	WebRetriever:      yaml_dir + "web-retriever.yaml",
	ASR:               yaml_dir + "asr.yaml",
	TTS:               yaml_dir + "tts.yaml",
	SpeechT5:          yaml_dir + "speecht5.yaml",
	SpeechT5Gaudi:     yaml_dir + "speecht5_gaudi.yaml",
	Whisper:           yaml_dir + "whisper.yaml",
	WhisperGaudi:      yaml_dir + "whisper_gaudi.yaml",
	DataPrep:          yaml_dir + "data-prep.yaml",
}

// GMConnectorReconciler reconciles a GMConnector object
type GMConnectorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

type RouterCfg struct {
	Namespace   string
	SvcName     string
	DplymntName string
	NoProxy     string
	HttpProxy   string
	HttpsProxy  string
	GRAPH_JSON  string
}

func lookupManifestDir(step string) string {
	value, exist := yamlDict[step]
	if exist {
		return value
	} else {
		return ""
	}
}

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
				return nil, fmt.Errorf("failed to convert unstructured to service %s: %v", svc, err)
			}
			service_obj.SetName(svc)
			service_obj.Spec.Selector["app"] = svc
			err = scheme.Scheme.Convert(service_obj, obj, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to convert service %s to object: %v", svc, err)
			}
		} else if obj.GetKind() == Deployment {
			deployment_obj := &appsv1.Deployment{}
			err = scheme.Scheme.Convert(obj, deployment_obj, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to convert unstructured to deployment %s: %v", obj.GetName(), err)
			}
			if svc != "" {
				deployment_obj.SetName(svc + dplymtSubfix)
				// Set the labels if they're specified
				deployment_obj.Spec.Selector.MatchLabels["app"] = svc
				deployment_obj.Spec.Template.Labels["app"] = svc
			}

			// append the user defined ENVs
			var newEnvVars []corev1.EnvVar
			for name, value := range *svcCfg {
				if name == "endpoint" || name == "nodes" {
					continue
				}
				if isDownStreamEndpointKey(name) {
					ds := findDownStreamService(value, stepCfg, nodeCfg)
					value, err = getDownstreamSvcEndpoint(graphNs, value, ds)
					// value = getDsEndpoint(platform, name, graphNs, ds)
					if err != nil {
						return nil, fmt.Errorf("failed to find downstream service endpoint %s-%s: %v", name, value, err)
					}
				}
				itemEnvVar := corev1.EnvVar{
					Name:  name,
					Value: value,
				}
				newEnvVars = append(newEnvVars, itemEnvVar)
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
				return nil, fmt.Errorf("failed to convert deployment %s to obj: %v", deployment_obj.GetName(), err)
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

func isDownStreamEndpointKey(keyname string) bool {
	return keyname == "TEI_EMBEDDING_ENDPOINT" ||
		keyname == "TEI_RERANKING_ENDPOINT" ||
		keyname == "TGI_LLM_ENDPOINT" ||
		keyname == "REDIS_URL" ||
		keyname == "ASR_ENDPOINT" ||
		keyname == "TTS_ENDPOINT" ||
		keyname == "TEI_ENDPOINT"
}

func findDownStreamService(dsName string, stepCfg *mcv1alpha3.Step, nodeCfg *mcv1alpha3.Router) *mcv1alpha3.Step {
	if stepCfg == nil || nodeCfg == nil {
		return nil
	}
	fmt.Printf("find downstream service for %s with name %s \n", stepCfg.StepName, dsName)

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
	tmplt := lookupManifestDir(stepCfg.StepName)
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

		if stepCfg.StepName == VectorDB {
			return fmt.Sprintf("redis://%s.%s.svc.cluster.local:%d", altSvcName, altNs, port), nil
		} else {
			return fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", altSvcName, altNs, port), nil
		}
	} else {
		return "", errors.New(fmt.Sprintf("failed to get service details for %s: %v\n", dsName, err))
	}
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

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) []string {
	var result []string
	for _, item := range slice {
		if item != s {
			result = append(result, item)
		}
	}
	return result
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

	// Check if the GMConnector instance is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	if graph.GetDeletionTimestamp() != nil {
		if len(graph.GetFinalizers()) != 0 {
			// Run finalization logic for gmConnectorFinalizer. If the
			// finalization logic fails, don't remove the finalizer so
			// that we can retry during the next reconciliation.
			if err := r.finalizeGMConnector(ctx, graph); err != nil {
				return ctrl.Result{}, err
			}
		}

		// Remove finalizer
		graph.SetFinalizers(removeString(graph.GetFinalizers(), gmcFinalizer))
		if err := r.Update(ctx, graph); err != nil {
			return ctrl.Result{}, err
		}
		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	// Add finalizer for this CR if not already present
	if !containsString(graph.GetFinalizers(), gmcFinalizer) {
		graph.SetFinalizers(append(graph.GetFinalizers(), gmcFinalizer))
		if err := r.Update(ctx, graph); err != nil {
			return ctrl.Result{}, err
		}
		// Return and requeue after adding the finalizer
		return ctrl.Result{Requeue: true}, nil
	}

	var totalService uint
	var externalService uint
	var successService uint
	var updateExistGraph bool = false
	var oldAnnotations map[string]string

	if len(graph.Status.Annotations) == 0 {
		graph.Status.Annotations = make(map[string]string)
	} else {
		updateExistGraph = true
		//save the old annotations
		oldAnnotations = graph.Status.Annotations
		graph.Status.Annotations = make(map[string]string)
	}

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

				objs, err := reconcileResource(ctx, r.Client, graph.Namespace, &step, &node)
				if err != nil {
					return reconcile.Result{Requeue: true}, errors.Wrapf(err, "Failed to reconcile service for %s", step.StepName)
				}
				if len(objs) != 0 {
					for _, obj := range objs {
						success, err := recordResourceStatus(graph, &step, obj)
						if err != nil {
							return reconcile.Result{Requeue: true}, errors.Wrapf(err, "Resource created with failure %s", step.StepName)
						}
						successService += success
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
	totalService += 1
	err := reconcileRouterService(ctx, r.Client, graph)
	if err != nil {
		return reconcile.Result{Requeue: true}, errors.Wrapf(err, "Failed to reconcile router service")
	} else {
		successService += 1
	}

	graph.Status.Status = fmt.Sprintf("%d/%d/%d", successService, externalService, totalService)
	if err = r.Status().Update(ctx, graph); err != nil {
		return reconcile.Result{Requeue: true}, errors.Wrapf(err, "Failed to Update CR status to %s", graph.Status.Status)
	}

	if updateExistGraph {
		//check if the old annotations are still in the new graph
		for k := range oldAnnotations {
			if _, ok := graph.Status.Annotations[k]; !ok {
				//if not, remove the resource from k8s
				kind := strings.Split(k, ":")[0]
				name := strings.Split(k, ":")[1]
				ns := strings.Split(k, ":")[2]
				fmt.Printf("delete resource %s %s %s\n", kind, name, ns)
				obj := &unstructured.Unstructured{}
				obj.SetKind(kind)
				obj.SetName(name)
				obj.SetNamespace(ns)
				err := r.Delete(ctx, obj)
				if err != nil {
					return reconcile.Result{Requeue: true}, errors.Wrapf(err, "Failed to delete resource %s", name)
				} else {
					fmt.Printf("Success to delete %s: %s\n", kind, name)
				}
			}
		}
	}

	return ctrl.Result{}, nil
}

// finalizeGMConnector contains the logic to clean up resources before the CR is deleted
func (r *GMConnectorReconciler) finalizeGMConnector(ctx context.Context, graph *mcv1alpha3.GMConnector) error {
	// for compatibility consideration, the old GMC didn't save annotation
	// so skip this part let the old graph deleted by k8s
	// or it will be stuck
	if len(graph.Status.Annotations) == 0 {
		fmt.Printf("skip resource deletion due to no record\n")
		return nil
	}
	//check if the old annotations are still in the new graph
	for k := range graph.Status.Annotations {
		//if not, remove the resource from k8s
		kind := strings.Split(k, ":")[0]
		apiVersion := strings.Split(k, ":")[1]
		name := strings.Split(k, ":")[2]
		ns := strings.Split(k, ":")[3]
		fmt.Printf("delete resource %s %s %s\n", kind, name, ns)
		obj := &unstructured.Unstructured{}
		obj.SetKind(kind)
		obj.SetName(name)
		obj.SetNamespace(ns)
		obj.SetAPIVersion(apiVersion)

		// // Fetch the resource to get its full metadata
		// err := r.Get(ctx, client.ObjectKey{Name: name, Namespace: ns}, obj)
		// if err != nil {
		// 	return errors.Wrapf(err, "Failed to fetch resource %s %s %s\n", kind, name, ns)
		// }

		err := r.Delete(ctx, obj)
		if err != nil {
			return errors.Wrapf(err, "Failed to delete resource %s %s %s\n", kind, name, ns)
		} else {
			fmt.Printf("Success to delete %s %s %s\n", kind, name, ns)
		}
	}
	return nil
}

func recordResourceStatus(graph *mcv1alpha3.GMConnector, step *mcv1alpha3.Step, obj *unstructured.Unstructured) (uint, error) {
	// var statusStr string
	var success uint = 0
	// graph.SetFinalizers(append(graph.GetFinalizers(), fmt.Sprintf("%s-.-%s-.-%s", obj.GetKind(), obj.GetNamespace(), obj.GetName())))
	// save the resource name into annotation for status update and resource management
	graph.Status.Annotations[fmt.Sprintf("%s:%s:%s:%s", obj.GetKind(), obj.GetAPIVersion(), obj.GetName(), obj.GetNamespace())] = "provisioned"

	if obj.GetKind() == Service {
		service := &corev1.Service{}
		err := scheme.Scheme.Convert(obj, service, nil)
		if err != nil {
			return success, errors.Wrapf(err, "Failed to convert service %s", obj.GetName())
		}

		if step != nil {
			url := getServiceURL(service) + step.InternalService.Config["endpoint"]
			//set this for router
			step.ServiceURL = url
			graph.Status.Annotations[fmt.Sprintf("%s:%s:%s:%s", obj.GetKind(), obj.GetAPIVersion(), obj.GetName(), obj.GetNamespace())] = url
			fmt.Printf("the service URL is: %s\n", url)
		} else {
			url := getServiceURL(service)
			graph.Status.Annotations[fmt.Sprintf("%s:%s:%s:%s", obj.GetKind(), obj.GetAPIVersion(), obj.GetName(), obj.GetNamespace())] = url
			fmt.Printf("the router URL is: %s\n", url)
		}
	}
	if obj.GetKind() == Deployment {
		deployment := &appsv1.Deployment{}
		err := scheme.Scheme.Convert(obj, deployment, nil)
		if err != nil {
			return success, errors.Wrapf(err, "Failed to convert deployment %s", obj.GetName())
		}
		if len(deployment.Status.Conditions) > 0 {
			graph.Status.Annotations[fmt.Sprintf("%s:%s:%s:%s", obj.GetKind(), obj.GetAPIVersion(), obj.GetName(), obj.GetNamespace())] =
				fmt.Sprintf("Status:%s\n\tReason:%s\n\tMessage%s", deployment.Status.Conditions[len(deployment.Status.Conditions)-1].Status,
					deployment.Status.Conditions[len(deployment.Status.Conditions)-1].Reason,
					deployment.Status.Conditions[len(deployment.Status.Conditions)-1].Message)
			// statusStr = deployment.Status.Conditions[len(deployment.Status.Conditions)-1].Type
			if deployment.Status.Conditions[len(deployment.Status.Conditions)-1].Type == appsv1.DeploymentAvailable {
				success += 1
			}
		}

	}

	return success, nil
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
	configForRouter := make(map[string]string)
	for k, v := range graph.Spec.RouterConfig.Config {
		configForRouter[k] = v
	}
	var routerNs string
	var routerServiceName string
	var routerDeploymentName string
	jsonBytes, err := json.Marshal(graph.Spec)
	if err != nil {
		// handle error
		return errors.Wrapf(err, "Failed to Marshal routes for %s", graph.Spec.RouterConfig.Name)
	}
	jsonString := string(jsonBytes)
	escapedString := strings.ReplaceAll(jsonString, "'", "\\'")
	configForRouter["nodes"] = "'" + escapedString + "'"

	if graph.Spec.RouterConfig.NameSpace != "" {
		routerNs = graph.Spec.RouterConfig.NameSpace
	} else {
		routerNs = graph.Namespace
	}
	configForRouter["namespace"] = routerNs

	if graph.Spec.RouterConfig.ServiceName != "" {
		routerServiceName = graph.Spec.RouterConfig.ServiceName
		routerDeploymentName = graph.Spec.RouterConfig.ServiceName + dplymtSubfix
	} else {
		routerServiceName = DefaultRouterServiceName
		routerDeploymentName = DefaultRouterServiceName + dplymtSubfix
	}
	configForRouter["svcName"] = routerServiceName
	configForRouter["dplymntName"] = routerDeploymentName

	templateBytes, err := getTemplateBytes(Router)
	if err != nil {
		return errors.Wrapf(err, "Failed to get template bytes for %s", Router)
	}
	var resources []string
	appliedCfg, err := applyRouterConfigToTemplates(Router, &configForRouter, templateBytes)
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

		err = applyResourceToK8s(ctx, client, obj)
		if err != nil {
			return fmt.Errorf("failed to reconcile resource: %v", err)
		} else {
			fmt.Printf("Success to reconcile %s: %s\n", obj.GetKind(), obj.GetName())
		}
		// graph.SetFinalizers(append(graph.GetFinalizers(), fmt.Sprintf("%s-.-%s-.-%s", obj.GetKind(), obj.GetNamespace(), obj.GetName())))
		// save the resource name into annotation for status update and resource management
		_, err = recordResourceStatus(graph, nil, obj)
		if err != nil {
			return fmt.Errorf("resource created with failure %s: %v", obj.GetName(), err)
		}
	}

	return nil
}

func applyRouterConfigToTemplates(step string, svcCfg *map[string]string, yamlFile []byte) (string, error) {
	var userDefinedCfg RouterCfg
	if step == "router" {
		userDefinedCfg = RouterCfg{
			Namespace:   (*svcCfg)["namespace"],
			SvcName:     (*svcCfg)["svcName"],
			DplymntName: (*svcCfg)["dplymntName"],
			NoProxy:     (*svcCfg)["no_proxy"],
			HttpProxy:   (*svcCfg)["http_proxy"],
			HttpsProxy:  (*svcCfg)["https_proxy"],
			GRAPH_JSON:  (*svcCfg)["nodes"]}
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

func isMetadataChanged(oldObject, newObject *metav1.ObjectMeta) bool {
	if oldObject == nil || newObject == nil {
		fmt.Printf("Metadata changes detected, old/new object is nil\n")
		return oldObject != newObject
	}
	// only care limited changes
	if oldObject.Name != newObject.Name {
		fmt.Printf("Metadata changes detected, Name changed from %s to %s\n", oldObject.Name, newObject.Name)
		return true
	}
	if oldObject.Namespace != newObject.Namespace {
		fmt.Printf("Metadata changes detected, Namespace changed from %s to %s\n", oldObject.Namespace, newObject.Namespace)
		return true
	}
	if !reflect.DeepEqual(oldObject.Labels, newObject.Labels) {
		fmt.Printf("Metadata changes detected, Labels changed from %v to %v\n", oldObject.Labels, newObject.Labels)
		return true
	}
	if !reflect.DeepEqual(oldObject.DeletionTimestamp, newObject.DeletionTimestamp) {
		fmt.Printf("Metadata changes detected, DeletionTimestamp changed from %v to %v\n", oldObject.DeletionTimestamp, newObject.DeletionTimestamp)
		return true
	}
	// Add more fields as needed
	return false
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
			metadataChanged := isMetadataChanged(&(oldObject.ObjectMeta), &(newObject.ObjectMeta))

			fmt.Printf("\nspec changed %t | meta changed: %t\n", specChanged, metadataChanged)

			// Compare the old and new spec, ignore metadata, status changes
			// metadata change: name, namespace, such change should create a new GMC
			// status change: depoyment status
			return specChanged || metadataChanged
		},
		// Other funcs like CreateFunc, DeleteFunc, GenericFunc can be left as default
		// if you only want to customize the UpdateFunc behavior.
	}

	// Setup the watch with the predicate to filter events
	return ctrl.NewControllerManagedBy(mgr).
		For(&mcv1alpha3.GMConnector{}).
		WithEventFilter(ignoreStatusUpdatePredicate). // Use the predicate here
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
