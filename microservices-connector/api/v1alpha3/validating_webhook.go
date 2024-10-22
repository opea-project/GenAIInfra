/*
* Copyright (C) 2024 Intel Corporation
* SPDX-License-Identifier: Apache-2.0
 */

package v1alpha3

import (
	"fmt"
	"slices"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:docs-gen:collapse=Go imports

var (
	//setup a logger for the webhooks.
	vlog      = logf.Log.WithName("validating-webhook")
	stepNames = []string{
		"TeiEmbedding",
		"TeiEmbeddingGaudi",
		"Embedding",
		"VectorDB",
		"Retriever",
		"Reranking",
		"TeiReranking",
		"Tgi",
		"TgiGaudi",
		"TgiNvidia",
		"Llm",
		"DocSum",
		"Router",
		"WebRetriever",
		"Asr",
		"Tts",
		"SpeechT5",
		"SpeechT5Gaudi",
		"Whisper",
		"WhisperGaudi",
		"DataPrep",
		"UI",
	}
)

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (r *GMConnector) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-gmc-opea-io-v1alpha3-gmconnector,mutating=false,failurePolicy=fail,groups=gmc.opea.io,resources=gmconnectors,versions=v1alpha3,name=vgmcconnector.gmc.opea.io,sideEffects=None,admissionReviewVersions=v1

var _ webhook.Validator = &GMConnector{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *GMConnector) ValidateCreate() (admission.Warnings, error) {
	vlog.Info("validate create", "name", r.Name)

	return nil, r.validateGMConnector()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *GMConnector) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	vlog.Info("validate update", "name", r.Name)

	return nil, r.validateGMConnector()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *GMConnector) ValidateDelete() (admission.Warnings, error) {
	vlog.Info("validate delete", "name", r.Name)
	return nil, nil
}

/*
validate the name and the spec of the GMConnector.
*/
func (r *GMConnector) validateGMConnector() error {
	if err := r.checkfields(); err != nil {
		return apierrors.NewInvalid(
			schema.GroupKind{Group: GroupVersion.Group, Kind: "GMCConnector"},
			r.Name, err)
	}

	return nil

}

func (r *GMConnector) checkfields() field.ErrorList {
	// The field helpers from the kubernetes API machinery help us return nicely
	// structured validation errors.
	var allErrs field.ErrorList
	if errs := validateNames(r.Spec.Nodes, field.NewPath("spec").Child("nodes")); len(errs) > 0 {
		allErrs = errs
	}
	if err := validateRootExistance(r.Spec.Nodes, field.NewPath("spec").Child("nodes")); err != nil {
		allErrs = append(allErrs, err)
	}

	if len(allErrs) == 0 {
		return nil
	}
	return allErrs
}

func checkStepName(s Step, idx int, fldRoot *field.Path, nodeName string) *field.Error {
	if len(s.StepName) == 0 {
		return field.Invalid(fldRoot.Child(nodeName).Child(fmt.Sprintf("steps[%d]", idx)).Child("name"),
			s,
			fmt.Sprintf("the step name for node %v cannot be empty", nodeName))
	}
	if !slices.Contains(stepNames, s.StepName) {
		return field.Invalid(fldRoot.Child(nodeName).Child(fmt.Sprintf("steps[%d]", idx)).Child("name"),
			s,
			fmt.Sprintf("invalid step name: %s for node %v", s.StepName, nodeName))
	}
	return nil
}

func nodeNameExists(name string, nodes []string) bool {
	// node name is not set, skip check
	if len(name) == 0 {
		return true
	}
	return slices.Contains(nodes, name)
}

func getKeys(m map[string]Router) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// validate step name and node name
func validateNames(nodes map[string]Router, fldPath *field.Path) field.ErrorList {
	nodeNames := getKeys(nodes)
	serviceNames := []string{}
	var errs field.ErrorList

	for name, router := range nodes {
		for idx, step := range router.Steps {
			// validate step name
			if err := checkStepName(step, idx, fldPath, name); err != nil {
				errs = append(errs, err)
			}

			// check node name has been defined in the spec
			if !nodeNameExists(step.NodeName, nodeNames) {
				errs = append(errs, field.Invalid(fldPath.Child(name).Child(fmt.Sprintf("steps[%d]", idx)).Child("nodeName"),
					step,
					fmt.Sprintf("node name: %v in step %v does not exist", step.NodeName, step.StepName)))
			}

			// check service name uniqueness
			if len(step.InternalService.ServiceName) != 0 && slices.Contains(serviceNames, step.InternalService.ServiceName) {
				errs = append(errs, field.Invalid(fldPath.Child(name).Child(fmt.Sprintf("steps[%d]", idx)).Child("internalService").Child("serviceName"),
					step,
					fmt.Sprintf("service name: %v in node %v already exists", step.InternalService.ServiceName, name)))
			} else {
				serviceNames = append(serviceNames, step.InternalService.ServiceName)
			}
		}
	}
	return errs
}

// check root node exists
func validateRootExistance(nodes map[string]Router, fldPath *field.Path) *field.Error {
	if _, ok := nodes["root"]; !ok {
		return field.Invalid(fldPath, nodes, "a root node is required")
	}
	return nil
}

// +kubebuilder:docs-gen:collapse=Existing Validation
