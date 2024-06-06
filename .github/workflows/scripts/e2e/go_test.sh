#!/bin/bash
# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

set -xe
USER_ID=$(whoami)
LOG_PATH=/home/$(whoami)/logs
MOUNT_DIR=/home/$USER_ID/charts-mnt
IMAGE_REPO=${OPEA_IMAGE_REPO:-docker.io}

function install_gmc() {
    # Make sure you have to use image tag $VERSION for microservice-connector installation
    echo "install microservice-connector, using repo $DOCKER_REGISTRY and tag $VERSION"
    echo "using namespace $SYSTEM_NAMESPACE and $APP_NAMESPACE"
}

function validate_gmc() {
    echo "validate microservice-connector"
}

function cleanup_gmc() {
    echo "clean up microservice-connector"
    # clean up the images
    docker rmi $DOCKER_REGISTRY/gmcrouter:$VERSION
    docker rmi $DOCKER_REGISTRY/gmcmanager:$VERSION
}

if [ $# -eq 0 ]; then
    echo "Usage: $0 <function_name>"
    exit 1
fi

case "$1" in
    install_gmc)
        install_gmc
        ;;
    validate_gmc)
        validate_gmc
        ;;
    cleanup_gmc)
        cleanup_gmc
        ;;
    *)
        echo "Unknown function: $1"
        ;;
esac
