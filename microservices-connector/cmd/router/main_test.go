/*
* Copyright (C) 2024 Intel Corporation
* SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	mcv1alpha3 "github.com/opea-project/GenAIInfra/genai-microservices-connector/api/v1alpha3"
	"github.com/stretchr/testify/assert"
	"knative.dev/pkg/apis"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func init() {
	logf.SetLogger(zap.New())
}

func TestSimpleModelChainer(t *testing.T) {
	// Start a local HTTP server
	service1 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, err := io.ReadAll(req.Body)
		if err != nil {
			return
		}
		response := map[string]interface{}{"predictions": "1"}
		responseBytes, _ := json.Marshal(response)
		_, err = rw.Write(responseBytes)
		if err != nil {
			return
		}
	}))
	service1Url, err := apis.ParseURL(service1.URL)
	if err != nil {
		t.Fatalf("Failed to parse model url")
	}
	defer service1.Close()
	service2 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, err := io.ReadAll(req.Body)
		if err != nil {
			return
		}
		response := map[string]interface{}{"predictions": "2"}
		responseBytes, _ := json.Marshal(response)
		_, err = rw.Write(responseBytes)
		if err != nil {
			return
		}
	}))
	service2Url, err := apis.ParseURL(service2.URL)
	if err != nil {
		t.Fatalf("Failed to parse model url")
	}
	defer service2.Close()

	gmcGraph := mcv1alpha3.GMConnector{
		Spec: mcv1alpha3.GMConnectorSpec{
			Nodes: map[string]mcv1alpha3.Router{
				"root": {
					RouterType: mcv1alpha3.Sequence,
					Steps: []mcv1alpha3.Step{
						{
							StepName:   "service1",
							ServiceURL: service1Url.String(),
							Executor: mcv1alpha3.Executor{
								InternalService: mcv1alpha3.GMCTarget{
									NameSpace:   "default",
									ServiceName: "embedding-service",
								},
							},
						},
						{
							StepName:   "service2",
							ServiceURL: service2Url.String(),
							Executor: mcv1alpha3.Executor{
								InternalService: mcv1alpha3.GMCTarget{
									NameSpace:   "default",
									ServiceName: "tei-embedding-service",
								},
							},
							Data: "$response",
						},
					},
				},
			},
		},
	}
	input := map[string]interface{}{
		"instances": []string{
			"test",
			"test2",
		},
	}
	jsonBytes, _ := json.Marshal(input)
	headers := http.Header{
		"Authorization": {"Bearer Token"},
	}

	res, _, err := routeStep("root", gmcGraph, jsonBytes, headers)
	if err != nil {
		return
	}
	var response map[string]interface{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return
	}
	expectedResponse := map[string]interface{}{
		"predictions": "2",
	}
	fmt.Printf("final response:%v\n", response)
	assert.Equal(t, expectedResponse, response)
}

func TestSimpleServiceEnsemble(t *testing.T) {
	// Start a local HTTP server
	service1 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, err := io.ReadAll(req.Body)
		if err != nil {
			return
		}
		response := map[string]interface{}{"predictions": "1"}
		responseBytes, err := json.Marshal(response)
		if err != nil {
			return
		}
		_, err = rw.Write(responseBytes)
		if err != nil {
			return
		}
	}))
	service1Url, err := apis.ParseURL(service1.URL)
	if err != nil {
		t.Fatalf("Failed to parse model url")
	}
	defer service1.Close()
	service2 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, err := io.ReadAll(req.Body)
		if err != nil {
			return
		}
		response := map[string]interface{}{"predictions": "2"}
		responseBytes, err := json.Marshal(response)
		if err != nil {
			return
		}
		_, err = rw.Write(responseBytes)
		if err != nil {
			return
		}
	}))
	service2Url, err := apis.ParseURL(service2.URL)
	if err != nil {
		t.Fatalf("Failed to parse model url")
	}
	defer service2.Close()

	gmcGraph := mcv1alpha3.GMConnector{
		Spec: mcv1alpha3.GMConnectorSpec{
			Nodes: map[string]mcv1alpha3.Router{
				"root": {
					RouterType: mcv1alpha3.Ensemble,
					Steps: []mcv1alpha3.Step{
						{
							StepName:   "service1",
							ServiceURL: service1Url.String(),
							Executor: mcv1alpha3.Executor{
								InternalService: mcv1alpha3.GMCTarget{
									NameSpace:   "default",
									ServiceName: "embedding-service",
								},
							},
						},
						{
							StepName:   "service2",
							ServiceURL: service2Url.String(),
							Executor: mcv1alpha3.Executor{
								InternalService: mcv1alpha3.GMCTarget{
									NameSpace:   "default",
									ServiceName: "tei-embedding-service",
								},
							},
						},
					},
				},
			},
		},
	}

	input := map[string]interface{}{
		"instances": []string{
			"test",
			"test2",
		},
	}
	jsonBytes, _ := json.Marshal(input)
	headers := http.Header{
		"Authorization": {"Bearer Token"},
	}
	res, _, err := routeStep("root", gmcGraph, jsonBytes, headers)
	if err != nil {
		return
	}
	var response map[string]interface{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return
	}
	expectedResponse := map[string]interface{}{
		"service1": map[string]interface{}{
			"predictions": "1",
		},
		"service2": map[string]interface{}{
			"predictions": "2",
		},
	}
	fmt.Printf("final response:%v\n", response)
	assert.Equal(t, expectedResponse, response)
}

func TestMCWithCondition(t *testing.T) {
	// Start a local HTTP server
	service1 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, err := io.ReadAll(req.Body)
		if err != nil {
			return
		}
		response := map[string]interface{}{
			"predictions": []map[string]interface{}{
				{
					"label": "cat",
					"score": []float32{
						0.1, 0.9,
					},
				},
			},
		}
		responseBytes, err := json.Marshal(response)
		if err != nil {
			return
		}
		_, err = rw.Write(responseBytes)
		if err != nil {
			return
		}
	}))
	service1Url, err := apis.ParseURL(service1.URL)
	if err != nil {
		t.Fatalf("Failed to parse model url")
	}
	defer service1.Close()
	service2 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, err := io.ReadAll(req.Body)
		if err != nil {
			return
		}
		response := map[string]interface{}{
			"predictions": []map[string]interface{}{
				{
					"label": "dog",
					"score": []float32{
						0.8, 0.2,
					},
				},
			},
		}
		responseBytes, err := json.Marshal(response)
		if err != nil {
			return
		}
		_, err = rw.Write(responseBytes)
		if err != nil {
			return
		}
	}))
	service2Url, err := apis.ParseURL(service2.URL)
	if err != nil {
		t.Fatalf("Failed to parse model url")
	}
	defer service2.Close()

	// Start a local HTTP server
	service3 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, err := io.ReadAll(req.Body)
		if err != nil {
			return
		}
		response := map[string]interface{}{
			"predictions": []map[string]interface{}{
				{
					"label": "beagle",
					"score": []float32{
						0.1, 0.9,
					},
				},
			},
		}
		responseBytes, err := json.Marshal(response)
		if err != nil {
			return
		}
		_, err = rw.Write(responseBytes)
		if err != nil {
			return
		}
	}))
	service3Url, err := apis.ParseURL(service3.URL)
	if err != nil {
		t.Fatalf("Failed to parse model url")
	}
	defer service3.Close()
	service4 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, err := io.ReadAll(req.Body)
		if err != nil {
			return
		}
		response := map[string]interface{}{
			"predictions": []map[string]interface{}{
				{
					"label": "poodle",
					"score": []float32{
						0.8, 0.2,
					},
				},
			},
		}
		responseBytes, err := json.Marshal(response)
		if err != nil {
			return
		}
		_, err = rw.Write(responseBytes)
		if err != nil {
			return
		}
	}))
	service4Url, err := apis.ParseURL(service4.URL)
	if err != nil {
		t.Fatalf("Failed to parse model url")
	}
	defer service4.Close()

	gmcGraph := mcv1alpha3.GMConnector{
		Spec: mcv1alpha3.GMConnectorSpec{
			Nodes: map[string]mcv1alpha3.Router{
				"root": {
					RouterType: mcv1alpha3.Sequence,
					Steps: []mcv1alpha3.Step{
						{
							StepName: "step1",
							Executor: mcv1alpha3.Executor{
								NodeName: "animal-categorize",
								InternalService: mcv1alpha3.GMCTarget{
									NameSpace:   "default",
									ServiceName: "tei-embedding-service",
								},
							},
						},
						{
							StepName: "step2",
							Executor: mcv1alpha3.Executor{
								NodeName: "breed-categorize",
								InternalService: mcv1alpha3.GMCTarget{
									NameSpace:   "default",
									ServiceName: "tgi-service",
								},
							},
							Condition: "predictions.#(label==\"dog\")",
						},
					},
				},
				"animal-categorize": {
					RouterType: mcv1alpha3.Switch,
					Steps: []mcv1alpha3.Step{
						{
							StepName:   "service1",
							ServiceURL: service1Url.String(),
							Executor: mcv1alpha3.Executor{
								InternalService: mcv1alpha3.GMCTarget{
									NameSpace:   "default",
									ServiceName: "tei-embedding-service",
								},
							},
							Condition: "instances.#(modelId==\"1\")",
						},
						{
							StepName:   "service2",
							ServiceURL: service2Url.String(),
							Executor: mcv1alpha3.Executor{
								InternalService: mcv1alpha3.GMCTarget{
									NameSpace:   "default",
									ServiceName: "tgi-service",
								},
							},
							Condition: "instances.#(modelId==\"2\")",
						},
					},
				},
				"breed-categorize": {
					RouterType: mcv1alpha3.Ensemble,
					Steps: []mcv1alpha3.Step{
						{
							StepName:   "service3",
							ServiceURL: service3Url.String(),
							Executor: mcv1alpha3.Executor{
								InternalService: mcv1alpha3.GMCTarget{
									NameSpace:   "default",
									ServiceName: "tei-embedding-service",
								},
							},
						},
						{
							StepName:   "service4",
							ServiceURL: service4Url.String(),
							Executor: mcv1alpha3.Executor{
								InternalService: mcv1alpha3.GMCTarget{
									NameSpace:   "default",
									ServiceName: "tgi-service",
								},
							},
						},
					},
				},
			},
		},
	}
	input := map[string]interface{}{
		"instances": []map[string]string{
			{"modelId": "2"},
		},
	}
	jsonBytes, _ := json.Marshal(input)
	headers := http.Header{
		"Authorization": {"Bearer Token"},
	}
	res, _, err := routeStep("root", gmcGraph, jsonBytes, headers)
	if err != nil {
		return
	}
	var response map[string]interface{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return
	}
	expectedservice3Response := map[string]interface{}{
		"predictions": []interface{}{
			map[string]interface{}{
				"label": "beagle",
				"score": []interface{}{
					0.1, 0.9,
				},
			},
		},
	}

	expectedservice4Response := map[string]interface{}{
		"predictions": []interface{}{
			map[string]interface{}{
				"label": "poodle",
				"score": []interface{}{
					0.8, 0.2,
				},
			},
		},
	}
	fmt.Printf("final response:%v\n", response)
	assert.Equal(t, expectedservice3Response, response["service3"])
	assert.Equal(t, expectedservice4Response, response["service4"])
}

func TestCallServiceWhenNoneHeadersToPropagateIsEmpty(t *testing.T) {
	// Start a local HTTP server
	service1 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, err := io.ReadAll(req.Body)
		if err != nil {
			return
		}
		// Putting headers as part of response so that we can assert the headers' presence later
		response := make(map[string]interface{})
		response["predictions"] = "1"
		responseBytes, err := json.Marshal(response)
		if err != nil {
			return
		}
		_, err = rw.Write(responseBytes)
		if err != nil {
			return
		}
	}))
	service1Url, err := apis.ParseURL(service1.URL)
	if err != nil {
		t.Fatalf("Failed to parse model url")
	}
	defer service1.Close()

	input := map[string]interface{}{
		"instances": []string{
			"test",
			"test2",
		},
	}
	jsonBytes, _ := json.Marshal(input)
	headers := http.Header{
		"Authorization":   {"Bearer Token"},
		"Test-Header-Key": {"Test-Header-Value"},
	}

	step := &mcv1alpha3.Step{
		StepName:   "service1",
		ServiceURL: service1Url.String(),
		Executor: mcv1alpha3.Executor{
			InternalService: mcv1alpha3.GMCTarget{
				NameSpace:   "default",
				ServiceName: "tei-embedding-service",
			},
		},
		Condition: "instances.#(modelId==\"1\")",
	}

	res, _, err := callService(step, service1Url.String(), jsonBytes, headers)
	if err != nil {
		return
	}
	var response map[string]interface{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return
	}
	expectedResponse := map[string]interface{}{
		"predictions": "1",
	}
	fmt.Printf("final response:%v\n", response)
	assert.Equal(t, expectedResponse, response)
}

func TestMalformedURL(t *testing.T) {
	malformedURL := "http://single-1.default.{$your-domain}/switch"
	step := &mcv1alpha3.Step{
		StepName:   "service1",
		ServiceURL: malformedURL,
		Executor: mcv1alpha3.Executor{
			InternalService: mcv1alpha3.GMCTarget{
				NameSpace:   "default",
				ServiceName: "tei-embedding-service",
			},
		},
		Condition: "instances.#(modelId==\"1\")",
	}
	_, response, err := callService(step, malformedURL, []byte{}, http.Header{})
	if err != nil {
		assert.Equal(t, 500, response)
	}
}
