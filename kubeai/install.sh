#!/bin/bash

# Copyright (C) 2025 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

error_exit () {
	name=${0##*/}
	cat <<EOF

Usage: [HF_TOKEN=<token>] $name [HF token file] [Prometheus release name]

HuggingFace token for accessing models can be given either as an
environment variable, or as a name of a file containing the token
value.

If script finds deployed "kube-prometheus-stack" Helm chart release,
it enables Prometheus monitoring for the vLLM inference engine pods.
vLLM dashboards (configMaps) are installed if also a running Grafana
instance is detected.

Prometheus release name can also be given as an argument,

ERROR: $1!
EOF
    exit 1
}

release=""
for arg in "$@"; do
	if [ -f "$arg" ]; then
		echo "Using HF token from '$arg' file."
		HF_TOKEN=$(cat "$arg")
	else
		release="$arg"
	fi
done

if [ -z "$HF_TOKEN" ]; then
	error_exit "HF token missing"
fi

if [ -z "$release" ]; then
	if [ -z "$(which jq)" ]; then
		error_exit "please install 'jq' to parse Helm releases info"
	fi
	# check whether cluster has deployed Prometheus Helm chart release, if none specified
	release=$(helm list -A -o json | jq '.[] | select(.chart|match("^kube-prometheus-stack")) | select(.status=="deployed") | .name' | tr -d '"')
fi

metrics=""
if [ -n "$release" ]; then
	running="status.phase=Running"
	grafana="app.kubernetes.io/name=grafana"
	jsonpath="{.items[0].metadata.namespace}"

	# check for Grafana namespace
	ns=$(kubectl get -A pod --field-selector="$running" --selector="$grafana" -o jsonpath="$jsonpath")
	if [ -n "$ns" ]; then
		echo "Grafana available, installing vLLM dashboards to '$ns' namespace."
		kubectl apply -n $ns -f $DIR/grafana/vllm-scaling.yaml -f $DIR/grafana/vllm-details.yaml
	fi

	echo "Enabling vLLM pod monitoring for '$release' Prometheus Helm install."
	metrics="--set metrics.prometheusOperator.vLLMPodMonitor.labels.release=$release"
	metrics="$metrics --set metrics.prometheusOperator.vLLMPodMonitor.enabled=true"
fi

helm upgrade --install opea-kubeai -n kubeai kubeai/kubeai \
    --create-namespace \
    --set secrets.huggingface.token="$HF_TOKEN" \
    -f $DIR/opea-values.yaml $metrics
