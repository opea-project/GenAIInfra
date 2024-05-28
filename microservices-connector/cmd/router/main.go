/*
* Copyright (C) 2024 Intel Corporation
* SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	// "regexp"
	"strconv"
	// "strings"
	"time"

	"github.com/tidwall/gjson"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	// "crypto/rand"
	// "math/big"

	gmcv1alpha3 "github.com/opea-project/GenAIInfra/genai-microservices-connector/api/v1alpha3"
	flag "github.com/spf13/pflag"
)

var (
	jsonGraph       = flag.String("graph-json", "", "serialized json graph def")
	log             = logf.Log.WithName("GMCGraphRouter")
	gmcGraph        *gmcv1alpha3.GMConnector
	defaultNodeName = "root"
)

const (
	ServiceURL  = "serviceUrl"
	ServiceNode = "node"
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

func pickupRouteByCondition(input []byte, routes []gmcv1alpha3.Step) *gmcv1alpha3.Step {
	if !gjson.ValidBytes(input) {
		return nil
	}
	for _, route := range routes {
		if gjson.GetBytes(input, route.Condition).Exists() {
			return &route
		}
	}
	return nil
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

func callService(step *gmcv1alpha3.Step, serviceUrl string, input []byte, headers http.Header) ([]byte, int, error) {
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
func getServiceURLByStepTarget(step *gmcv1alpha3.Step, svcNameSpace string) string {
	if step.ServiceURL == "" {
		serviceURL := fmt.Sprintf("http://%s.%s.svc.cluster.local", step.StepName, svcNameSpace)
		return serviceURL
	}
	return step.ServiceURL
}

func executeStep(
	step *gmcv1alpha3.Step,
	graph gmcv1alpha3.GMConnector,
	input []byte,
	headers http.Header,
) ([]byte, int, error) {
	if step.NodeName != "" {
		// when nodeName is specified make a recursive call for routing to next step
		return routeStep(step.NodeName, graph, input, headers)
	}
	serviceURL := getServiceURLByStepTarget(step, graph.Namespace)
	return callService(step, serviceURL, input, headers)
}

func handleSwitchNode(
	route *gmcv1alpha3.Step,
	graph gmcv1alpha3.GMConnector,
	input []byte,
	headers http.Header,
) ([]byte, int, error) {
	var statusCode int
	var responseBytes []byte
	var err error
	stepType := ServiceURL
	if route.NodeName != "" {
		stepType = ServiceNode
	}
	log.Info("Starting execution of step", "type", stepType, "stepName", route.StepName)
	if responseBytes, statusCode, err = executeStep(route, graph, input, headers); err != nil {
		return nil, 500, err
	}

	if route.Dependency == gmcv1alpha3.Hard && !isSuccessFul(statusCode) {
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

func routeStep(nodeName string, graph gmcv1alpha3.GMConnector, input []byte, headers http.Header) ([]byte, int, error) {
	defer timeTrack(time.Now(), "node", nodeName)
	currentNode := graph.Spec.Nodes[nodeName]

	if currentNode.RouterType == gmcv1alpha3.Switch {
		var err error
		route := pickupRouteByCondition(input, currentNode.Steps)
		if route == nil {
			errorMessage := "None of the routes matched with the switch condition"
			err = errors.New(errorMessage)
			log.Error(err, errorMessage)
			return nil, 404, err
		}
		return handleSwitchNode(route, graph, input, headers)
	}

	if currentNode.RouterType == gmcv1alpha3.Ensemble {
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
				output, statusCode, err := executeStep(step, graph, input, headers)
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
				if !isSuccessFul(ensembleStepOutput.StepStatusCode) && currentNode.Steps[i].Dependency == gmcv1alpha3.Hard {
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

	if currentNode.RouterType == gmcv1alpha3.Sequence {
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
			if responseBytes, statusCode, err = executeStep(step, graph, request, headers); err != nil {
				return nil, 500, err
			}

			log.Info("Print Response Bytes", "Response Bytes", responseBytes, "Status Code", statusCode)
			/*
			   Only if a step is a hard dependency, we will check for its success.
			*/
			if step.Dependency == gmcv1alpha3.Hard {
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
	log.Error(nil, "invalid route type", "type", currentNode.RouterType)
	return nil, 500, fmt.Errorf("invalid route type: %v", currentNode.RouterType)
}

func gmcGraphHandler(w http.ResponseWriter, req *http.Request) {
	inputBytes, _ := io.ReadAll(req.Body)
	if response, statusCode, err := routeStep(defaultNodeName, *gmcGraph, inputBytes, req.Header); err != nil {
		log.Error(err, "failed to process request")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if _, err := w.Write(prepareErrorResponse(err, "Failed to process request")); err != nil {
			log.Error(err, "failed to write gmcGraphHandler response")
		}
	} else {
		if json.Valid(response) {
			w.Header().Set("Content-Type", "application/json")
		}
		w.WriteHeader(statusCode)
		if _, err := w.Write(response); err != nil {
			log.Error(err, "failed to write gmcGraphHandler response")
		}
	}
}

func main() {
	flag.Parse()
	logf.SetLogger(zap.New())

	gmcGraph = &gmcv1alpha3.GMConnector{}
	err := json.Unmarshal([]byte(*jsonGraph), gmcGraph)
	if err != nil {
		log.Error(err, "failed to unmarshall gmc graph json")
		os.Exit(1)
	}

	http.HandleFunc("/", gmcGraphHandler)

	server := &http.Server{
		// specify the address and port
		Addr: ":8080",
		// specify your HTTP handler
		Handler: http.HandlerFunc(gmcGraphHandler),
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
