#!/bin/bash
# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

set -xe
USER_ID=$(whoami)
LOG_PATH=/home/$(whoami)/logs
MOUNT_DIR=/home/$USER_ID/.cache/huggingface/hub
# IMAGE_REPO is $OPEA_IMAGE_REPO, or else ""
IMAGE_REPO=${OPEA_IMAGE_REPO:-""}

function init_docsum() {
    # executed under path manifest/docsum/xeon
    # replace the mount dir "path: /mnt/model" with "path: $CHART_MOUNT"
    find . -name '*.yaml' -type f -exec sed -i "s#path: /mnt#path: $MOUNT_DIR#g" {} \;
    # replace the repository "image: opea/*" with "image: ${IMAGE_REPO}opea/"
    find . -name '*.yaml' -type f -exec sed -i "s#image: \"opea/*#image: \"${IMAGE_REPO}opea/#g" {} \;
    # set huggingface token
    find . -name '*.yaml' -type f -exec sed -i "s#insert-your-huggingface-token-here#$(cat /home/$USER_ID/.cache/huggingface/token)#g" {} \;
}

function init_codetrans() {
    # executed under path manifest/codetrans/xeon
    # replace the mount dir "path: /mnt/model" with "path: $CHART_MOUNT"
    find . -name '*.yaml' -type f -exec sed -i "s#path: /mnt#path: $MOUNT_DIR#g" {} \;
    # replace the repository "image: opea/*" with "image: ${IMAGE_REPO}opea/"
    find . -name '*.yaml' -type f -exec sed -i "s#image: \"opea/*#image: \"${IMAGE_REPO}opea/#g" {} \;
    # set huggingface token
    find . -name '*.yaml' -type f -exec sed -i "s#insert-your-huggingface-token-here#$(cat /home/$USER_ID/.cache/huggingface/token)#g" {} \;
}

function init_codegen() {
    # executed under path manifest/codegen/xeon
    # replace the mount dir "path: /mnt/model" with "path: $CHART_MOUNT"
    find . -name '*.yaml' -type f -exec sed -i "s#path: /mnt#path: $MOUNT_DIR#g" {} \;
    # replace the repository "image: opea/*" with "image: ${IMAGE_REPO}opea/"
    find . -name '*.yaml' -type f -exec sed -i "s#image: \"opea/*#image: \"${IMAGE_REPO}opea/#g" {} \;
    # set huggingface token
    find . -name '*.yaml' -type f -exec sed -i "s#insert-your-huggingface-token-here#$(cat /home/$USER_ID/.cache/huggingface/token)#g" {} \;
}

function install_docsum {
    echo "namespace is $NAMESPACE"
    find . -name 'qna_configmap_gaudi.yaml' -type f -exec sed -i "s#default#${NAMESPACE}#g" {} \;
    kubectl apply -f qna_configmap_gaudi.yaml -n $NAMESPACE
    kubectl apply -f docsum_gaudi_llm.yaml -n $NAMESPACE
    kubectl apply -f tgi_gaudi_service.yaml -n $NAMESPACE
}

function install_codetrans {
    echo "namespace is $NAMESPACE"
    kubectl apply -f . -n $NAMESPACE
}

function install_codegen {
    echo "namespace is $NAMESPACE"
    kubectl apply -f . -n $NAMESPACE
}

function init_chatqna() {
    # executed under path manifest/chatqna/xeon
    # replace the mount dir "path: /mnt" with "path: $CHART_MOUNT"
    find . -name '*.yaml' -type f -exec sed -i "s#path: /mnt/models#path: $MOUNT_DIR#g" {} \;
    # replace the repository "image: opea/*" with "image: ${IMAGE_REPO}opea/"
    find . -name '*.yaml' -type f -exec sed -i "s#image: opea/*#image: ${IMAGE_REPO}opea/#g" {} \;
    # set huggingface token
    find . -name '*.yaml' -type f -exec sed -i "s#insert-your-huggingface-token-here#$(cat /home/$USER_ID/.cache/huggingface/token)#g" {} \;
}

function install_chatqna {
    # replace namespace "default" with real namespace
    find . -name '*.yaml' -type f -exec sed -i "s#default.svc#$NAMESPACE.svc#g" {} \;
    # for very yaml file in yaml_files, apply it to the k8s cluster
    yaml_files=("qna_configmap_gaudi" "redis-vector-db"  "tei_embedding_gaudi_service" "tei_reranking_service" "tgi_gaudi_service" "retriever" "embedding" "reranking" "llm")
    for yaml_file in ${yaml_files[@]}; do
        kubectl apply -f $yaml_file.yaml -n $NAMESPACE
    done
    sleep 60
    kubectl apply -f chaqna-xeon-backend-server.yaml -n $NAMESPACE
}

function validate_docsum() {
    ip_address=$(kubectl get svc $SERVICE_NAME -n $NAMESPACE -o jsonpath='{.spec.clusterIP}')
    port=$(kubectl get svc $SERVICE_NAME -n $NAMESPACE -o jsonpath='{.spec.ports[0].port}')
    echo "try to curl http://${ip_address}:${port}/v1/chat/docsum..."
    # Curl the DocSum LLM Service
    curl http://${ip_address}:${port}/v1/chat/docsum \
      -X POST \
      -d '{"query":"Text Embeddings Inference (TEI) is a toolkit for deploying and serving open source text embeddings and sequence classification models. TEI enables high-performance extraction for the most popular models, including FlagEmbedding, Ember, GTE and E5."}' \
      -H 'Content-Type: application/json' > $LOG_PATH/curl_docsum.log
    exit_code=$?
    if [ $exit_code -ne 0 ]; then
        echo "LLM for docsum failed, please check the logs in ${LOG_PATH}!"
        exit 1
    fi

    echo "Checking response results, make sure the output is reasonable. "
    local status=false
    if [[ -f $LOG_PATH/curl_docsum.log ]] && \
    [[ $(grep -c "TEI" $LOG_PATH/curl_docsum.log) != 0 ]]; then
        status=true
    fi

    if [ $status == false ]; then
        echo "Response check failed, please check the logs in artifacts!"
    else
        echo "Response check succeed!"
    fi
}

function validate_codetrans() {
    ip_address=$(kubectl get svc $SERVICE_NAME -n $NAMESPACE -o jsonpath='{.spec.clusterIP}')
    port=$(kubectl get svc $SERVICE_NAME -n $NAMESPACE -o jsonpath='{.spec.ports[0].port}')
    echo "try to curl http://${ip_address}:${port}/v1/chat/completions..."
    # Curl the CodeTrans LLM Service
    curl http://${ip_address}:${port}/v1/chat/completions \
      -X POST \
      -d '{"query":"    ### System: Please translate the following Golang codes into  Python codes.    ### Original codes:    '\'''\'''\''Golang    \npackage main\n\nimport \"fmt\"\nfunc main() {\n    fmt.Println(\"Hello, World!\");\n    '\'''\'''\''    ### Translated codes:"}' \
      -H 'Content-Type: application/json' > $LOG_PATH/curl_codetrans.log
    exit_code=$?
    if [ $exit_code -ne 0 ]; then
        echo "LLM for codetrans failed, please check the logs in ${LOG_PATH}!"
        exit 1
    fi

    echo "Checking response results, make sure the output is reasonable. "
    local status=false
    if [[ -f $LOG_PATH/curl_codetrans.log ]] && \
    [[ $(grep -c "Hello" $LOG_PATH/curl_codetrans.log) != 0 ]]; then
        status=true
    fi

    if [ $status == false ]; then
        echo "Response check failed, please check the logs in artifacts!"
    else
        echo "Response check succeed!"
    fi
}

function validate_codegen() {
    ip_address=$(kubectl get svc $SERVICE_NAME -n $NAMESPACE -o jsonpath='{.spec.clusterIP}')
    port=$(kubectl get svc $SERVICE_NAME -n $NAMESPACE -o jsonpath='{.spec.ports[0].port}')
    echo "try to curl http://${ip_address}:${port}/v1/codegen..."
    # Curl the Mega Service
    curl http://${ip_address}:${port}/v1/codegen \
    -H "Content-Type: application/json" \
    -d '{"messages": "Implement a high-level API for a TODO list application. The API takes as input an operation request and updates the TODO list in place. If the request is invalid, raise an exception."}' > $LOG_PATH/curl_codegen.log
    exit_code=$?
    if [ $exit_code -ne 0 ]; then
        echo "Megaservice codegen failed, please check the logs in ${LOG_PATH}!"
        exit 1
    fi

    echo "Checking response results, make sure the output is reasonable. "
    local status=false
    if [[ -f $LOG_PATH/curl_codegen.log ]] && \
    [[ $(grep -c "print" $LOG_PATH/curl_codegen.log) != 0 ]]; then
        status=true
    fi

    if [ $status == false ]; then
        echo "Response check failed, please check the logs in artifacts!"
    else
        echo "Response check succeed!"
    fi
}

function validate_chatqna() {
    # make sure microservice retriever is ready
    test_embedding=$(python3 -c "import random; embedding = [random.uniform(-1, 1) for _ in range(768)]; print(embedding)")
    until curl http://retriever-svc.$NAMESPACE:7000/v1/retrieval -X POST \
    -d '{"text":"What is the revenue of Nike in 2023?","embedding":"'"${test_embedding}"'"}' \
    -H 'Content-Type: application/json'; do sleep 10; done

    # make sure microservice tgi-svc is ready
    until curl http://tgi-gaudi-svc.$NAMESPACE:9009/generate -X POST \
    -d '{"inputs":"What is Deep Learning?","parameters":{"max_new_tokens":17, "do_sample": true}}' \
    -H 'Content-Type: application/json'; do sleep 10; done

    # Curl the Mega Service
    curl http://chaqna-xeon-backend-server-svc.$NAMESPACE:8888/v1/chatqna -H "Content-Type: application/json" \
    -d '{ "messages": "What is the revenue of Nike in 2023?" }' > $LOG_PATH/curl_chatqna.log
    exit_code=$?
    if [ $exit_code -ne 0 ]; then
        echo "Megaservice chatqna failed, please check the logs in ${LOG_PATH}!"
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

if [ $# -eq 0 ]; then
    echo "Usage: $0 <function_name>"
    exit 1
fi

case "$1" in
    init_docsum)
        cp manifests/ChatQnA/qna_configmap_gaudi.yaml manifests/DocSum/gaudi/
        pushd manifests/DocSum/gaudi
        init_docsum
        popd
        ;;
    init_codetrans)
        pushd manifests/CodeTrans/gaudi
        init_codetrans
        popd
        ;;
    init_codegen)
        pushd manifests/CodeGen/gaudi
        init_codegen
        popd
        ;;
    init_chatqna)
        pushd manifests/ChatQnA
        init_chatqna
        popd
        ;;
    install_docsum)
        pushd manifests/DocSum/gaudi
        NAMESPACE=$2
        install_docsum
        popd
        ;;
    install_codetrans)
        pushd manifests/CodeTrans/gaudi
        NAMESPACE=$2
        install_codetrans
        popd
        ;;
    install_codegen)
        pushd manifests/CodeGen/gaudi
        NAMESPACE=$2
        install_codegen
        popd
        ;;
    install_chatqna)
        pushd manifests/ChatQnA
        NAMESPACE=$2
        install_chatqna
        popd
        ;;
    validate_docsum)
        NAMESPACE=$2
        SERVICE_NAME=docsum-llm-uservice
        validate_docsum
        ;;
    validate_codetrans)
        NAMESPACE=$2
        SERVICE_NAME=codetrans-llm-uservice
        validate_codetrans
        ;;
    validate_codegen)
        NAMESPACE=$2
        SERVICE_NAME=codegen
        validate_codegen
        ;;
    validate_chatqna)
        NAMESPACE=$2
        SERVICE_NAME=chaqna-xeon-backend-server-svc
        validate_chatqna
        ;;
    *)
        echo "Unknown function: $1"
        ;;
esac
