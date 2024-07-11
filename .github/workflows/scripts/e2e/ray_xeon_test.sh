#!/bin/bash
# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

set -exo pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

function install_kuberay_and_start_ray() {
    # Install KubeRay
    ./install-kuberay.sh

    # Start Ray Cluster with Autoscaling
    ./start-ray-cluster.sh
}

function validate_ray() {
    echo "Install KubeRay and Start Ray Cluster with Autoscaling"

    install_kuberay_and_start_ray

    # Wait for ray cluster to be ready
    sleep 20

    # Check if kuberay-operator is ready
    kubectl get pods | grep "kuberay-operator" | grep "Running"

    # Check if raycluster-autoscaler-head is ready
    kubectl get pods | grep "raycluster-autoscaler-head" | grep "Running"

    # Port forward to allow local tests
    kubectl port-forward services/raycluster-autoscaler-head-svc 10001:10001 8265:8265 6379:6379 8080:8080 -n default &

    # Run ray test
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
        popd
        ;;
    *)
        echo "Unknown function: $1"
        ;;
esac

cleanup
