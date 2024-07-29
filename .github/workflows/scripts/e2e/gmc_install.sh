#!/bin/bash
# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

set -xe

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
source ${DIR}/utils.sh

USER_ID=$(whoami)
LOG_PATH=/home/$(whoami)/logs
MOUNT_DIR=${KIND_MOUNT_DIR:-"/home/$USER_ID/.cache/huggingface/hub"}
TOKEN_DIR=${KIND_TOKEN_DIR:-"/home/$USER_ID/.cache/huggingface/token"}
IMAGE_REPO=${OPEA_IMAGE_REPO:-""}
VERSION=${VERSION:-"latest"}

function install_gmc() {
    PLATFORM=${1:-"xeon"}
    # Make sure you have to use image tag $VERSION for microservice-connector installation
    echo "install microservice-connector on $PLATFORM, using repo $IMAGE_REPO and tag $VERSION"
    echo "using namespace $SYSTEM_NAMESPACE"

    init_gmc

    kubectl apply -f $(pwd)/config/crd/bases/gmc.opea.io_gmconnectors.yaml
    kubectl apply -f $(pwd)/config/rbac/gmc-manager-rbac.yaml
    kubectl create configmap gmcyaml -n $SYSTEM_NAMESPACE --from-file $(pwd)/config/manifests
    kubectl apply -f $(pwd)/config/manager/gmc-manager.yaml

    # Wait until the gmc controller pod is ready
    wait_until_pod_ready "gmc-controller" $SYSTEM_NAMESPACE "gmc-controller"
    kubectl get pods -n $SYSTEM_NAMESPACE
}

function copy_manifests() {
    # Copy manifest into gmc
    mkdir -p $(pwd)/config/manifests
    cp $(dirname $(pwd))/manifests/common/*.yaml -p $(pwd)/config/manifests/
}

function init_gmc() {
    # copy manifests
    copy_manifests

    # replace tag with for the gmc-router and gmc-manager image
    sed -i "s|opea/\(.*\):latest|opea/\1:$VERSION|g" $(pwd)/config/gmcrouter/gmc-router.yaml
    sed -i "s|opea/\(.*\):latest|opea/\1:$VERSION|g" $(pwd)/config/manager/gmc-manager.yaml
    # replace the pull policy "IfNotPresent" with "Always"
    sed -i "s#imagePullPolicy: IfNotPresent#imagePullPolicy: Always#g" $(pwd)/config/gmcrouter/gmc-router.yaml
    sed -i "s#imagePullPolicy: IfNotPresent#imagePullPolicy: Always#g" $(pwd)/config/manager/gmc-manager.yaml

    cp $(pwd)/config/gmcrouter/gmc-router.yaml -p $(pwd)/config/manifests/

    # replace namespace for gmc-router and gmc-manager
    sed -i "s|namespace: system|namespace: $SYSTEM_NAMESPACE|g"  $(pwd)/config/manager/gmc-manager.yaml
    sed -i "s|namespace: system|namespace: $SYSTEM_NAMESPACE|g"  $(pwd)/config/rbac/gmc-manager-rbac.yaml
    sed -i "s|name: system|name: $SYSTEM_NAMESPACE|g" $(pwd)/config/rbac/gmc-manager-rbac.yaml

    # replace the mount dir "path: /mnt/model" with "path: $CHART_MOUNT"
    # find . -name '*.yaml' -type f -exec sed -i "s#path: /mnt/models#path: $MOUNT_DIR#g" {} \;
    find . -name '*.yaml' -type f -exec sed -i "s#path: /mnt/opea-models#path: $MOUNT_DIR#g" {} \;
    # replace the repository "image: opea/*" with "image: ${IMAGE_REPO}opea/"
    find . -name '*.yaml' -type f -exec sed -i "s#image: opea/*#image: ${IMAGE_REPO}opea/#g" {} \;
    find . -name '*.yaml' -type f -exec sed -i "s#image: \"opea/*#image: \"${IMAGE_REPO}opea/#g" {} \;
    # set huggingface token
    #find . -name '*.yaml' -type f -exec sed -i "s#insert-your-huggingface-token-here#$(cat $TOKEN_DIR)#g" {} \;
    # replace the pull policy "IfNotPresent" with "Always"
    find . -name '*.yaml' -type f -exec sed -i "s#imagePullPolicy: IfNotPresent#imagePullPolicy: Always#g" {} \;
}

function cleanup_gmc() {
    echo "clean up microservice-connector"
    if kubectl get namespace $SYSTEM_NAMESPACE > /dev/null 2>&1; then
        echo "Deleting namespace: $SYSTEM_NAMESPACE"
        kubectl delete namespace "$SYSTEM_NAMESPACE"
        kubectl delete crd gmconnectors.gmc.opea.io
    else
        echo "Namespace $SYSTEM_NAMESPACE does not exist"
    fi
}

#------MAIN-----------
if [ $# -eq 0 ]; then
    echo "Usage: $0 <function_name>"
    exit 1
fi
case "$1" in
    install_gmc)
        pushd microservices-connector
        install_gmc
        popd
        ;;
    cleanup_gmc)
        cleanup_gmc
        ;;
    *)
        echo "Unknown function: $1"
        ;;
esac
