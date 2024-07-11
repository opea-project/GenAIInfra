#!/bin/bash
# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

set -exo pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

function setup_kuberay_and_start_ray() {
    # Install kuberay
    ./install-kuberay.sh

    # Start ray cluster
    ./start-ray-cluster.sh
}

function validate_ray() {
    echo "setup kuberay and start ray cluster"
    setup_kuberay_and_start_ray

    # Wait for ray cluster to be ready
    sleep 20

    # Check if kuberay-operator is ready
    kubectl get pods | grep "kuberay-operator" | grep "Running"

    # Check if raycluster-autoscaler-head is ready
    kubectl get pods | grep "raycluster-autoscaler-head" | grep "Running"

    # Run ray test
    python ray-test.py
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
