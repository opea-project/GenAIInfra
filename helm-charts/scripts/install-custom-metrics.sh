#!/bin/sh
#
# Copyright (C) 2025 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

set -e

# make sure curl access goes to local cluster
unset http_proxy HTTP_PROXY

# what component data is being manipulated
component=Prometheus-adapter

# label selector for it
prom_selector=app.kubernetes.io/name=prometheus-adapter

# Helm generated config map for custom metrics
rules_ns=default
rules_cm=custom-metrics

# where to save it
conf_file=adapter-config.yaml

usage ()
{
	name=${0#*/}
	cat <<EOF
Usage:
	$name <Prometheus namespace> <chart release>

Install custom metric rules for Prometheus-adapter from '<release>-$rules_cm' configMap
(created to '$rules_ns' namespace). Backup of new rules will be in '$conf_file', and of
old ones in '$conf_file.bak'.

Examples:
	ERAG: $name observability hpa
	OPEA: $name monitoring chatqna

ERROR: $1!
EOF
	exit 1
}

if [ -z "$(which kubectl)" ]; then
	usage "please install 'kubectl' to query Kubernetes"
fi

# ------- setup --------

if [ $# -ne 2 ]; then
	usage "incorrect number of arguments"
fi
prom_ns="$1"
release="$2"
rules_cm="$release-$rules_cm"

echo "Saving current $component custom metric rules configMap to: ${conf_file}.bak"
prom_cm=$(kubectl -n "$prom_ns" get cm --selector $prom_selector -o name | cut -d/ -f2)
kubectl -n "$prom_ns" get "cm/$prom_cm" -o yaml > ${conf_file}.bak

echo "Saving Helm generated '$rules_cm' configMap to: ${conf_file}"
kubectl -n $rules_ns get "cm/$rules_cm" -o yaml | sed \
  -e "s/name: $rules_cm\$/name: $prom_cm/" \
  -e "s/namespace: $rules_ns\$/namespace: $prom_ns/" \
> ${conf_file}

echo "Replacing $component configMap with that"
kubectl delete -n "$prom_ns" "cm/$prom_cm"
kubectl apply -f ${conf_file}

echo "Deleting $component pod to restart it / apply new rules"
prom_pod=$(kubectl -n "$prom_ns" get pod --selector $prom_selector -o name)
kubectl -n "$prom_ns" delete "$prom_pod"

echo "=> custom metric rules installed!"
echo
echo "To track the results, you can use:"
echo "  scale-monitor-helm.sh <chart namespace> $release $prom_ns <Prometheus service>"
