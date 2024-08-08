/*
* Copyright (C) 2024 Intel Corporation
* SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	mcv1alpha3 "github.com/opea-project/GenAIInfra/microservices-connector/api/v1alpha3"
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

	res, _, err := routeStep("root", gmcGraph, jsonBytes, jsonBytes, headers)
	if err != nil {
		return
	}
	var response map[string]interface{}
	responseBytes, rerr := io.ReadAll(res)
	if rerr != nil {
		t.Fatalf("Error while reading the response body: %v", rerr)
		return
	}
	err = json.Unmarshal(responseBytes, &response)
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
	res, _, err := routeStep("root", gmcGraph, jsonBytes, jsonBytes, headers)
	if err != nil {
		return
	}
	var response map[string]interface{}
	responseBytes, rerr := io.ReadAll(res)
	if rerr != nil {
		t.Fatalf("Error while reading the response body")
		return
	}
	err = json.Unmarshal(responseBytes, &response)
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
	res, _, err := routeStep("root", gmcGraph, jsonBytes, jsonBytes, headers)
	if err != nil {
		return
	}
	var response map[string]interface{}
	responseBytes, rerr := io.ReadAll(res)
	if rerr != nil {
		t.Fatalf("Error while reading the response body")
		return
	}
	err = json.Unmarshal(responseBytes, &response)
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
	responseBytes, rerr := io.ReadAll(res)
	if rerr != nil {
		t.Fatalf("Error while reading the response body")
		return
	}
	err = json.Unmarshal(responseBytes, &response)
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
func TestPrepareErrorResponse(t *testing.T) {
	err := errors.New("test error")
	errorMessage := "Test error message"
	expectedResponse := []byte(`{"error":"Test error message","cause":"test error"}`)

	response := prepareErrorResponse(err, errorMessage)

	assert.Equal(t, expectedResponse, response)
}

func TestMcGraphHandler(t *testing.T) {
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

	mockGraph := mcv1alpha3.GMConnector{
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
			},
		},
	}

	mcGraph = &mockGraph
	// Create a request with a sample input
	input := []byte(`{"instances": ["test", "test2"]}`)
	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(input))
	if err != nil {
		t.Fatal(err)
	}

	// Create a response recorder to capture the response
	rr := httptest.NewRecorder()

	// Call the handler function
	mcGraphHandler(rr, req)

	// Check the response status code
	// if rr.Code != http.StatusOK {
	// 	t.Errorf("expected status code %d, got %d", http.StatusOK, rr.Code)
	// }

	// // Check the response body
	// expectedResponse := []byte(`{"service1": {"predictions": "1"}, "service2": {"predictions": "2"}}`)
	// if !bytes.Equal(rr.Body.Bytes(), expectedResponse) {
	// 	t.Errorf("expected response body %s, got %s", expectedResponse, rr.Body.String())
	// }
}

// Mock os.Exit to prevent exiting the process
//
//	func mockOsExit() func() {
//		originalOsExit := os.Exit
//		os.Exit = func(code int) {}
//		return func() { os.Exit = originalOsExit }
//	}
func TestMain(t *testing.T) {
	// Create a new HTTP request
	_, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	jsonGraph := `
	{
		"root": {
			"routerType": "sequence",
			"steps": [
				{
					"stepName": "step1",
					"executor": {
						"nodeName": "animal-categorize",
						"internalService": {
							"namespace": "default",
							"serviceName": "tei-embedding-service"
						}
					}
				},
				{
					"stepName": "step2",
					"executor": {
						"nodeName": "breed-categorize",
						"internalService": {
							"namespace": "default",
							"serviceName": "tgi-service"
						}
					},
					"condition": "predictions.#(label==\"dog\")"
				}
			]
		}
	}
	`

	os.Args = []string{"main", "--graph-json", jsonGraph}

	// Mock os.Exit
	// defer mockOsExit()()

	// Call the main function, which handles the request
	go main()

	// Simulate doing some work or waiting for a condition
	time.Sleep(2 * time.Second)
}

func TestMcGraphHandler_Timeout(t *testing.T) {
	// Mock server with a context timeout of 1 second
	handler := http.HandlerFunc(mcGraphHandler)
	server := httptest.NewServer(handler)
	defer server.Close()

	client := server.Client()

	// Create a request with a short context timeout
	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	ctx, cancel := context.WithTimeout(req.Context(), time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("error closing response body: %v", err)
		}
	}()

	// Check if the response status code is StatusInternalServerError (500)
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status %d; got %d", http.StatusInternalServerError, resp.StatusCode)
	}

	// Read and validate the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	expectedErrorMessage := "Failed to process request"
	if !strings.Contains(string(body), expectedErrorMessage) {
		t.Errorf("expected error message '%s'; got '%s'", expectedErrorMessage, string(body))
	}
}

func TestMcDataHandler(t *testing.T) {
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

	// Create a buffer to store the multipart form data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add form fields
	err = writer.WriteField("key", "value")
	if err != nil {
		t.Fatalf("failed to write form field: %v", err)
	}

	// Add a file field
	part, err := writer.CreateFormFile("file", "filename.txt")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	_, err = part.Write([]byte("file content"))
	if err != nil {
		t.Fatalf("failed to write to form file: %v", err)
	}

	// Close the writer to finalize the multipart form
	err = writer.Close()
	if err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	// Create a new HTTP request with the multipart form data
	req := httptest.NewRequest(http.MethodPost, "/dataprep", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Create a ResponseRecorder to capture the response
	rr := httptest.NewRecorder()

	// Mock the mcGraph data
	mcGraph = &mcv1alpha3.GMConnector{
		Spec: mcv1alpha3.GMConnectorSpec{
			Nodes: map[string]mcv1alpha3.Router{
				"root": {
					Steps: []mcv1alpha3.Step{
						{
							StepName:   "DataPrep",
							ServiceURL: service1Url.String(),
							Executor: mcv1alpha3.Executor{
								InternalService: mcv1alpha3.GMCTarget{
									NameSpace:   "default",
									ServiceName: "example-service",
								},
							},
						},
					},
				},
			},
		},
	}

	// Call the mcDataHandler function
	mcDataHandler(rr, req)

	// Check the response status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body if needed
	expected := "{\"predictions\":\"1\"}"
	if strings.TrimSpace(rr.Body.String()) != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
