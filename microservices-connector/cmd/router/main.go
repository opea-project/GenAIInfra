/*
* Copyright (C) 2024 Intel Corporation
* SPDX-License-Identifier: Apache-2.0
 */

/* Modifications made to this file by [Intel] on [2024]
*  Portions of this file are derived from kserve: https://github.com/kserve/kserve
*  Copyright 2022 The KServe Author
 */

package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	// "regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/gjson"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	// "crypto/rand"
	// "math/big"

	mcv1alpha3 "github.com/opea-project/GenAIInfra/microservices-connector/api/v1alpha3"
	flag "github.com/spf13/pflag"
)

var (
	jsonGraph       = flag.String("graph-json", "", "serialized json graph def")
	log             = logf.Log.WithName("GMCGraphRouter")
	mcGraph         *mcv1alpha3.GMConnector
	defaultNodeName = "root"
)

const (
	ChunkSize   = 1024
	ServiceURL  = "serviceUrl"
	ServiceNode = "node"
	DataPrep    = "DataPrep"
)

type EnsembleStepOutput struct {
	StepResponse   map[string]interface{}
	StepStatusCode int
}

type GMCGraphRoutingError struct {
	ErrorMessage string `json:"error"`
	Cause        string `json:"cause"`
}

func (e *GMCGraphRoutingError) Error() string {
	return fmt.Sprintf("%s. %s", e.ErrorMessage, e.Cause)
}

func timeTrack(start time.Time, nodeOrStep string, name string) {
	elapsed := time.Since(start)
	log.Info("elapsed time", nodeOrStep, name, "time", elapsed)
}

func isSuccessFul(statusCode int) bool {
	if statusCode >= 200 && statusCode <= 299 {
		return true
	}
	return false
}

func pickupRouteByCondition(initInput []byte, condition string) bool {
	//sample config supported by gjson
	//"instances" : [
	//	{"model_id", "1"},
	//  ]
	// sample condition support by gjson query: "instances.#(modelId==\"1\")""
	if !gjson.ValidBytes(initInput) {
		fmt.Println("the initInput json format is invalid")
		return false
	}

	if gjson.GetBytes(initInput, condition).Exists() {
		return true
	}
	// ' and # will define a gjson query
	if strings.ContainsAny(condition, ".") || strings.ContainsAny(condition, "#") {
		return false
	}
	// key == value without nested json
	// sample config support by direct query {"model_id", "1"}
	// smaple condition support by json query: "modelId==\"1\""
	index := strings.Index(condition, "==")
	if index == -1 {
		fmt.Println("No '==' found in the route with condition [", condition, "]")
		return false
	} else {
		key := strings.TrimSpace(condition[:index])
		value := strings.TrimSpace(condition[index+2:])
		v := gjson.GetBytes(initInput, key).String()
		if v == value {
			return true
		}
	}
	return false
}

func prepareErrorResponse(err error, errorMessage string) []byte {
	igRoutingErr := &GMCGraphRoutingError{
		errorMessage,
		fmt.Sprintf("%v", err),
	}
	errorResponseBytes, err := json.Marshal(igRoutingErr)
	if err != nil {
		log.Error(err, "marshalling error")
	}
	return errorResponseBytes
}

func callService(step *mcv1alpha3.Step, serviceUrl string, input []byte, headers http.Header) ([]byte, int, error) {
	defer timeTrack(time.Now(), "step", serviceUrl)
	log.Info("Entering callService", "url", serviceUrl)

	// log the http header from the original request
	log.Info("Print the http request headers", "HTTP_Header", headers)

	if step.InternalService.Config != nil {
		err := os.Setenv("no_proxy", step.InternalService.Config["no_proxy"])
		if err != nil {
			log.Error(err, "Error setting environment variable", "no_proxy", step.InternalService.Config["no_proxy"])
			return nil, 400, err
		}
	}
	req, err := http.NewRequest("POST", serviceUrl, bytes.NewBuffer(input))
	if err != nil {
		log.Error(err, "An error occurred while preparing request object with serviceUrl.", "serviceUrl", serviceUrl)
		return nil, 500, err
	}

	if val := req.Header.Get("Content-Type"); val == "" {
		req.Header.Add("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Error(err, "An error has occurred while calling service", "service", serviceUrl)
		return nil, 500, err
	}

	defer func() {
		if resp.Body != nil {
			err := resp.Body.Close()
			if err != nil {
				log.Error(err, "An error has occurred while closing the response body")
			}
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error(err, "Error while reading the response")
	}

	return body, resp.StatusCode, err
}

// Use step service name to create a K8s service if serviceURL is empty
// TODO: add more features here, such as K8s service selector, labels, etc.
func getServiceURLByStepTarget(step *mcv1alpha3.Step, svcNameSpace string) string {
	if step.ServiceURL == "" {
		serviceURL := fmt.Sprintf("http://%s.%s.svc.cluster.local", step.StepName, svcNameSpace)
		return serviceURL
	}
	return step.ServiceURL
}

func executeStep(
	step *mcv1alpha3.Step,
	graph mcv1alpha3.GMConnector,
	initInput []byte,
	input []byte,
	headers http.Header,
) ([]byte, int, error) {
	if step.NodeName != "" {
		// when nodeName is specified make a recursive call for routing to next step
		return routeStep(step.NodeName, graph, initInput, input, headers)
	}
	serviceURL := getServiceURLByStepTarget(step, graph.Namespace)
	return callService(step, serviceURL, input, headers)
}

func handleSwitchNode(
	route *mcv1alpha3.Step,
	graph mcv1alpha3.GMConnector,
	initInput []byte,
	request []byte,
	headers http.Header,
) ([]byte, int, error) {
	var statusCode int
	var responseBytes []byte
	var err error
	stepType := ServiceURL
	if route.NodeName != "" {
		stepType = ServiceNode
	}
	log.Info("Starting execution of step", "Node Name", route.NodeName, "type", stepType, "stepName", route.StepName)
	if responseBytes, statusCode, err = executeStep(route, graph, initInput, request, headers); err != nil {
		return nil, 500, err
	}

	if route.Dependency == mcv1alpha3.Hard && !isSuccessFul(statusCode) {
		log.Info(
			"This step is a hard dependency and it is unsuccessful",
			"stepName",
			route.StepName,
			"statusCode",
			statusCode,
		)
	}
	return responseBytes, statusCode, nil
}

func handleSwitchPipeline(nodeName string,
	graph mcv1alpha3.GMConnector,
	initInput []byte,
	input []byte,
	headers http.Header,
) ([]byte, int, error) {
	currentNode := graph.Spec.Nodes[nodeName]
	var statusCode int
	var responseBytes []byte
	var err error
	for index, route := range currentNode.Steps {
		if route.InternalService.IsDownstreamService {
			log.Info(
				"InternalService DownstreamService is true, skip the execution of step",
				"type",
				currentNode.RouterType,
				"stepName",
				route.StepName,
			)
			continue
		}
		log.Info("Current Step Information", "Node Name", nodeName, "Step Index", index)
		request := input
		if route.Data == "$response" && index > 0 {
			request = responseBytes
		}
		if route.Condition == "" {
			responseBytes, statusCode, err = handleSwitchNode(&route, graph, initInput, request, headers)
			if err != nil {
				return responseBytes, statusCode, err
			}
		} else {
			if pickupRouteByCondition(initInput, route.Condition) {
				responseBytes, statusCode, err = handleSwitchNode(&route, graph, initInput, request, headers)
				if err != nil {
					return responseBytes, statusCode, err
				}
			}
		}
		log.Info("Print Response Bytes", "Response Bytes", responseBytes, "Status Code", statusCode)
	}
	return responseBytes, statusCode, err
}

func handleEnsemblePipeline(nodeName string,
	graph mcv1alpha3.GMConnector,
	initInput []byte,
	input []byte,
	headers http.Header,
) ([]byte, int, error) {
	currentNode := graph.Spec.Nodes[nodeName]
	ensembleRes := make([]chan EnsembleStepOutput, len(currentNode.Steps))
	errChan := make(chan error)
	for i := range currentNode.Steps {
		step := &currentNode.Steps[i]
		stepType := ServiceURL
		if step.NodeName != "" {
			stepType = ServiceNode
		}
		log.Info("Starting execution of step", "type", stepType, "stepName", step.StepName)
		resultChan := make(chan EnsembleStepOutput)
		ensembleRes[i] = resultChan
		go func() {
			output, statusCode, err := executeStep(step, graph, initInput, input, headers)
			if err == nil {
				var res map[string]interface{}
				if err = json.Unmarshal(output, &res); err == nil {
					resultChan <- EnsembleStepOutput{
						StepResponse:   res,
						StepStatusCode: statusCode,
					}
					return
				}
			}
			errChan <- err
		}()
	}
	// merge responses from parallel steps
	response := map[string]interface{}{}
	ensembleStepOutput := EnsembleStepOutput{}
	for i, resultChan := range ensembleRes {
		key := currentNode.Steps[i].StepName
		if key == "" {
			key = strconv.Itoa(i) // Use index if no step name
		}
		select {
		case ensembleStepOutput = <-resultChan:
			if !isSuccessFul(ensembleStepOutput.StepStatusCode) && currentNode.Steps[i].Dependency == mcv1alpha3.Hard {
				log.Info(
					"This step is a hard dependency and it is unsuccessful",
					"stepName",
					currentNode.Steps[i].StepName,
					"statusCode",
					ensembleStepOutput.StepStatusCode,
				)
				stepResponse, _ := json.Marshal(ensembleStepOutput.StepResponse)
				return stepResponse, ensembleStepOutput.StepStatusCode, nil
			} else {
				response[key] = ensembleStepOutput.StepResponse
			}
		case err := <-errChan:
			return nil, 500, err
		}
	}
	// return json.Marshal(response)
	combinedResponse, _ := json.Marshal(response) // TODO check if you need err handling for Marshalling
	return combinedResponse, 200, nil
}

func handleSequencePipeline(nodeName string,
	graph mcv1alpha3.GMConnector,
	initInput []byte,
	input []byte,
	headers http.Header,
) ([]byte, int, error) {
	currentNode := graph.Spec.Nodes[nodeName]
	var statusCode int
	var responseBytes []byte
	var err error
	for i := range currentNode.Steps {
		step := &currentNode.Steps[i]
		stepType := ServiceURL
		if step.NodeName != "" {
			stepType = ServiceNode
		}
		if step.InternalService.IsDownstreamService {
			log.Info(
				"InternalService DownstreamService is true, skip the execution of step",
				"type",
				stepType,
				"stepName",
				step.StepName,
			)
			continue
		}
		log.Info("Starting execution of step", "type", stepType, "stepName", step.StepName)
		request := input
		if step.Data == "$response" && i > 0 {
			request = responseBytes
		}
		if step.Condition != "" {
			if !gjson.ValidBytes(responseBytes) {
				return nil, 500, fmt.Errorf("invalid response")
			}
			// if the condition does not match for the step in the sequence we stop and return the response
			if !gjson.GetBytes(responseBytes, step.Condition).Exists() {
				return responseBytes, 500, nil
			}
		}
		if responseBytes, statusCode, err = executeStep(step, graph, initInput, request, headers); err != nil {
			return nil, 500, err
		}
		log.Info("Print Response Bytes", "Response Bytes", responseBytes, "Status Code", statusCode)
		/*
		   Only if a step is a hard dependency, we will check for its success.
		*/
		if step.Dependency == mcv1alpha3.Hard {
			if !isSuccessFul(statusCode) {
				log.Info(
					"This step is a hard dependency and it is unsuccessful",
					"stepName",
					step.StepName,
					"statusCode",
					statusCode,
				)
				// Stop the execution of sequence right away if step is a hard dependency and is unsuccessful
				return responseBytes, statusCode, nil
			}
		}
	}
	return responseBytes, statusCode, nil
}

func routeStep(nodeName string,
	graph mcv1alpha3.GMConnector,
	initInput, input []byte,
	headers http.Header,
) ([]byte, int, error) {
	defer timeTrack(time.Now(), "node", nodeName)
	currentNode := graph.Spec.Nodes[nodeName]
	log.Info("Current Node", "Node Name", nodeName)

	if currentNode.RouterType == mcv1alpha3.Switch {
		return handleSwitchPipeline(nodeName, graph, initInput, input, headers)
	}

	if currentNode.RouterType == mcv1alpha3.Ensemble {
		return handleEnsemblePipeline(nodeName, graph, initInput, input, headers)
	}

	if currentNode.RouterType == mcv1alpha3.Sequence {
		return handleSequencePipeline(nodeName, graph, initInput, input, headers)
	}
	log.Error(nil, "invalid route type", "type", currentNode.RouterType)
	return nil, 500, fmt.Errorf("invalid route type: %v", currentNode.RouterType)
}

func mcGraphHandler(w http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), time.Minute)
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)

		inputBytes, _ := io.ReadAll(req.Body)
		response, statusCode, err := routeStep(defaultNodeName, *mcGraph, inputBytes, inputBytes, req.Header)

		if err != nil {
			log.Error(err, "failed to process request")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			if _, err := w.Write(prepareErrorResponse(err, "Failed to process request")); err != nil {
				log.Error(err, "failed to write mcGraphHandler response")
			}
			return
		}
		if json.Valid(response) {
			w.Header().Set("Content-Type", "application/json")
		}
		w.WriteHeader(statusCode)

		writer := bufio.NewWriter(w)
		defer func() {
			if err := writer.Flush(); err != nil {
				log.Error(err, "error flushing writer when processing response")
			}
		}()

		for start := 0; start < len(response); start += ChunkSize {
			end := start + ChunkSize
			if end > len(response) {
				end = len(response)
			}
			if _, err := writer.Write(response[start:end]); err != nil {
				log.Error(err, "failed to write mcGraphHandler response")
				return
			}
		}
	}()

	select {
	case <-ctx.Done():
		log.Error(errors.New("failed to process request"), "request timed out")
		http.Error(w, "request timed out", http.StatusGatewayTimeout)
	case <-done:
		log.Info("mcGraphHandler is done")
	}
}

func mcDataHandler(w http.ResponseWriter, r *http.Request) {
	var isDataHandled bool
	serviceName := r.Header.Get("SERVICE_NAME")
	defaultNode := mcGraph.Spec.Nodes[defaultNodeName]
	for i := range defaultNode.Steps {
		step := &defaultNode.Steps[i]
		if DataPrep == step.StepName {
			if serviceName != "" && serviceName != step.InternalService.ServiceName {
				continue
			}
			log.Info("Starting execution of step", "stepName", step.StepName)
			serviceURL := getServiceURLByStepTarget(step, mcGraph.Namespace)
			log.Info("ServiceURL is", "serviceURL", serviceURL)
			// Parse the multipart form in the request
			// err := r.ParseMultipartForm(64 << 20) // 64 MB is the default used by ParseMultipartForm

			// Set no limit on multipart form size
			err := r.ParseMultipartForm(0)
			if err != nil {
				http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
				return
			}
			// Create a buffer to hold the new form data
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)
			// Copy all form fields from the original request to the new request
			for key, values := range r.MultipartForm.Value {
				for _, value := range values {
					err := writer.WriteField(key, value)
					if err != nil {
						handleMultipartError(writer, err)
						http.Error(w, "Failed to write form field", http.StatusInternalServerError)
						return
					}
				}
			}
			// Copy all files from the original request to the new request
			for key, fileHeaders := range r.MultipartForm.File {
				for _, fileHeader := range fileHeaders {
					file, err := fileHeader.Open()
					if err != nil {
						handleMultipartError(writer, err)
						http.Error(w, "Failed to open file", http.StatusInternalServerError)
						return
					}
					defer func() {
						if err := file.Close(); err != nil {
							log.Error(err, "error closing file")
						}
					}()
					part, err := writer.CreateFormFile(key, fileHeader.Filename)
					if err != nil {
						handleMultipartError(writer, err)
						http.Error(w, "Failed to create form file", http.StatusInternalServerError)
						return
					}
					_, err = io.Copy(part, file)
					if err != nil {
						handleMultipartError(writer, err)
						http.Error(w, "Failed to copy file", http.StatusInternalServerError)
						return
					}
				}
			}

			err = writer.Close()
			if err != nil {
				http.Error(w, "Failed to close writer", http.StatusInternalServerError)
				return
			}

			req, err := http.NewRequest(r.Method, serviceURL, &buf)
			if err != nil {
				http.Error(w, "Failed to create new request", http.StatusInternalServerError)
				return
			}
			// Copy headers from the original request to the new request
			for key, values := range r.Header {
				for _, value := range values {
					req.Header.Add(key, value)
				}
			}
			req.Header.Set("Content-Type", writer.FormDataContentType())
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				http.Error(w, "Failed to send request to backend", http.StatusInternalServerError)
				return
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					log.Error(err, "error closing response body stream")
				}
			}()
			// Copy the response headers from the backend service to the original client
			for key, values := range resp.Header {
				for _, value := range values {
					w.Header().Add(key, value)
				}
			}
			w.WriteHeader(resp.StatusCode)
			// Copy the response body from the backend service to the original client
			_, err = io.Copy(w, resp.Body)
			if err != nil {
				log.Error(err, "failed to copy response body")
			}
			isDataHandled = true
		}
	}

	if !isDataHandled {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(404)
		if _, err := w.Write([]byte("\n Message: None dataprep endpoint is available! \n")); err != nil {
			log.Info("Message: ", "failed to write mcDataHandler response")
		}
	}
}

func handleMultipartError(writer *multipart.Writer, err error) {
	// In case of an error, close the writer to clean up
	werr := writer.Close()
	if werr != nil {
		log.Error(werr, "Error during close writer")
		return
	}
	// Handle the error as needed, such as logging or returning an error response
	log.Error(err, "Error during multipart creation")
}

func initializeRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", mcGraphHandler)
	mux.HandleFunc("/dataprep", mcDataHandler)
	return mux
}

func main() {
	flag.Parse()
	logf.SetLogger(zap.New())

	mcGraph = &mcv1alpha3.GMConnector{}
	err := json.Unmarshal([]byte(*jsonGraph), mcGraph)
	if err != nil {
		log.Error(err, "failed to unmarshall gmc graph json")
		os.Exit(1)
	}

	mcRouter := initializeRoutes()

	server := &http.Server{
		// specify the address and port
		Addr: ":8080",
		// specify the HTTP routers
		Handler: mcRouter,
		// set the maximum duration for reading the entire request, including the body
		ReadTimeout: time.Minute,
		// set the maximum duration before timing out writes of the response
		WriteTimeout: time.Minute,
		// set the maximum amount of time to wait for the next request when keep-alive are enabled
		IdleTimeout: 3 * time.Minute,
	}
	err = server.ListenAndServe()

	if err != nil {
		log.Error(err, "failed to listen on 8080")
		os.Exit(1)
	}
}
