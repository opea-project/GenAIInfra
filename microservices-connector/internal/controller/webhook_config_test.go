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
