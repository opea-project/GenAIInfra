/*
* Copyright (C) 2024 Intel Corporation
* SPDX-License-Identifier: Apache-2.0
 */

/* Modifications made to this file by [Intel] on [2024]
*  Portions of this file are derived from kserve: https://github.com/kserve/kserve
*  Copyright 2022 The KServe Author
 */

package v1alpha3

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GMCTarget represents the structure to hold either a PredefinedType or a ServiceReference.
type GMCTarget struct {
	ServiceName string `json:"serviceName,omitempty"`

	// +optional
	NameSpace string `json:"nameSpace,omitempty"`

	// +optional
	Config map[string]string `json:"config,omitempty"`

	// in the OPEA context, some service can automatically trigger another one
	// if this field is not empty, means the downstream service will be invoked
	// +optional
	IsDownstreamService bool `json:"isDownstreamService,omitempty"`
}

// StepDependencyType constant for step dependency
// +k8s:openapi-gen=true
// +kubebuilder:validation:Enum=Soft;Hard
type StepDependencyType string

// StepDependency Enum
const (
	// Soft
	Soft StepDependencyType = "Soft"

	// Hard
	Hard StepDependencyType = "Hard"
)

// StepNameType constant for step
// +k8s:openapi-gen=true
// +kubebuilder:validation:Enum=Soft;Hard
type StepNameType string

// StepDependency Enum
const (
	// Emebdding
	Embedding StepNameType = "Embedding"
	// Tei-Embedding
	TeiEmbedding StepNameType = "TeiEmbedding"
	// VectorDB
	VectorDB StepNameType = "VectorDB"
	// Retriever
	Retriever StepNameType = "Retriever"
	// Reranking
	Reranking StepNameType = "Reranking"
	// Tei-Reranking
	TeiReranking StepNameType = "TeiReranking"
	// Tgi
	Tgi StepNameType = "Tgi"
	// Llm
	Llm StepNameType = "Llm"
	// LLMGuardInput
	LLMGuardInput StepNameType = "LLMGuardInput"
	// LLMGuardOutput
	LLMGuardOutput StepNameType = "LLMGuardOutput"
	// VLLMGaudi
	VLLMGaudi StepNameType = "VLLMGaudi"
	// VLLM
	VLLM StepNameType = "VLLM"
	// VLLMOpenVino
	VLLMOpenVino StepNameType = "VLLMOpenVino"
	// Language-Detection
	LanguageDetection StepNameType = "LanguageDetection"
)

type Executor struct {
	// The node name for routing as the next step.
	// +optional
	NodeName string `json:"nodeName,omitempty"`
	// InternalService URL, mutually exclusive with ExternalService.
	// +optional
	InternalService GMCTarget `json:"internalService,omitempty"`
	// ExternalService URL, mutually exclusive with InternalService.
	// +optional
	ExternalService string `json:"externalService,omitempty"`
}

// Step defines the target of the current step with condition, weights and data.
// +k8s:openapi-gen=true
type Step struct {
	// Unique name for the step within this node
	StepName string `json:"name,omitempty"`

	// Node or service used to process this step
	Executor `json:",inline"`

	// request data sent to the next route with input/output from the previous step
	// $request
	// $response.predictions
	// +optional
	Data string `json:"data,omitempty"`

	// routing based on the condition
	// +optional
	Condition string `json:"condition,omitempty"`

	// to decide whether a step is a hard or a soft dependency in the Graph
	// +optional
	Dependency StepDependencyType `json:"dependency,omitempty"`

	// this is not for the users to set
	// when the service is ready, save the URL here for router to call
	// +optional
	ServiceURL string `json:"serviceUrl,omitempty"`
}

// RouterType constant for routing types
// +k8s:openapi-gen=true
// +kubebuilder:validation:Enum=Sequence;Ensemble;Switch
type RouterType string

// GMCRouterType Enum
const (
	// Sequence Default type only route to subsequent destination
	Sequence RouterType = "Sequence"

	// Ensemble router routes the requests to multiple destinations and then merge the responses
	Ensemble RouterType = "Ensemble"

	// Switch routes the request to the destination based on certain condition
	Switch RouterType = "Switch"
)

type Router struct {
	// RouterType
	//
	// - `Sequence:` chain multiple steps with input/output from previous step
	//
	// - `Ensemble:` routes the request to multiple destinations and then merge the responses
	//
	// - `Switch:` routes the request to one of the steps based on condition
	//
	RouterType RouterType `json:"routerType"`

	// Steps defines destinations for the current router node
	// +optional
	Steps []Step `json:"steps,omitempty"`
}

type RouterConfig struct {
	Name        string `json:"name"`
	ServiceName string `json:"serviceName"`
	// +optional
	NameSpace string `json:"nameSpace"`
	// +optional
	Config map[string]string `json:"config"`
}

// GMConnectorSpec defines the desired state of GMConnector
type GMConnectorSpec struct {
	Nodes        map[string]Router `json:"nodes"`
	RouterConfig RouterConfig      `json:"routerConfig"`
}

type ConditionType string

// Well-known condition types for GMConnector status.
const (
	// Success indicates the required services in the GMConnector was ready
	ConnectorSuccess ConditionType = "Success"
	// Failed indicates the GMConnector failed to get required service ready
	ConnectorFailed ConditionType = "Failed"
)

// GMConnectorCondition describes a condition of a GMConnector object
type GMConnectorCondition struct {
	// type of the condition. Known conditions are "Success", and "Failed".
	//
	// A "Success" indicating the required services in the GMConnector was ready.
	//
	// A "Failed" indicating the GMConnector failed to get required service ready.
	Type ConditionType `json:"type,omitempty"`
	// message contains a human readable message with details about the GMConnector state
	// +optional
	Message string `json:"message,omitempty"`
	// The reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty" description:"one-word CamelCase reason for the condition's last transition"`
	// lastUpdateTime is the time of the last update to this condition
	// +optional
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
}

// GMConnectorStatus defines the observed state of GMConnector.
// +k8s:openapi-gen=true
type GMConnectorStatus struct {
	// conditions applied to the GMConnector. Known conditions are "Success", and "Failed".
	// +optional
	Condition GMConnectorCondition `json:"condition,omitempty"`

	// Conditions for GMConnector
	Status string `json:"status,omitempty"`

	// AccessURL of the entrypoint for the GMConnector
	// +optional
	AccessURL string `json:"accessUrl,omitempty"`

	// Annotations is additional Status fields for the Resource to save some
	// additional State as well as convey more information to the user. This is
	// roughly akin to Annotations on any k8s resource, just the reconciler conveying
	// richer information outwards.
	Annotations map[string]string `json:"annotations,omitempty"`
}

// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="URL",type="string",JSONPath=".status.accessUrl"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.status"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:path=gmconnectors,shortName=gmc,singular=gmconnectors
// GMConnector is the Schema for the gmconnectors API
type GMConnector struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GMConnectorSpec   `json:"spec,omitempty"`
	Status GMConnectorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// GMConnectorList contains a list of GMConnector
type GMConnectorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GMConnector `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GMConnector{}, &GMConnectorList{})
}
