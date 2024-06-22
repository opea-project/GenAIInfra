/*
* Copyright (C) 2024 Intel Corporation
* SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"context"
	"fmt"
	"time"

	mcv1alpha3 "github.com/opea-project/GenAIInfra/microservices-connector/api/v1alpha3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type GMC_monitor struct {
	client.Client
	Scheme *runtime.Scheme

	// ResourceStatus is a map of service name to its status
	ResourceStatus map[NsName](MonitorCategory)
}
type MonitorCategory struct {
	Graph         NsName
	resources     []ServiceAndDeploy
	readyCount    int
	externalCount int
	totalCount    int
}

type NsName struct {
	Namespace string
	Name      string
}

type ServiceAndDeploy struct {
	srvc NsName
	dply NsName
}

func (m GMC_monitor) Start(ch <-chan MonitorCategory, stopCh <-chan struct{}) {
	go func() {
		fmt.Println("Monitor task initiated.")

		for {
			select {
			case mc := <-ch:
				fmt.Printf(" received monitor object: %v \n", mc)
				// Process the MonitorCategory
				m.ResourceStatus[mc.Graph] = mc
			case <-stopCh:
				// Exit the goroutine when stop signal is received
				return
			}

		}
	}()

	// Start polling in a separate goroutine
	go func() {
		ticker := time.NewTicker(30 * time.Second) // Adjust the interval as needed
		defer ticker.Stop()
		fmt.Println("Monitor task starts to work.")

		for {
			select {
			case <-ticker.C:
				// Loop through every item in ServiceStatus and check their status
				for graph, record := range m.ResourceStatus {
					readyCnt := 0
					for _, svc := range record.resources {
						if m.getStatus(svc) {
							readyCnt += 1
						}
					}

					if readyCnt != record.readyCount {
						fmt.Printf("%v status changed from %d to %d\n", graph, record.readyCount, readyCnt)
						// Update the MonitorCategory with the new count
						// This is a placeholder. Replace with actual logic to update count.
						record.readyCount = readyCnt
						if m.updateGMCstatus(graph, record.readyCount, record.externalCount, record.totalCount) {
							m.ResourceStatus[graph] = record
						}

					}
				}
			case <-stopCh:
				// Exit the goroutine when stop signal is received
				return
			}
		}
	}()
}

func (m GMC_monitor) getStatus(resource ServiceAndDeploy) bool {
	// Implement the logic to get the status of the service.
	// Return true if the service is active, false otherwise.
	service := &corev1.Service{}
	err := m.Client.Get(context.Background(), client.ObjectKey{Namespace: resource.srvc.Namespace, Name: resource.srvc.Name}, service)
	if err != nil {
		fmt.Printf("Failed to get service %s@%s: %v\n", resource.srvc.Name, resource.srvc.Namespace, err)
		return false
	}
	deployment := &appsv1.Deployment{}
	err = m.Client.Get(context.Background(), client.ObjectKey{Namespace: resource.dply.Namespace, Name: resource.dply.Name}, deployment)
	if err != nil {
		fmt.Printf("Failed to get deployment %s@%s: %v\n", resource.dply.Name, resource.dply.Namespace, err)
		return false
	}
	if GetServiceURL(service) != "" && (deployment.Status.ReadyReplicas == *deployment.Spec.Replicas) {
		return true
	} else {
		return false
	}
}

func (m GMC_monitor) updateGMCstatus(gmc NsName, readyCount, externalCount, totalCount int) bool {
	graph := &mcv1alpha3.GMConnector{}
	graph.Name = gmc.Name
	graph.Namespace = gmc.Namespace

	latest := &unstructured.Unstructured{}
	latest.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "gmc.opea.io", // Replace with the actual group of GMConnector
		Version: "v1alpha3",    // Use the correct version
		Kind:    "GMConnector", // Ensure this matches the Kind of the GMConnector
	})

	if err := m.Client.Get(context.Background(), client.ObjectKeyFromObject(graph), latest); err != nil {
		if apierr.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			fmt.Printf("Failed to get GMC %s@%s: %v\n", gmc.Name, gmc.Namespace, err)
		} else {
			fmt.Printf("get GMC %s@%s error : %v\n", gmc.Name, gmc.Namespace, err)
		}
	} else {
		statusObj, ok := latest.Object["status"].(map[string]interface{})
		if !ok {
			fmt.Println("Failed to type assert latest.Object['status'] to map[string]interface{}")
		}
		statusObj["status"] = fmt.Sprintf("%d/%d/%d", readyCount, externalCount, totalCount)
		if err := m.Client.Status().Update(context.Background(), latest); err != nil {
			fmt.Printf("Failed to update GMC status to %s: %v\n", latest.Object["status"], err)
		} else {
			fmt.Printf("Updated GMC %s@%s status to %s\n", gmc.Name, gmc.Namespace, latest.Object["status"])
			return true
		}
	}
	return false
}
