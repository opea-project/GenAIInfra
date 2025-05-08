#!/bin/sh
#
# Copyright (C) 2025 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

set -e

# make sure curl access goes to local cluster
unset http_proxy HTTP_PROXY

usage ()
{
	name=${0#*/}
	cat <<EOF
Monitor Helm chart services scaling.

Usage:
	$name [options] <chart namespace> <chart release name> [<Prometheus namespace> <Prometheus service>]

Options:
	--pods -- debug output to list chart pods that are already providing metrics

Examples:
	watch -n 5 $name default chatqna
	watch -n 5 $name default hpa observability observability-kube-prom-prometheus

(First example assumes OPEA defaults for Prometheus: monitoring / prometheus-stack-kube-prom-prometheus.)

ERROR: $1!
EOF
	exit 1
}

if [ -z "$(which kubectl)" ]; then
	usage "please install 'kubectl' to query Kubernetes"
fi
if [ -z "$(which jq)" ]; then
	usage "please install 'jq' to process custom metrics output"
fi

# ------- setup --------

show_pods="false"
if [ "$1" = "--pods" ]; then
	show_pods="true"
	shift
fi

if [ $# -eq 4 ]; then
	prom_ns=$3
	prom_svc=$4
elif [ $# -eq 2 ]; then
	prom_ns=monitoring
	prom_svc=prometheus-stack-kube-prom-prometheus
else
	usage "incorrect number of arguments"
fi
scale_ns=$1
chart=$2

set +e
prom_url=http://$(kubectl -n "$prom_ns" get -o jsonpath="{.spec.clusterIP}:{.spec.ports[0].port}" "svc/$prom_svc")
# shellcheck disable=SC2181
if [ $? -ne 0 ]; then
	echo
	echo "Services available in '$prom_ns' namespace are:"
	kubectl -n "$prom_ns" get svc
	echo
	echo "ERROR: invalid Prometheus service name (or namespace)!  See above --^"
	exit 1
fi
set -e

custom_api="/apis/custom.metrics.k8s.io/v1beta1"

# -------- outputs --------

echo "HPA scaling rules for '$scale_ns' namespace '$chart' targets:"
kubectl get -n "$scale_ns" hpa | grep -e NAME -e "$chart" || true

set +e
names=$(kubectl get -n "$scale_ns" hpa -o name | cut -d/ -f2 | grep "$chart")
set -e

if [ -n "$names" ]; then
	echo
	echo "Corresponding deployment Ready counts:"
	# shellcheck disable=SC2086
	kubectl get -n "$scale_ns" deployments $names || true
fi

echo
echo "Prometheus scrape target counts for '$chart' chart:"
#echo "curl --no-progress-meter $prom_url/metrics | grep scrape_pool_targets.*$chart"
curl --no-progress-meter "$prom_url/metrics" | grep "scrape_pool_targets.*/$scale_ns/$chart"

echo
echo "Custom metrics and their values, available for HPA scaling:"
for metric in $(kubectl get --raw $custom_api | jq .resources[].name | cut -d/ -f2 | sort -u | tr -d '"'); do
	printf "* %s: " "$metric"
	kubectl get --raw "$custom_api/namespaces/$scale_ns/metrics/$metric" | jq .items[].value
	# kubectl get --raw "$custom_api/namespaces/$scale_ns/service/*/$metric" | jq .items[].value
done

# --- metric debugging outputs ---

if [ "$show_pods" != "true" ]; then
	exit 0
fi

echo
echo "Metrics scrape target pods:"
curl --no-progress-meter "$prom_url/api/v1/query?" --data-urlencode \
  'query={namespace="'"$scale_ns"'",pod=~"'"$chart"'-.*"}' \
  | jq .data.result[].metric.pod | sort -u
