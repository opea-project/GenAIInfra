#!/bin/bash
# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

set -xe

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
source ${DIR}/utils.sh

USER_ID=$(whoami)
LOG_PATH=/home/$(whoami)/logs
CHATQNA_NAMESPACE="${APP_NAMESPACE}-chatqna"
CODEGEN_NAMESPACE="${APP_NAMESPACE}-codegen"
CODETRANS_NAMESPACE="${APP_NAMESPACE}-codetrans"
DOCSUM_NAMESPACE="${APP_NAMESPACE}-docsum"

function validate_gmc() {
    echo "validate chat-qna"
    validate_chatqna

    echo "validate codegen"
    validate_codegen

    echo "validate codetrans"
    validate_codetrans

    echo "validate docsum"
    validate_docsum

    get_gmc_controller_logs
}

function cleanup_apps() {
    echo "clean up microservice-connector"
    namespaces=("$CHATQNA_NAMESPACE" "$CODEGEN_NAMESPACE" "$CODETRANS_NAMESPACE" "$DOCSUM_NAMESPACE")
    for ns in "${namespaces[@]}"; do
        if kubectl get namespace $ns > /dev/null 2>&1; then
            echo "Deleting namespace: $ns"
            kubectl delete namespace "$ns"
        else
            echo "Namespace $ns does not exist"
        fi
    done
}

function validate_chatqna() {
   kubectl create ns $CHATQNA_NAMESPACE
   sed -i "s|namespace: chatqa|namespace: $CHATQNA_NAMESPACE|g"  $(pwd)/config/samples/chatQnA_gaudi.yaml
   kubectl apply -f $(pwd)/config/samples/chatQnA_gaudi.yaml

   # Wait until the router service is ready
   echo "Waiting for the chatqa router service to be ready..."
   wait_until_pod_ready "chatqna router" $CHATQNA_NAMESPACE "router-service"
   output=$(kubectl get pods -n $CHATQNA_NAMESPACE)
   echo $output

   # Wait until the tgi pod is ready
   TGI_POD_NAME=$(kubectl get pods --namespace=$CHATQNA_NAMESPACE | grep ^tgi-gaudi-svc | awk '{print $1}')
   kubectl describe pod $TGI_POD_NAME -n $CHATQNA_NAMESPACE
   kubectl wait --for=condition=ready pod/$TGI_POD_NAME --namespace=$CHATQNA_NAMESPACE --timeout=300s

   # deploy client pod for testing
   kubectl create deployment client-test -n $CHATQNA_NAMESPACE --image=python:3.8.13 -- sleep infinity

   # wait for client pod ready
   wait_until_pod_ready "client-test" $CHATQNA_NAMESPACE "client-test"
   # giving time to populating data
   sleep 90

   kubectl get pods -n $CHATQNA_NAMESPACE
   # send request to chatqna
   export CLIENT_POD=$(kubectl get pod -n $CHATQNA_NAMESPACE -l app=client-test -o jsonpath={.items..metadata.name})
   echo "$CLIENT_POD"
   accessUrl=$(kubectl get gmc -n $CHATQNA_NAMESPACE -o jsonpath="{.items[?(@.metadata.name=='chatqa')].status.accessUrl}")
   kubectl exec "$CLIENT_POD" -n $CHATQNA_NAMESPACE -- curl $accessUrl  -X POST  -d '{"text":"What is the revenue of Nike in 2023?","parameters":{"max_new_tokens":17, "do_sample": true}}' -H 'Content-Type: application/json' > $LOG_PATH/curl_chatqna.log
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
       if [[ -f $LOG_PATH/curl_chatqna.log ]]; then
           cat $LOG_PATH/curl_chatqna.log
       fi
       echo "Response check failed, please check the logs in artifacts!"
       exit 1
   else
       echo "Response check succeed!"
   fi
}

function validate_codegen() {
   kubectl create ns $CODEGEN_NAMESPACE
   sed -i "s|namespace: codegen|namespace: $CODEGEN_NAMESPACE|g"  $(pwd)/config/samples/codegen_gaudi.yaml
   kubectl apply -f $(pwd)/config/samples/codegen_gaudi.yaml

   # Wait until the router service is ready
   echo "Waiting for the codegen router service to be ready..."
   wait_until_pod_ready "codegen router" $CODEGEN_NAMESPACE "router-service"
   output=$(kubectl get pods -n $CODEGEN_NAMESPACE)
   echo $output


   # deploy client pod for testing
   kubectl create deployment client-test -n $CODEGEN_NAMESPACE --image=python:3.8.13 -- sleep infinity

   # wait for client pod ready
   wait_until_pod_ready "client-test" $CODEGEN_NAMESPACE "client-test"
   # giving time to populating data
   sleep 60

   kubectl get pods -n $CODEGEN_NAMESPACE
   # send request to codegen
   export CLIENT_POD=$(kubectl get pod -n $CODEGEN_NAMESPACE -l app=client-test -o jsonpath={.items..metadata.name})
   echo "$CLIENT_POD"
   accessUrl=$(kubectl get gmc -n $CODEGEN_NAMESPACE -o jsonpath="{.items[?(@.metadata.name=='codegen')].status.accessUrl}")
   kubectl exec "$CLIENT_POD" -n $CODEGEN_NAMESPACE -- curl $accessUrl  -X POST  -d '{"query": "def print_hello_world():"}' -H 'Content-Type: application/json' > $LOG_PATH/gmc_codegen.log
   exit_code=$?
   if [ $exit_code -ne 0 ]; then
       echo "codegen failed, please check the logs in ${LOG_PATH}!"
       exit 1
   fi

   echo "Checking response results, make sure the output is reasonable. "
   local status=false
   if [[ -f $LOG_PATH/gmc_codegen.log ]] && \
   [[ $(grep -c "[DONE]" $LOG_PATH/gmc_codegen.log) != 0 ]]; then
       status=true
   fi
   if [ $status == false ]; then
       if [[ -f $LOG_PATH/gmc_codegen.log ]]; then
           cat $LOG_PATH/gmc_codegen.log
       fi
       echo "Response check failed, please check the logs in artifacts!"
       cat $LOG_PATH/gmc_codegen.log
       exit 1
   else
       echo "Response check succeed!"
   fi
}

function validate_codetrans() {
   kubectl create ns $CODETRANS_NAMESPACE
   sed -i "s|namespace: codetrans|namespace: $CODETRANS_NAMESPACE|g"  $(pwd)/config/samples/codetrans_gaudi.yaml
   kubectl apply -f $(pwd)/config/samples/codetrans_gaudi.yaml

   # Wait until the router service is ready
   echo "Waiting for the codetrans router service to be ready..."
   wait_until_pod_ready "codetrans router" $CODETRANS_NAMESPACE "router-service"
   output=$(kubectl get pods -n $CODETRANS_NAMESPACE)
   echo $output


   # deploy client pod for testing
   kubectl create deployment client-test -n $CODETRANS_NAMESPACE --image=python:3.8.13 -- sleep infinity

   # wait for client pod ready
   wait_until_pod_ready "client-test" $CODETRANS_NAMESPACE "client-test"
   # giving time to populating data
   sleep 60

   kubectl get pods -n $CODETRANS_NAMESPACE
   # send request to codetrans
   export CLIENT_POD=$(kubectl get pod -n $CODETRANS_NAMESPACE -l app=client-test -o jsonpath={.items..metadata.name})
   echo "$CLIENT_POD"
   accessUrl=$(kubectl get gmc -n $CODETRANS_NAMESPACE -o jsonpath="{.items[?(@.metadata.name=='codetrans')].status.accessUrl}")
   kubectl exec "$CLIENT_POD" -n $CODETRANS_NAMESPACE -- curl $accessUrl  -X POST  -d '{"query":"    ### System: Please translate the following Golang codes into  Python codes.    ### Original codes:    '\'''\'''\''Golang    \npackage main\n\nimport \"fmt\"\nfunc main() {\n    fmt.Println(\"Hello, World!\");\n    '\'''\'''\''    ### Translated codes:"}' -H 'Content-Type: application/json' > $LOG_PATH/gmc_codetrans.log
   exit_code=$?
   if [ $exit_code -ne 0 ]; then
       echo "codetrans failed, please check the logs in ${LOG_PATH}!"
       exit 1
   fi

   echo "Checking response results, make sure the output is reasonable. "
   local status=false
   if [[ -f $LOG_PATH/gmc_codetrans.log ]] && \
   [[ $(grep -c "[DONE]" $LOG_PATH/gmc_codetrans.log) != 0 ]]; then
       status=true
   fi
   if [ $status == false ]; then
       if [[ -f $LOG_PATH/gmc_codetrans.log ]]; then
           cat $LOG_PATH/gmc_codetrans.log
       fi
       echo "Response check failed, please check the logs in artifacts!"
       exit 1
   else
       echo "Response check succeed!"
   fi
}

function validate_docsum() {
   kubectl create ns $DOCSUM_NAMESPACE
   sed -i "s|namespace: docsum|namespace: $DOCSUM_NAMESPACE|g"  $(pwd)/config/samples/docsum_gaudi.yaml
   kubectl apply -f $(pwd)/config/samples/docsum_gaudi.yaml

   # Wait until the router service is ready
   echo "Waiting for the docsum router service to be ready..."
   wait_until_pod_ready "docsum router" $DOCSUM_NAMESPACE "router-service"
   output=$(kubectl get pods -n $DOCSUM_NAMESPACE)
   echo $output

   # deploy client pod for testing
   kubectl create deployment client-test -n $DOCSUM_NAMESPACE --image=python:3.8.13 -- sleep infinity

   # wait for client pod ready
   wait_until_pod_ready "client-test" $DOCSUM_NAMESPACE "client-test"
   # giving time to populating data
   sleep 60

   kubectl get pods -n $DOCSUM_NAMESPACE
   # send request to docsum
   export CLIENT_POD=$(kubectl get pod -n $DOCSUM_NAMESPACE -l app=client-test -o jsonpath={.items..metadata.name})
   echo "$CLIENT_POD"
   accessUrl=$(kubectl get gmc -n $DOCSUM_NAMESPACE -o jsonpath="{.items[?(@.metadata.name=='docsum')].status.accessUrl}")
   kubectl exec "$CLIENT_POD" -n $DOCSUM_NAMESPACE -- curl $accessUrl  -X POST  -d '{"query":"Text Embeddings Inference (TEI) is a toolkit for deploying and serving open source text embeddings and sequence classification models. TEI enables high-performance extraction for the most popular models, including FlagEmbedding, Ember, GTE and E5."}'  -H 'Content-Type: application/json' > $LOG_PATH/gmc_docsum.log
   exit_code=$?
   if [ $exit_code -ne 0 ]; then
       echo "docsum failed, please check the logs in ${LOG_PATH}!"
       exit 1
   fi

   echo "Checking response results, make sure the output is reasonable. "
   local status=false
   if [[ -f $LOG_PATH/gmc_docsum.log ]] && \
   [[ $(grep -c "[DONE]" $LOG_PATH/gmc_docsum.log) != 0 ]]; then
       status=true
   fi
   if [ $status == false ]; then
       if [[ -f $LOG_PATH/gmc_docsum.log ]]; then
           cat $LOG_PATH/gmc_docsum.log
       fi
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
    validate_gmc)
        pushd microservices-connector
        validate_gmc
        popd
        ;;
    cleanup_apps)
        cleanup_apps
        ;;
    *)
        echo "Unknown function: $1"
        ;;
esac
