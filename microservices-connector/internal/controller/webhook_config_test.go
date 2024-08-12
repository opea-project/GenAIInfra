/*
* Copyright (C) 2024 Intel Corporation
* SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"bytes"
	"context"
	"testing"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCreateOrUpdateValidatingWebhookConfiguration(t *testing.T) {
	client := fake.NewSimpleClientset()
	vClient := client.AdmissionregistrationV1()
	type args struct {
		caPEM            *bytes.Buffer
		port             int32
		webhookService   string
		webhookNamespace string
		isUpdate         bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "create a new webhook config",
			args: args{
				caPEM:            bytes.NewBufferString("test"),
				port:             9443,
				webhookService:   "test-service",
				webhookNamespace: "default",
				isUpdate:         false,
			},
			wantErr: false,
		},
		{
			name: "update a webhook config",
			args: args{
				caPEM:            bytes.NewBufferString("test"),
				port:             9443,
				webhookService:   "test-service",
				webhookNamespace: "default",
				isUpdate:         true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// if tt.args.isUpdate {
			// 	sideEffect := admissionregistrationv1.SideEffectClassNone
			// 	fail := admissionregistrationv1.Fail
			// 	validatingWHConfig := &admissionregistrationv1.ValidatingWebhookConfiguration{
			// 		ObjectMeta: metav1.ObjectMeta{
			// 			Name: webhookConfigName,
			// 		},
			// 		Webhooks: []admissionregistrationv1.ValidatingWebhook{{
			// 			Name: fmt.Sprintf("v%s.%s", resource, apiGroup),
			// 			ClientConfig: admissionregistrationv1.WebhookClientConfig{
			// 				CABundle: tt.args.caPEM.Bytes(), // self-generated CA for the webhook
			// 				Service: &admissionregistrationv1.ServiceReference{
			// 					Name:      tt.args.webhookService,
			// 					Namespace: tt.args.webhookNamespace,
			// 					Path:      &validatingPath,
			// 					Port:      &tt.args.port,
			// 				},
			// 			},
			// 			AdmissionReviewVersions: []string{"v1"},
			// 			SideEffects:             &sideEffect,
			// 			Rules: []admissionregistrationv1.RuleWithOperations{
			// 				{
			// 					Operations: []admissionregistrationv1.OperationType{
			// 						admissionregistrationv1.Create,
			// 						admissionregistrationv1.Update,
			// 					},
			// 					Rule: admissionregistrationv1.Rule{
			// 						APIGroups:   []string{apiGroup},
			// 						APIVersions: []string{apiVersion},
			// 						Resources:   []string{fmt.Sprintf("%ss", resource)},
			// 					},
			// 				},
			// 			},
			// 			FailurePolicy: &fail,
			// 		}},
			// 	}
			// 	if _, err := vClient.ValidatingWebhookConfigurations().Create(context.TODO(), validatingWHConfig, metav1.CreateOptions{}); err != nil {
			// 		t.Errorf("failed to create the validatingWebhookConfiguration(%s): %v", webhookConfigName, err)
			// 		return
			// 	}
			// }
			if err := CreateOrUpdateValidatingWebhookConfiguration(
				client,
				tt.args.caPEM,
				tt.args.port,
				tt.args.webhookService,
				tt.args.webhookNamespace); (err != nil) != tt.wantErr {
				t.Errorf("CreateOrUpdateValidatingWebhookConfiguration() error = %v, wantErr %v", err, tt.wantErr)
			}
			_, err := vClient.ValidatingWebhookConfigurations().Get(context.TODO(),
				webhookConfigName,
				metav1.GetOptions{})
			if err != nil && apierrors.IsNotFound(err) {
				t.Errorf("Webhook %s is not found", webhookConfigName)
			}
		})
	}
}
