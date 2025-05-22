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

If Prometheus Helm release name is given, Prometheus monitoring is
enabled for the inference engine instances, and vLLM dashboard
(configMap) is installed for Grafana.

ERROR: $1!
EOF
    exit 1
}

metrics=""
for arg in "$@"; do
	if [ -f "$arg" ]; then
		echo "Using HF token from '$arg' file."
		HF_TOKEN=$(cat "$arg")
	else
		echo "Enabling vLLM inference pod monitoring for '$arg' Prometheus Helm install."
		metrics="--set metrics.prometheusOperator.vLLMPodMonitor.labels.release=$arg"
		metrics="$metrics --set metrics.prometheusOperator.vLLMPodMonitor.enabled=true"
	fi
done

if [ -z "$HF_TOKEN" ]; then
	error_exit "HF token missing"
fi

helm upgrade --install opea -n kubeai kubeai/kubeai \
    --create-namespace \
    --set secrets.huggingface.token="$HF_TOKEN" \
    -f $DIR/opea-values.yaml $metrics
