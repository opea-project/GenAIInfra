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

    init_gmc

    kubectl apply -f $(pwd)/config/crd/bases/gmc.opea.io_gmconnectors.yaml
    kubectl apply -f $(pwd)/config/rbac/gmc-manager-rbac.yaml
    kubectl create configmap gmcyaml -n $SYSTEM_NAMESPACE --from-file $(pwd)/config/manifests
    kubectl apply -f $(pwd)/config/manager/gmc-manager.yaml

    # Wait until the gmc controller pod is ready
    GMC_CONTROLLER_POD=$(kubectl get pods --namespace=$SYSTEM_NAMESPACE | awk 'NR>1 {print $1; exit}')
    wait_until_pod_ready "gmc-controller" $GMC_CONTROLLER_POD $SYSTEM_NAMESPACE
}

function validate_gmc() {
    echo "validate chat-qna"
    validate_chatqna

}

function cleanup_gmc() {
    echo "clean up microservice-connector"
    kubectl delete ns $APP_NAMESPACE
    kubectl delete ns $SYSTEM_NAMESPACE
    kubectl delete crd gmconnectors.gmc.opea.io
    # clean up the images
    docker rmi $DOCKER_REGISTRY/gmcrouter:$VERSION
    docker rmi $DOCKER_REGISTRY/gmcmanager:$VERSION
}

function validate_chatqna() {

   kubectl get pods -n $SYSTEM_NAMESPACE
   # todo select gaudi or xeon
   kubectl create ns $APP_NAMESPACE
   sed -i "s|namespace: chatqa|namespace: $APP_NAMESPACE|g"  $(pwd)/config/samples/chatQnA_xeon.yaml
   kubectl apply -f $(pwd)/config/samples/chatQnA_xeon.yaml



   output=$(kubectl get pods)
   echo $output

   # Wait until the router service is ready
   echo "Waiting for the chatqa router service to be ready..."
   ROUTER_POD=$(kubectl get pods --namespace=$APP_NAMESPACE -l app=router-service | awk 'NR>1 {print $1; exit}')
   wait_until_pod_ready "chatqna router" $ROUTER_POD $APP_NAMESPACE

  # Wait until the tgi pod is ready
  TGI_POD_NAME=$(kubectl get pods --namespace=$APP_NAMESPACE | grep ^tgi-service | awk '{print $1}')
  kubectl describe pod $TGI_POD_NAME -n $APP_NAMESPACE
  wait_until_pod_ready "tgi service" $TGI_POD_NAME $APP_NAMESPACE


   # deploy client pod for testing
   kubectl create deployment client-test -n $APP_NAMESPACE --image=python:3.8.13 -- sleep infinity

   # wait for client pod ready
   CLIENT_POD=$(kubectl get pods --namespace=$APP_NAMESPACE -l app=client-test | awk 'NR>1 {print $1; exit}')
   wait_until_pod_ready "client-test" $CLIENT_POD $APP_NAMESPACE
   # giving time to populating data
   sleep 120
   kubectl get pods -n $APP_NAMESPACE
   # send request to chatqnA
   export CLIENT_POD=$(kubectl get pod -n $APP_NAMESPACE -l app=client-test -o jsonpath={.items..metadata.name})
   echo "$CLIENT_POD"
   accessUrl=$(kubectl get gmc -n $APP_NAMESPACE -o jsonpath="{.items[?(@.metadata.name=='chatqa')].status.accessUrl}")
   kubectl exec "$CLIENT_POD" -n $APP_NAMESPACE -- curl $accessUrl  -X POST  -d '{"text":"What is the revenue of Nike in 2023?","parameters":{"max_new_tokens":17, "do_sample": true}}' -H 'Content-Type: application/json' > $LOG_PATH/curl_chatqna.log
   exit_code=$?
   if [ $exit_code -ne 0 ]; then
       echo "chatqna failed, please check the logs in ${LOG_PATH}!"
       exit 1
   fi

   echo "Checking response results, make sure the output is reasonable. "
   local status=false
   if [[ -f $LOG_PATH/curl_chatqna.log ]] && \
   [[ $(grep -c "billion" $LOG_PATH/curl_chatqna.log) != 0 ]]; then
       status=true
   fi
   if [ $status == false ]; then
       echo "Response check failed, please check the logs in artifacts!"
       exit 1
   else
       echo "Response check succeed!"
   fi
}

function init_gmc() {
    # Copy manifest into gmc
    mkdir -p $(pwd)/config/manifests
    cp $(dirname $(pwd))/manifests/ChatQnA/*.yaml -p $(pwd)/config/manifests/

    # replace tag with for the gmc-router and gmc-manager image
    sed -i "s|opea/\(.*\):latest|opea/\1:$VERSION|g" $(pwd)/config/gmcrouter/gmc-router.yaml
    sed -i "s|opea/\(.*\):latest|opea/\1:$VERSION|g" $(pwd)/config/manager/gmc-manager.yaml
    cp $(pwd)/config/gmcrouter/gmc-router.yaml -p $(pwd)/config/manifests/



    # replace namespace for gmc-router and gmc-manager
    sed -i "s|namespace: system|namespace: $SYSTEM_NAMESPACE|g"  $(pwd)/config/manager/gmc-manager.yaml
    sed -i "s|namespace: system|namespace: $SYSTEM_NAMESPACE|g"  $(pwd)/config/rbac/gmc-manager-rbac.yaml
    sed -i "s|name: system|name: $SYSTEM_NAMESPACE|g" $(pwd)/config/rbac/gmc-manager-rbac.yaml
    # replace the mount dir "path: /mnt/model" with "path: $CHART_MOUNT"
    find . -name '*.yaml' -type f -exec sed -i "s#path: /mnt/models#path: $MOUNT_DIR#g" {} \;
    # replace the repository "image: opea/*" with "image: $IMAGE_REPO/opea/"
    find . -name '*.yaml' -type f -exec sed -i "s#image: opea/*#image: $IMAGE_REPO/opea/#g" {} \;
    # set huggingface token
    # find . -name '*.yaml' -type f -exec sed -i "s#insert-your-huggingface-token-here#$(cat /home/$USER_ID/.cache/huggingface/token)#g" {} \;
    find . -name '*.yaml' -type f -exec sed -i "s#insert-your-huggingface-token-here#$(cat /home/$USER_ID/.cache/huggingface/token)#g" {} \;
    # replace namespace "default" with real namespace
    find . -name '*.yaml' -type f -exec sed -i "s#default.svc#$APP_NAMESPACE.svc#g" {} \;
}

function wait_until_pod_ready() {
    echo "Waiting for the $1 to be ready..."
    kubectl wait --for=condition=ready pod/$2 --namespace=$3 --timeout=300s
}


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
    validate_gmc)
        pushd microservices-connector
        validate_gmc
        popd
        ;;
    cleanup_gmc)
        cleanup_gmc
        ;;
    *)
        echo "Unknown function: $1"
        ;;
esac

