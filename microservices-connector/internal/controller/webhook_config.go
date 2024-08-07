/*
* Copyright (C) 2024 Intel Corporation
* SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	apiGroup          = "gmc.opea.io"
	apiVersion        = "v1alpha3"
	resource          = "gmconnector"
	webhookConfigName = "validating-webhook-configuration"
)

var (
	logw           = logf.Log.WithName("WebhookConfig")
	validatingPath = fmt.Sprintf("/validate-%s-%s-%s", strings.Replace(apiGroup, ".", "-", 2), apiVersion, resource)
)

func CreateOrUpdateMutatingWebhookConfiguration(caPEM *bytes.Buffer, port int32, webhookService, webhookNamespace string) error {
	// Initializing the kube client
	kubeconfig := os.Getenv("KUBECONFIG")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	validatingWebhookConfigV1Client := clientset.AdmissionregistrationV1()

	logw.Info("Creating or updating the validatingwebhookconfiguration", "webhook name", webhookConfigName)
	sideEffect := admissionregistrationv1.SideEffectClassNone
	fail := admissionregistrationv1.Fail
	validatingWHConfig := &admissionregistrationv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: webhookConfigName,
		},
		Webhooks: []admissionregistrationv1.ValidatingWebhook{{
			Name: fmt.Sprintf("v%s.%s", resource, apiGroup),
			ClientConfig: admissionregistrationv1.WebhookClientConfig{
				CABundle: caPEM.Bytes(), // self-generated CA for the webhook
				Service: &admissionregistrationv1.ServiceReference{
					Name:      webhookService,
					Namespace: webhookNamespace,
					Path:      &validatingPath,
					Port:      &port,
				},
			},
			AdmissionReviewVersions: []string{"v1"},
			SideEffects:             &sideEffect,
			Rules: []admissionregistrationv1.RuleWithOperations{
				{
					Operations: []admissionregistrationv1.OperationType{
						admissionregistrationv1.Create,
						admissionregistrationv1.Update,
					},
					Rule: admissionregistrationv1.Rule{
						APIGroups:   []string{apiGroup},
						APIVersions: []string{apiVersion},
						Resources:   []string{fmt.Sprintf("%ss", resource)},
					},
				},
			},
			FailurePolicy: &fail,
		}},
	}

	foundWebhookConfig, err := validatingWebhookConfigV1Client.ValidatingWebhookConfigurations().Get(context.TODO(), webhookConfigName, metav1.GetOptions{})
	if err != nil && apierrors.IsNotFound(err) {
		if _, err := validatingWebhookConfigV1Client.ValidatingWebhookConfigurations().Create(context.TODO(), validatingWHConfig, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("failed to create the validatingWebhookConfiguration(%s): %v", webhookConfigName, err)
		}
		logw.Info("Created validatingWebhookConfiguration", "webhookConfigName", webhookConfigName)
	} else if err != nil {
		return fmt.Errorf("failed to check the validatingWebhookConfiguration(%s): %v", webhookConfigName, err)
	} else {
		// there is an existing validatingWebhookConfiguration
		old := foundWebhookConfig.Webhooks[0]
		new := validatingWHConfig.Webhooks[0]
		if len(foundWebhookConfig.Webhooks) != len(validatingWHConfig.Webhooks) ||
			!(old.Name == new.Name &&
				reflect.DeepEqual(old.AdmissionReviewVersions, new.AdmissionReviewVersions) &&
				reflect.DeepEqual(old.SideEffects, new.SideEffects) &&
				reflect.DeepEqual(old.FailurePolicy, new.FailurePolicy) &&
				reflect.DeepEqual(old.Rules, new.Rules) &&
				reflect.DeepEqual(old.ClientConfig.CABundle, new.ClientConfig.CABundle) &&
				reflect.DeepEqual(old.ClientConfig.Service, new.ClientConfig.Service)) {
			validatingWHConfig.ObjectMeta.ResourceVersion = foundWebhookConfig.ObjectMeta.ResourceVersion
			if _, err := validatingWebhookConfigV1Client.ValidatingWebhookConfigurations().Update(context.TODO(), validatingWHConfig, metav1.UpdateOptions{}); err != nil {
				return fmt.Errorf("failed to update the validatingWebhookConfiguration(%s): %v", webhookConfigName, err)
			}
			logw.Info("Updated the validatingWebhookConfiguration", "webhookConfigName", webhookConfigName)
		} else {
			logw.Info("The validatingWebhookConfiguration already exists and has no change", "webhookConfigName", webhookConfigName)
		}
	}

	return nil
}
