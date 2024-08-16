/*
* Copyright (C) 2024 Intel Corporation
* SPDX-License-Identifier: Apache-2.0
 */

package v1alpha3

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	rt "k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var (
	k8sClient client.Client
	testEnv   *envtest.Environment
	schem     *rt.Scheme
)

func TestMain(m *testing.M) {
	schem = rt.NewScheme()
	utilruntime.Must(AddToScheme(schem))

	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,

		BinaryAssetsDirectory: filepath.Join("..", "..", "bin", "k8s",
			fmt.Sprintf("1.29.0-%s-%s", runtime.GOOS, runtime.GOARCH)),
	}

	cfg, err := testEnv.Start()
	if err != nil {
		panic(err)
	}
	defer func() {
		// Stop the test environment
		stopErr := testEnv.Stop()
		if stopErr != nil {
			panic(stopErr)
		}
	}()

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		panic(err)
	}

	code := m.Run()
	os.Exit(code)
}

func TestGMConnector_SetupWebhookWithManager(t *testing.T) {
	type args struct {
		mgr ctrl.Manager
	}
	port := 9440
	host := "localhost"

	m, err := ctrl.NewManager(testEnv.Config,
		ctrl.Options{
			Scheme:        schem,
			WebhookServer: webhook.NewServer(webhook.Options{Port: port, Host: host}),
		})
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				mgr: m,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := (&GMConnector{}).SetupWebhookWithManager(tt.args.mgr); (err != nil) != tt.wantErr {
				t.Errorf("GMConnector.SetupWebhookWithManager() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_checkStepName(t *testing.T) {
	testNode := "test-node"
	type args struct {
		s        Step
		fldRoot  *field.Path
		nodeName string
	}
	tests := []struct {
		name string
		args args
		want *field.Error
	}{
		{
			name: "empty step name",
			args: args{
				fldRoot:  field.NewPath("spec").Child("nodes"),
				s:        Step{},
				nodeName: testNode,
			},
			want: field.Invalid(
				field.NewPath("spec").Child("nodes").Child("test-node").Child("stepName"),
				Step{},
				fmt.Sprintf("the step name for node %v cannot be empty", testNode),
			),
		},
		{
			name: "invalid step name",
			args: args{
				fldRoot: field.NewPath("spec").Child("nodes"),
				s: Step{
					StepName: "invalid",
					Executor: Executor{},
				},
				nodeName: testNode,
			},
			want: field.Invalid(
				field.NewPath("spec").Child("nodes").Child("test-node").Child("stepName"),
				Step{
					StepName: "invalid",
					Executor: Executor{},
				},
				fmt.Sprintf("invalid step name: %s for node %v", "invalid", testNode),
			),
		},
		{
			name: "success",
			args: args{
				fldRoot: field.NewPath("spec").Child("nodes"),
				s: Step{
					StepName: "Embedding",
					Executor: Executor{},
				},
				nodeName: testNode,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkStepName(tt.args.s, tt.args.fldRoot, tt.args.nodeName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("checkStepName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_nodeNameExists(t *testing.T) {
	type args struct {
		name  string
		nodes []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "nodeName is unset",
			args: args{
				name:  "",
				nodes: []string{"root", "test-node"},
			},
			want: true,
		},
		{
			name: "unknown nodeName",
			args: args{
				name:  "unknown",
				nodes: []string{"root", "test-node"},
			},
			want: false,
		},
		{
			name: "existing nodeName",
			args: args{
				name:  "root",
				nodes: []string{"root", "test-node"},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nodeNameExists(tt.args.name, tt.args.nodes); got != tt.want {
				t.Errorf("nodeNameExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateRootExistance(t *testing.T) {
	type args struct {
		nodes   map[string]Router
		fldPath *field.Path
	}
	tests := []struct {
		name string
		args args
		want *field.Error
	}{
		{
			name: "root node exists",
			args: args{
				nodes: map[string]Router{
					"root": {},
				},
				fldPath: field.NewPath("spec").Child("nodes"),
			},
			want: nil,
		},
		{
			name: "root node does not exist",
			args: args{
				nodes:   map[string]Router{},
				fldPath: field.NewPath("spec").Child("nodes"),
			},
			want: field.Invalid(field.NewPath("spec").Child("nodes"),
				map[string]Router{},
				"a root node is required"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateRootExistance(tt.args.nodes, tt.args.fldPath); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("validateRootExistance() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getKeys(t *testing.T) {
	type args struct {
		m map[string]Router
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "succes",
			args: args{
				m: map[string]Router{
					"root": {},
				},
			},
			want: []string{"root"},
		},
		{
			name: "empty map",
			args: args{
				m: map[string]Router{},
			},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getKeys(tt.args.m); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateNames(t *testing.T) {
	var errs field.ErrorList
	type args struct {
		nodes   map[string]Router
		fldPath *field.Path
	}
	tests := []struct {
		name string
		args args
		want field.ErrorList
	}{
		{
			name: "duplicate service name",
			args: args{
				nodes: map[string]Router{
					"root": {
						Steps: []Step{
							{
								StepName: "Embedding",
								Executor: Executor{
									InternalService: GMCTarget{
										ServiceName: "embedding_svc",
									},
								},
							},
							{
								StepName: "Embedding",
								Executor: Executor{
									InternalService: GMCTarget{
										ServiceName: "embedding_svc",
									},
								},
							},
						},
					},
				},
				fldPath: field.NewPath("spec").Child("nodes"),
			},
			want: append(errs, field.Invalid(
				field.NewPath("spec").Child("nodes").Child("root").Child("internalService").Child("serviceName"),
				Step{
					StepName: "Embedding",
					Executor: Executor{
						InternalService: GMCTarget{
							ServiceName: "embedding_svc",
						},
					},
				},
				"service name: embedding_svc in node root already exists")),
		},
		{
			name: "no error",
			args: args{
				nodes: map[string]Router{
					"root": {
						Steps: []Step{
							{
								StepName: "Embedding",
								Executor: Executor{
									InternalService: GMCTarget{
										ServiceName: "embedding_svc1",
									},
								},
							},
							{
								StepName: "Embedding",
								Executor: Executor{
									InternalService: GMCTarget{
										ServiceName: "embedding_svc2",
									},
								},
							},
						},
					},
				},
				fldPath: field.NewPath("spec").Child("nodes"),
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateNames(tt.args.nodes, tt.args.fldPath); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("validateNames() = %v, want %v", got, tt.want)
			}
		})
	}
}
