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
	"strconv"
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
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
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
	TgiNvidia                = "TgiNvidia"
	Llm                      = "Llm"
	DocSum                   = "DocSum"
	Router                   = "router"
	DataPrep                 = "DataPrep"
	xeon                     = "xeon"
	gaudi                    = "gaudi"
	nvidia                   = "nvidia"
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
	UI                       = "UI"
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
	TgiNvidia:         yaml_dir + "tgi_nv.yaml",
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
	UI:                yaml_dir + "ui.yaml",
}

var (
	_log = ctrl.Log.WithName("GMC")
)

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

func (r *GMConnectorReconciler) reconcileResource(ctx context.Context, graphNs string, stepCfg *mcv1alpha3.Step, nodeCfg *mcv1alpha3.Router, graph *mcv1alpha3.GMConnector) ([]*unstructured.Unstructured, error) {
	if stepCfg == nil || nodeCfg == nil {
		return nil, errors.New("invalid svc config")
	}

	_log.V(1).Info("processing step", "config", *stepCfg)

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
		_log.Error(err, "Failed to get template bytes for", "step", stepCfg.StepName)
		return nil, err
	}

	resources := strings.Split(string(yamlFile), "---")
	for _, res := range resources {
		if res == "" || !strings.Contains(res, "kind:") {
			continue
		}

		decUnstructured := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		obj := &unstructured.Unstructured{}
		_, _, err := decUnstructured.Decode([]byte(res), nil, obj)
		if err != nil {
			_log.Error(err, "Failed to decode YAML")
			return nil, err
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
				_log.Error(err, "Failed to convert unstructured to service", "name", svc)
				return nil, err
			}
			service_obj.SetName(svc)
			service_obj.Spec.Selector["app"] = svc
			err = scheme.Scheme.Convert(service_obj, obj, nil)
			if err != nil {
				_log.Error(err, "Failed to convert service to object", "name", svc)
				return nil, err
			}
		} else if obj.GetKind() == Deployment {
			deployment_obj := &appsv1.Deployment{}
			err = scheme.Scheme.Convert(obj, deployment_obj, nil)
			if err != nil {
				_log.Error(err, "Failed to convert unstructured to deployment", "name", obj.GetName())
				return nil, err
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
					if err != nil {
						_log.Error(err, "Failed to find downstream service endpoint", "name", name, "value", value)
						return nil, err
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
				_log.Error(err, "Failed to convert deployment to obj", "name", deployment_obj.GetName())
				return nil, err
			}
		}

		err = r.applyResourceToK8s(graph, ctx, obj)
		if err != nil {
			_log.Error(err, "Failed to reconcile resource", "name", obj.GetName())
			return nil, err
		} else {
			_log.Info("Success to reconcile resource", "kind", obj.GetKind(), "name", obj.GetName())
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
	_log.Info("Find downstream service for step", "name", stepCfg.StepName, "downstream", dsName)

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

// +kubebuilder:rbac:groups=gmc.opea.io,resources=gmconnectors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=gmc.opea.io,resources=gmconnectors/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=gmc.opea.io,resources=gmconnectors/finalizers,verbs=update
// +kubebuilder:rbac:groups=gmc.opea.io,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=gmc.opea.io,resources=deployments/status,verbs=get
// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// the GMConnector object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *GMConnectorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// _ = log.FromContext(ctx)
	_log.Info("----RECONCILE REQUEST----", "req", req)

	graph := &mcv1alpha3.GMConnector{}
	if err := r.Get(ctx, req.NamespacedName, graph); err != nil {
		if apierr.IsNotFound(err) {
			// Object not found, could be deployments
			deployment := &appsv1.Deployment{}
			err := r.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: req.Name}, deployment)
			if err == nil {
				return r.handleStatusUpdate(ctx, deployment)
			} else {
				_log.Info("Resource not found or deleted, ignore", "name", req.Name, "err", err)
				return ctrl.Result{}, nil
			}
		} else {
			return reconcile.Result{}, errors.Wrapf(err, "Failed to get GMConnector %s", req.Name)
		}
	}

	// in case the type meta is not set, in ut
	// check if typemeta is empty
	if reflect.DeepEqual(graph.TypeMeta, metav1.TypeMeta{}) {
		graph.TypeMeta = metav1.TypeMeta{
			APIVersion: "gmc.opea.io/v1alpha3",
			Kind:       "GMConnector",
		}
	}

	var totalService uint
	var externalService uint
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
				_log.Info("This is a nested step", "step", step.StepName)
				continue
			}
			_log.Info("Reconcile step", "graph", graph.Name, "name", step.StepName)
			totalService += 1
			if step.Executor.ExternalService == "" {
				_log.Info("Trying to reconcile internal service", " service", step.Executor.InternalService.ServiceName)

				objs, err := r.reconcileResource(ctx, graph.Namespace, &step, &node, graph)
				if err != nil {
					return reconcile.Result{Requeue: true}, errors.Wrapf(err, "Failed to reconcile service for %s", step.StepName)
				}
				if len(objs) != 0 {
					for _, obj := range objs {
						err := recordResource(graph, nodeName, i, obj)
						if err != nil {
							return reconcile.Result{Requeue: true}, errors.Wrapf(err, "Resource created with failure %s", step.StepName)
						}
					}
				}
			} else {
				_log.Info("External service is found", "name", step.ExternalService)
				graph.Spec.Nodes[nodeName].Steps[i].ServiceURL = step.ExternalService
				externalService += 1
			}
		}
	}

	//to start a router service
	//in case the graph changes, we need to apply the changes to router service
	//so we need to apply the router config every time
	err := r.reconcileRouterService(ctx, graph)
	if err != nil {
		return reconcile.Result{Requeue: true}, errors.Wrapf(err, "Failed to reconcile router service")
	}

	if updateExistGraph {
		//check if the old annotations are still in the new graph
		for k := range oldAnnotations {
			if _, ok := graph.Status.Annotations[k]; !ok {
				//if not, remove the resource from k8s
				r.deleteRecordedResource(k, ctx)
			}
		}
	}

	graph.Status.Status = fmt.Sprintf("%d/%d/%d", 0, externalService, 0)
	err = r.collectResourceStatus(graph, ctx)
	if err != nil {
		return reconcile.Result{Requeue: true}, errors.Wrapf(err, "Failed to collect service status")
	}

	return ctrl.Result{}, nil
}

func (r *GMConnectorReconciler) handleStatusUpdate(ctx context.Context, deployment *appsv1.Deployment) (ctrl.Result, error) {
	for _, owner := range deployment.OwnerReferences {
		if owner.Kind == "GMConnector" {
			// Get the GMConnector object
			graph := &mcv1alpha3.GMConnector{}
			err := r.Get(ctx, types.NamespacedName{Namespace: deployment.Namespace, Name: owner.Name}, graph)
			if err == nil {
				ue := r.collectResourceStatus(graph, ctx)
				if ue != nil {
					_log.Error(err, "Failed to get graph before update status", "name", graph.Name)
					return reconcile.Result{}, err
				}
			}
		}
	}
	return ctrl.Result{}, nil
}

func (r *GMConnectorReconciler) deleteRecordedResource(key string, ctx context.Context) {
	kind := strings.Split(key, ":")[0]
	apiVersion := strings.Split(key, ":")[1]
	name := strings.Split(key, ":")[2]
	ns := strings.Split(key, ":")[3]
	obj := &unstructured.Unstructured{}
	obj.SetKind(kind)
	obj.SetName(name)
	obj.SetNamespace(ns)
	obj.SetAPIVersion(apiVersion)
	err := r.Delete(ctx, obj)
	// the resource may have been deleted by other means, i.e. user manually delete or delete namespace
	// ignore the error if delete failed i.e resource not found
	// since I don't want to block the process for not clearing the finalizer
	if err != nil {
		_log.Info("Failed to delete resource", "namespace", ns, "kind", kind, "name", name, "error", err)
	} else {
		_log.Info("Success to delete resource", "namespace", ns, "kind", kind, "name", name)
	}
}

func (r *GMConnectorReconciler) collectResourceStatus(graph *mcv1alpha3.GMConnector, ctx context.Context) error {
	if graph == nil || len(graph.Status.Annotations) == 0 {
		return errors.New("graph is empty or no annotations")
	}
	var totalCnt uint = 0
	var readyCnt uint = 0
	for resName := range graph.Status.Annotations {
		kind := strings.Split(resName, ":")[0]
		name := strings.Split(resName, ":")[2]
		ns := strings.Split(resName, ":")[3]

		if kind == Deployment {
			totalCnt += 1

			deployment := &appsv1.Deployment{}
			err := r.Get(ctx, client.ObjectKey{Namespace: ns, Name: name}, deployment)
			if err != nil {
				_log.Info("Collecting status: failed to get deployment", "name", name, "error", err)
				continue
			}
			var deploymentStatus strings.Builder
			deploymentStatus.WriteString(fmt.Sprintf("Replicas: %d desired | %d updated | %d total | %d available | %d unavailable\n",
				*deployment.Spec.Replicas,
				deployment.Status.UpdatedReplicas,
				deployment.Status.Replicas,
				deployment.Status.AvailableReplicas,
				deployment.Status.UnavailableReplicas))
			deploymentStatus.WriteString("Conditions:\n")
			for _, condition := range deployment.Status.Conditions {
				deploymentStatus.WriteString(fmt.Sprintf("  Type: %s\n", condition.Type))
				deploymentStatus.WriteString(fmt.Sprintf("  Status: %s\n", condition.Status))
				deploymentStatus.WriteString(fmt.Sprintf("  Reason: %s\n", condition.Reason))
				deploymentStatus.WriteString(fmt.Sprintf("  Message: %s\n", condition.Message))
			}
			graph.Status.Annotations[resName] = deploymentStatus.String()
			if deployment.Status.AvailableReplicas == *deployment.Spec.Replicas {
				readyCnt += 1
			}
		}
	}
	externalResourceCntStr := strings.Split(graph.Status.Status, "/")[1]
	externalResourceCnt, err := strconv.Atoi(externalResourceCntStr)
	if err != nil {
		return errors.Wrapf(err, "Error converting externalResourceCnt to int")
	}
	graph.Status.Status = fmt.Sprintf("%d/%d/%d", readyCnt, externalResourceCnt, totalCnt)

	//update the revision in case it has changed
	var latestGraph mcv1alpha3.GMConnector
	err = r.Client.Get(ctx, types.NamespacedName{Namespace: graph.Namespace, Name: graph.Name}, &latestGraph)
	if err != nil && apierr.IsNotFound(err) {
		_log.Info("Failed to get graph before update status", "name", graph.Name, "error", err)
	} else {
		graph.SetResourceVersion(latestGraph.GetResourceVersion())
	}

	if err = r.Status().Update(ctx, graph); err != nil {
		return errors.Wrapf(err, "Failed to Update CR status to %s", graph.Status.Status)
	}

	return nil
}

func recordResource(graph *mcv1alpha3.GMConnector, nodeName string, stepIdx int, obj *unstructured.Unstructured) error {
	// save the resource name into annotation for status update and resource management
	graph.Status.Annotations[fmt.Sprintf("%s:%s:%s:%s", obj.GetKind(), obj.GetAPIVersion(), obj.GetName(), obj.GetNamespace())] = "provisioned"

	if obj.GetKind() == Service {
		service := &corev1.Service{}
		err := scheme.Scheme.Convert(obj, service, nil)
		if err != nil {
			return errors.Wrapf(err, "Failed to convert service %s", obj.GetName())
		}

		if len(graph.Spec.Nodes) != 0 && len(graph.Spec.Nodes[nodeName].Steps) != 0 {
			url := getServiceURL(service) + graph.Spec.Nodes[nodeName].Steps[stepIdx].InternalService.Config["endpoint"]
			//set this for router
			graph.Spec.Nodes[nodeName].Steps[stepIdx].ServiceURL = url
			graph.Status.Annotations[fmt.Sprintf("%s:%s:%s:%s", obj.GetKind(), obj.GetAPIVersion(), obj.GetName(), obj.GetNamespace())] = url
			_log.Info("Service URL is: ", "URL", url)
		} else {
			url := getServiceURL(service)
			graph.Status.Annotations[fmt.Sprintf("%s:%s:%s:%s", obj.GetKind(), obj.GetAPIVersion(), obj.GetName(), obj.GetNamespace())] = url
			graph.Status.AccessURL = url
			_log.Info("Router URL is: ", "URL", url)
		}
	}
	return nil
}

func getTemplateBytes(resourceType string) ([]byte, error) {
	tmpltFile := lookupManifestDir(resourceType)
	if tmpltFile == "" {
		return nil, errors.New("unexpected target")
	}
	yamlBytes, err := os.ReadFile(tmpltFile)
	if err != nil {
		return nil, err
	}
	return yamlBytes, nil
}

func (r *GMConnectorReconciler) reconcileRouterService(ctx context.Context, graph *mcv1alpha3.GMConnector) error {
	configForRouter := make(map[string]string)

	var routerNs string
	var routerServiceName string
	var routerDeploymentName string
	var graphJson mcv1alpha3.GMConnector
	graphJson.Spec = *graph.Spec.DeepCopy()

	jsonBytes, err := json.Marshal(graphJson)
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
		_log.Error(err, "Failed to apply user config")
		return err
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
			_log.Error(err, "Failed to decode YAML")
			return err
		}

		err = r.applyResourceToK8s(graph, ctx, obj)
		if err != nil {
			_log.Error(err, "Failed to reconcile resource", "name", obj.GetName())
			return err
		} else {
			_log.Info("Success to reconcile resource", "kind", obj.GetKind(), "name", obj.GetName())
		}
		// save the resource name into annotation for status update and resource management
		err = recordResource(graph, "", 0, obj)
		if err != nil {
			_log.Error(err, "Resource created with failure", "name", obj.GetName())
			return err
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
		_log.V(1).Info("Apply the config to router", "content", userDefinedCfg)

		tmpl, err := template.New("yamlTemplate").Parse(string(yamlFile))
		if err != nil {
			return string(yamlFile), fmt.Errorf("error parsing template: %v", err)
		}

		var appliedCfg bytes.Buffer
		err = tmpl.Execute(&appliedCfg, userDefinedCfg)
		if err != nil {
			return string(yamlFile), fmt.Errorf("error executing template: %v", err)
		} else {
			_log.V(1).Info("applied config", "content", appliedCfg.String())
			return appliedCfg.String(), nil
		}
	} else {
		return string(yamlFile), nil
	}

}

func (r *GMConnectorReconciler) applyResourceToK8s(graph *mcv1alpha3.GMConnector, ctx context.Context, obj *unstructured.Unstructured) error {
	// Prepare the object for an update, assuming it already exists. If it doesn't, you'll need to handle that case.
	// This might involve trying an Update and, if it fails because the object doesn't exist, falling back to Create.
	// Retry updating the resource in case of transient errors.
	timeout := time.After(1 * time.Minute)
	tick := time.NewTicker(100 * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled")
		case <-timeout:
			return fmt.Errorf("timed out while trying to update or create resource")
		case <-tick.C:
			if err := controllerutil.SetControllerReference(graph, obj, r.Scheme); err != nil {
				return fmt.Errorf("failed to set controller reference: %v", err)
			}
			// Get the latest version of the object
			latest := &unstructured.Unstructured{}
			latest.SetGroupVersionKind(obj.GroupVersionKind())
			err := r.Client.Get(ctx, client.ObjectKeyFromObject(obj), latest)
			if err != nil {
				if apierr.IsNotFound(err) {
					// If the object doesn't exist, create it
					err = r.Client.Create(ctx, obj, &client.CreateOptions{})
					if err != nil {
						return fmt.Errorf("failed to create resource: %v", err)
					}
				} else {
					// If there was another error, continue
					_log.Info("Get object err", "message", err)
					continue
				}
			} else {
				// If the object does exist, update it
				obj.SetResourceVersion(latest.GetResourceVersion()) // Ensure we're updating the latest version
				err = r.Client.Update(ctx, obj, &client.UpdateOptions{})
				if err != nil {
					_log.Info("Update object err", "message", err)
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
		_log.Info("Metadata changes detected, old/new object is nil")
		return oldObject != newObject
	}
	// only care limited changes
	if oldObject.Name != newObject.Name {
		_log.Info("Metadata.Name changes detected", "old", oldObject.Name, "new", newObject.Name)
		return true
	}
	if oldObject.Namespace != newObject.Namespace {
		_log.Info("Metadata.Namespace changes detected", "old", oldObject.Namespace, "new", newObject.Namespace)
		return true
	}
	if !reflect.DeepEqual(oldObject.Labels, newObject.Labels) {
		_log.Info("Metadata.Labels changes detected", "old", oldObject.Labels, "new", newObject.Labels)
		return true
	}
	if !reflect.DeepEqual(oldObject.DeletionTimestamp, newObject.DeletionTimestamp) {
		_log.Info("Metadata.DeletionTimestamp changes detected", "old", oldObject.DeletionTimestamp, "new", newObject.DeletionTimestamp)
		return true
	}
	// Add more fields as needed
	return false
}

func isGMCSpecOrMetadataChanged(e event.UpdateEvent) bool {
	oldObject, ok1 := e.ObjectOld.(*mcv1alpha3.GMConnector)
	newObject, ok2 := e.ObjectNew.(*mcv1alpha3.GMConnector)
	if !ok1 || !ok2 {
		// Not the correct type, allow the event through
		return true
	}

	specChanged := !reflect.DeepEqual(oldObject.Spec, newObject.Spec)
	metadataChanged := isMetadataChanged(&(oldObject.ObjectMeta), &(newObject.ObjectMeta))

	_log.V(1).Info("Check trigger condition?", "spec changed", specChanged, "meta changed", metadataChanged)
	// Compare the old and new spec, ignore metadata, status changes
	// metadata change: name, namespace, such change should create a new GMC
	// status change: deployment status
	return specChanged || metadataChanged
}

func isDeploymentStatusChanged(e event.UpdateEvent) bool {
	oldDeployment, ok1 := e.ObjectOld.(*appsv1.Deployment)
	newDeployment, ok2 := e.ObjectNew.(*appsv1.Deployment)
	if !ok1 || !ok2 {
		// Not the correct type, allow the event through
		return true
	}

	if len(newDeployment.OwnerReferences) == 0 {
		_log.V(1).Info("No owner reference", "ns", newDeployment.Namespace, "name", newDeployment.Name)
		return false
	} else {
		for _, owner := range newDeployment.OwnerReferences {
			if owner.Kind == "GMConnector" {
				_log.V(1).Info("Owner is GMConnector", "ns", newDeployment.Namespace, "name", newDeployment.Name)
				break
			}
		}
	}

	oldStatus := corev1.ConditionUnknown
	newStatus := corev1.ConditionUnknown
	if !reflect.DeepEqual(oldDeployment.Status, newDeployment.Status) {
		if newDeployment.Status.Conditions != nil {
			for _, condition := range newDeployment.Status.Conditions {
				if condition.Type == appsv1.DeploymentAvailable {
					newStatus = condition.Status
				}
			}
		}
		if oldDeployment.Status.Conditions != nil {
			for _, condition := range oldDeployment.Status.Conditions {
				if condition.Type == appsv1.DeploymentAvailable {
					oldStatus = condition.Status
				}
			}
		}
		// Only trigger if the status has changed from true to false|unknown or vice versa
		if (oldStatus == corev1.ConditionTrue && oldStatus != newStatus) ||
			(newStatus == corev1.ConditionTrue && oldStatus != newStatus) {
			{
				_log.Info("status changed", "ns",
					newDeployment.Namespace, "name", newDeployment.Name,
					"from", oldStatus, "to", newStatus)
				return true
			}
		}
	}
	return false

}

// SetupWithManager sets up the controller with the Manager.
func (r *GMConnectorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Predicate to ignore updates to status subresource
	gmcfilter := predicate.Funcs{
		UpdateFunc: isGMCSpecOrMetadataChanged,
		// Other funcs like CreateFunc, DeleteFunc, GenericFunc can be left as default
		// if you only want to customize the UpdateFunc behavior.
	}

	// Predicate to only trigger on status changes for Deployment
	deploymentFilter := predicate.Funcs{
		UpdateFunc: isDeploymentStatusChanged,
		//ignore create and delete events, otherwise it will trigger the nested reconcile which is meaningless
		CreateFunc: func(e event.CreateEvent) bool {
			return false
		}, DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&mcv1alpha3.GMConnector{}, builder.WithPredicates(gmcfilter)).
		Watches(
			&appsv1.Deployment{},
			&handler.EnqueueRequestForObject{},
			builder.WithPredicates(deploymentFilter),
		).
		Complete(r)
}
