#!/bin/bash
# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

set -exo pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

function install_kuberay_and_start_ray_cluster() {
    # Install KubeRay
    ./install-kuberay.sh

    # Start Ray Cluster with Autoscaling
    ./start-ray-cluster.sh
}

function setup_client_side() {
    # Install Ray client
    pip install ray[client]==2.23.0

    echo "Current Python version: $(python --version | awk '{print $2}')"
    echo "Current Ray version: $(ray --version | awk '{print $3}')"
}

function wait_until_ray_head_pod_ready() {
    max_retries=10
    retry_count=0

    while true; do
        # Get the name of the Ray head pod, the name is like `raycluster-autoscaler-head-5snzb`
        pod_name=$(kubectl get pods --selector=ray.io/node-type=head -o custom-columns=POD:metadata.name --no-headers)

        # Check if the pod name was found
        if [ -z "$pod_name" ]; then
            pod_status = "False"
        else
            # Get the status of the Ray head pod
            pod_status=$(kubectl get pod $pod_name -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}')
        fi

        if [ "$pod_status" == "True" ]; then
            echo "Ray head pod is ready."
            break
        elif [ $retry_count -ge $max_retries ]; then
            echo "Ray head pod is not ready after waiting for a significant amount of time."
            exit 1
        else
            echo "Ray head pod is not ready yet. Retrying in 10 seconds..."
            sleep 10
            retry_count=$((retry_count + 1))
        fi
    done
}

function wait_until_kuberay_pod_ready() {
    max_retries=10
    retry_count=0

    while true; do
        # Get the name of the KubeRay pod, the name is like `kuberay-operator-7f85d8578-srj4c`
        pod_name=$(kubectl get pods --selector=app.kubernetes.io/name=kuberay-operator -o custom-columns=POD:metadata.name --no-headers)

        if [ -z "$pod_name" ]; then
            kuberay_operator_status = "False"
        else
            kuberay_operator_status=$(kubectl get pod $pod_name -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}')
        fi

        if [ "$kuberay_operator_status" == "True" ]; then
            echo "KubeRay operator is ready."
            break
        elif [ $retry_count -ge $max_retries ]; then
            echo "KubeRay operator is not ready after waiting for a significant amount of time."
            exit 1
        else
            echo "KubeRay operator is not ready yet. Retrying in 10 seconds..."
            sleep 10
            retry_count=$((retry_count + 1))
        fi
    done
}

function validate_ray() {
    echo "Install KubeRay and Start Ray Cluster with Autoscaling"
    install_kuberay_and_start_ray_cluster

    echo "Waiting for the KubeRay pod to be ready..."
    wait_until_kuberay_pod_ready

    echo "Waiting for the Ray head pod to be ready..."
    wait_until_ray_head_pod_ready

    echo "Port forward to allow local tests"
    kubectl port-forward services/raycluster-autoscaler-head-svc 10001:10001 8265:8265 6379:6379 8080:8080 &

    echo "Setup client-side environments"
    setup_client_side

    echo "Run basic Ray app test"
    python ray-test.py
}

function cleanup() {
    # Delete Ray Cluster
    ./delete-ray-cluster.sh

    # Uninstall KuebRay
    ./uninstall-kuberay.sh
}

if [ $# -eq 0 ]; then
    echo "Usage: $0 <function_name>"
    exit 1
fi

case "$1" in
    validate_ray)
        pushd scripts/ray
        validate_ray
        cleanup
        popd
        ;;
    *)
        echo "Unknown function: $1"
        ;;
esac
