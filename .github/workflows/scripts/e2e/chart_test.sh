#!/bin/bash
# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

LOG_PATH=.

USER_ID=$(whoami)
CHART_MOUNT=/home/$USER_ID/charts-mnt
IMAGE_REPO=${OPEA_IMAGE_REPO:-amr-registry.caas.intel.com/aiops}
function init_codegen() {
    # executed under path helm-charts/codegen
    # init var
    MODELREPO=m-a-p
    MODELNAME=OpenCodeInterpreter-DS-6.7B
    MODELID=$MODELREPO/$MODELNAME
    MODELDOWNLOADID=models--$MODELREPO--$MODELNAME
    # IMAGE_REPO is $OPEA_IMAGE_REPO, or else ""

    ### PREPARE MODEL
    # check if the model is already downloaded
    if [ -d "$CHART_MOUNT/$MODELDOWNLOADID" ]; then
        echo "Model $MODELID already downloaded!"
        USE_MODELDOWNLOADID=True
    else
        echo "Downloading model $MODELID..."
        MODELDIR=$CHART_MOUNT/$MODELNAME
        if [ ! -d "$MODELDIR" ]; then
            mkdir -p $MODELDIR
        fi
        huggingface-cli download $MODELID --local-dir $MODELDIR --local-dir-use-symlinks False
        USE_MODELDOWNLOADID=False
    fi

    ### CONFIG VALUES.YAML
    # insert a prefix before opea/.*, the prefix is IMAGE_REPO
    sed -i "s#repository: opea/*#repository: $IMAGE_REPO/opea/#g" values.yaml
    # set huggingface token
    sed -i "s#insert-your-huggingface-token-here#$(cat /home/$USER_ID/.cache/huggingface/token)#g" values.yaml
    # replace the mount dir "Volume: *" with "Volume: $CHART_MOUNT"
    sed -i "s#volume: .*#volume: $CHART_MOUNT#g" values.yaml
    # replace the model ID with local dir name "data/$MODELNAME"
    if [ "$USE_MODELDOWNLOADID" = "False" ]; then
        sed -i "s#LLM_MODEL_ID: .*#LLM_MODEL_ID: /data/$MODELNAME#g" values.yaml
    else
        sed -i "s#LLM_MODEL_ID: .*#LLM_MODEL_ID: $MODELID#g" values.yaml
    fi
}

function init_chatqna() {
    # replace volume: /mnt with volume: $CHART_MOUNT
    find . -name '*.yaml' -type f -exec sed -i "s#volume: /mnt#volume: $CHART_MOUNT#g" {} \;
    # replace the repository "image: opea/*" with "image: $IMAGE_REPO/opea/"
    find . -name '*.yaml' -type f -exec sed -i "s#repository: opea/*#repository: $IMAGE_REPO/opea/#g" {} \;
    # set huggingface token
    find . -name '*.yaml' -type f -exec sed -i "s#insert-your-huggingface-token-here#$(cat /home/$USER_ID/.cache/huggingface/token)#g" {} \;
}

function validate_codegen() {
    # validate mega service
    ip_address=$(kubectl get svc $RELEASE_NAME -n $NAMESPACE -o jsonpath='{.spec.clusterIP}')
    port=$(kubectl get svc $RELEASE_NAME -n $NAMESPACE -o jsonpath='{.spec.ports[0].port}')
    # Curl the Mega Service
    curl http://${ip_address}:${port}/v1/codegen -H "Content-Type: application/json" -d '{
        "model": "ise-uiuc/Magicoder-S-DS-6.7B",
        "messages": "Implement a high-level API for a TODO list application. The API takes as input an operation request and updates the TODO list in place. If the request is invalid, raise an exception."}' > curl_megaservice.log

    echo "Checking response results, make sure the output is reasonable. "
    local status=true
    if [[ -f curl_megaservice.log ]] && \
    [[ $(grep -c "billion" curl_megaservice.log) != 0 ]]; then
        status=true
    fi

    if [ $status == false ]; then
        echo "Response check failed, please check the logs in artifacts!"
    else
        echo "Response check succeed!"
    fi
}

function validate_chatqna() {
    sleep 60
    # make sure microservice retriever svcname=$RELEASE_NAME-retriever-usvc is ready
    ip_address=$(kubectl get svc $RELEASE_NAME-retriever-usvc -n $NAMESPACE -o jsonpath='{.spec.clusterIP}')
    port=$(kubectl get svc $RELEASE_NAME-retriever-usvc -n $NAMESPACE -o jsonpath='{.spec.ports[0].port}')
    until curl http://${ip_address}:${port}/v1/retrieval -X POST \
    -d '{"text":"What is the revenue of Nike in 2023?","embedding":"'"${your_embedding}"'"}' \
    -H 'Content-Type: application/json'; do sleep 10; done
    ip_address=$(kubectl get svc $RELEASE_NAME -n $NAMESPACE -o jsonpath='{.spec.clusterIP}')
    port=$(kubectl get svc $RELEASE_NAME -n $NAMESPACE -o jsonpath='{.spec.ports[0].port}')
    # Curl the Mega Service
    curl http://${ip_address}:${port}/v1/chatqna -H "Content-Type: application/json" -d '{
        "model": "Intel/neural-chat-7b-v3-3",
        "messages": "What is the revenue of Nike in 2023?"}' > ${LOG_PATH}/curl_megaservice.log
    exit_code=$?

    echo "Checking response results, make sure the output is reasonable. "
    local status=false
    if [[ -f $LOG_PATH/curl_megaservice.log ]] && \
    [[ $(grep -c "billion" $LOG_PATH/curl_megaservice.log) != 0 ]]; then
        status=true
    fi

    if [ $status == false ]; then
        echo "Response check failed, please check the logs in artifacts!"
        exit 1
    else
        echo "Response check succeed!"
    fi
}

if [ $# -eq 0 ]; then
    echo "Usage: $0 <function_name>"
    exit 1
fi

case "$1" in
    init_codegen)
        pushd helm-charts/codegen
        init_codegen
        popd
        ;;
    validate_codegen)
        RELEASE_NAME=$2
        NAMESPACE=$3
        validate_codegen
        ;;
    init_chatqna)
        pushd helm-charts/chatqna
        init_chatqna
        popd
        ;;
    validate_chatqna)
        RELEASE_NAME=$2
        NAMESPACE=$3
        validate_chatqna
        ;;
    *)
        echo "Unknown function: $1"
        ;;
esac
