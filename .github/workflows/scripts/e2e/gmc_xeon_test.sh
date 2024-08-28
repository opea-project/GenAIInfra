#!/bin/bash
# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

set -xe

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
source ${DIR}/utils.sh

USER_ID=$(whoami)
LOG_PATH=/home/$(whoami)/logs
AUDIOQA_NAMESPACE="${APP_NAMESPACE}-audioqa"
CHATQNA_NAMESPACE="${APP_NAMESPACE}-chatqna"
CHATQNA_DATAPREP_NAMESPACE="${APP_NAMESPACE}-chatqna-dataprep"
CHATQNA_SWITCH_NAMESPACE="${APP_NAMESPACE}-chatqna-switch"
CODEGEN_NAMESPACE="${APP_NAMESPACE}-codegen"
CODETRANS_NAMESPACE="${APP_NAMESPACE}-codetrans"
DOCSUM_NAMESPACE="${APP_NAMESPACE}-docsum"
DELETE_STEP_NAMESPACE="${APP_NAMESPACE}-delstep"
MODIFY_STEP_NAMESPACE="${APP_NAMESPACE}-modstep"
WEBHOOK_NAMESPACE="${APP_NAMESPACE}-webhook"

function validate_gmc() {
    echo "validate audio-qna"
    validate_audioqa

    echo "validate chat-qna"
    validate_chatqna

    echo "validate chat-qna with dataprep"
    validate_chatqna_with_dataprep

    echo "validate chat-qna in switch mode"
    validate_chatqna_in_switch

    echo "validate change graph"
    validate_modify_config
    validate_remove_step

    # echo "validate codegen"
    # validate_codegen

    # echo "validate codetrans"
    # validate_codetrans

    # echo "validate docsum"
    # validate_docsum

    echo "validate webhook"
    validate_webhook

    get_gmc_controller_logs
}

function validate_webhook() {
    kubectl create ns $WEBHOOK_NAMESPACE || echo "namespace $WEBHOOK_NAMESPACE is created."
    # validate root node existence
    yq ".spec.nodes.node123 = .spec.nodes.root | del(.spec.nodes.root)" config/samples/ChatQnA/chatQnA_xeon.yaml > /tmp/webhook-case1.yaml
    sed -i "s|namespace: chatqa|namespace: $WEBHOOK_NAMESPACE|g"  /tmp/webhook-case1.yaml
    output=$(! kubectl apply -f /tmp/webhook-case1.yaml 2>&1)
    if ! (echo $output | grep -q "a root node is required"); then
        echo "Root node existence validation error message is not found!"
        echo $output
        exit 1
    fi

    # StepName validation
    yq '(.spec.nodes.root.steps[] | select ( .name == "Llm")).name = "xyz"' config/samples/ChatQnA/chatQnA_gaudi.yaml > /tmp/webhook-case2.yaml
    sed -i "s|namespace: chatqa|namespace: $WEBHOOK_NAMESPACE|g"  /tmp/webhook-case2.yaml
    output=$(! kubectl apply -f /tmp/webhook-case2.yaml 2>&1)
    if ! (echo $output | grep -q "invalid step name"); then
        echo "Step name validation error message is not found!"
        echo $output
        exit 1
    fi

    # nodeName existence
    yq '(.spec.nodes.root.steps[] | select ( .name == "Embedding")).nodeName = "node123"' config/samples/ChatQnA/chatQnA_switch_xeon.yaml > /tmp/webhook-case3.yaml
    sed -i "s|namespace: switch|namespace: $WEBHOOK_NAMESPACE|g"  /tmp/webhook-case3.yaml
    output=$(! kubectl apply -f /tmp/webhook-case3.yaml 2>&1)
    if ! (echo $output | grep -q "node name: node123 in step Embedding does not exist"); then
        echo "nodeName existence validation error message is not found!"
        echo $output
        exit 1
    fi

    # serviceName uniqueness
    yq '(.spec.nodes.node1.steps[] | select ( .name == "Embedding")).internalService.serviceName = "tei-embedding-svc-bge15"' config/samples/ChatQnA/chatQnA_switch_xeon.yaml > /tmp/webhook-case4.yaml
    sed -i "s|namespace: switch|namespace: $WEBHOOK_NAMESPACE|g"  /tmp/webhook-case4.yaml
    output=$(! kubectl apply -f /tmp/webhook-case4.yaml 2>&1)
    if ! (echo $output | grep -q "service name: tei-embedding-svc-bge15 in node node1 already exists"); then
        echo "serviceName uniqueness validation error message is not found!"
        echo $output
        exit 1
    fi

    # clean up cases
    rm -f /tmp/webhook-case*.yaml
}

function cleanup_apps() {
    echo "clean up microservice-connector"
    # namespaces=("$CHATQNA_NAMESPACE" "$CHATQNA_DATAPREP_NAMESPACE" "$CHATQNA_SWITCH_NAMESPACE" "$CODEGEN_NAMESPACE" "$CODETRANS_NAMESPACE" "$DOCSUM_NAMESPACE")
    namespaces=("$AUDIOQA_NAMESPACE" "$CHATQNA_NAMESPACE" "$CHATQNA_DATAPREP_NAMESPACE" "$CHATQNA_SWITCH_NAMESPACE" "$WEBHOOK_NAMESPACE" "$MODIFY_STEP_NAMESPACE" "$DELETE_STEP_NAMESPACE")
    for ns in "${namespaces[@]}"; do
        if kubectl get namespace $ns > /dev/null 2>&1; then
            echo "Deleting namespace: $ns"
            kubectl delete namespace "$ns"
        else
            echo "Namespace $ns does not exist"
        fi
    done
}

function validate_audioqa() {
   kubectl create ns $AUDIOQA_NAMESPACE
   sed -i "s|namespace: audioqa|namespace: $AUDIOQA_NAMESPACE|g"  $(pwd)/config/samples/AudioQnA/audioQnA_xeon.yaml
   kubectl apply -f $(pwd)/config/samples/AudioQnA/audioQnA_xeon.yaml

   # Wait until the router service is ready
   echo "Waiting for the audioqa router service to be ready..."
   wait_until_pod_ready "audioqa router" $AUDIOQA_NAMESPACE "router-service"
   output=$(kubectl get pods -n $AUDIOQA_NAMESPACE)
   echo $output

   # deploy client pod for testing
   kubectl create deployment client-test -n $AUDIOQA_NAMESPACE --image=python:3.8.13 -- sleep infinity

   # Wait until all pods are ready
   wait_until_all_pod_ready $AUDIOQA_NAMESPACE 300s
   if [ $? -ne 0 ]; then
       echo "Error Some pods are not ready!"
       exit 1
   fi

    pods_count=$(kubectl get pods -n $AUDIOQA_NAMESPACE -o jsonpath='{.items[*].metadata.name}' | wc -w)
    # expected_ready_pods, expected_external_pods, expected_total_pods
    # pods_count-1 is to exclude the client pod in this namespace
    check_gmc_status $AUDIOQA_NAMESPACE 'audioqa' $((pods_count-1)) 0 7
    if [ $? -ne 0 ]; then
       echo "GMC status is not as expected"
       exit 1
    fi

   # giving time to populating data
   sleep 90

   kubectl get pods -n $AUDIOQA_NAMESPACE
   # send request to chatqnA
   export CLIENT_POD=$(kubectl get pod -n $AUDIOQA_NAMESPACE -l app=client-test -o jsonpath={.items..metadata.name})
   echo "$CLIENT_POD"
   accessUrl=$(kubectl get gmc -n $AUDIOQA_NAMESPACE -o jsonpath="{.items[?(@.metadata.name=='audioqa')].status.accessUrl}")
   byte_str=$(kubectl exec "$CLIENT_POD" -n $AUDIOQA_NAMESPACE -- curl $accessUrl -s -X POST  -d '{"byte_str": "UklGRigAAABXQVZFZm10IBIAAAABAAEARKwAAIhYAQACABAAAABkYXRhAgAAAAEA", "parameters":{"max_new_tokens":64, "do_sample": true, "streaming":false}}' -H 'Content-Type: application/json' | jq .byte_str)
   if [ -z "$byte_str" ]; then
       echo "audioqa failed, please check the the!"
       exit 1
   fi
   echo "Audioqa response check succeed!"

   kubectl delete gmc -n $AUDIOQA_NAMESPACE 'audioqa'
   echo "sleep 10s for cleaning up"
   sleep 10
   check_resource_cleared $AUDIOQA_NAMESPACE
}

function validate_chatqna() {
   kubectl create ns $CHATQNA_NAMESPACE
   sed -i "s|namespace: chatqa|namespace: $CHATQNA_NAMESPACE|g"  $(pwd)/config/samples/ChatQnA/chatQnA_xeon.yaml
   # workaround for issue #268
   #yq -i '(.spec.nodes.root.steps[] | select ( .name == "Tgi")).internalService.config.MODEL_ID = "bigscience/bloom-560m"' $(pwd)/config/samples/ChatQnA/chatQnA_xeon.yaml
   kubectl apply -f $(pwd)/config/samples/ChatQnA/chatQnA_xeon.yaml

   # Wait until the router service is ready
   echo "Waiting for the chatqa router service to be ready..."
   wait_until_pod_ready "chatqna router" $CHATQNA_NAMESPACE "router-service"
   output=$(kubectl get pods -n $CHATQNA_NAMESPACE)
   echo $output

   # deploy client pod for testing
   kubectl create deployment client-test -n $CHATQNA_NAMESPACE --image=python:3.8.13 -- sleep infinity

   # Wait until all pods are ready
   wait_until_all_pod_ready $CHATQNA_NAMESPACE 300s
   if [ $? -ne 0 ]; then
       echo "Error Some pods are not ready!"
       exit 1
   fi

    pods_count=$(kubectl get pods -n $CHATQNA_NAMESPACE -o jsonpath='{.items[*].metadata.name}' | wc -w)
    # expected_ready_pods, expected_external_pods, expected_total_pods
    # pods_count-1 is to exclude the client pod in this namespace
    check_gmc_status $CHATQNA_NAMESPACE 'chatqa' $((pods_count-1)) 0 9
    if [ $? -ne 0 ]; then
       echo "GMC status is not as expected"
       exit 1
    fi

   # giving time to populating data
   sleep 90

   kubectl get pods -n $CHATQNA_NAMESPACE
   # send request to chatqnA
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
   [[ $(grep -c "[DONE]" $LOG_PATH/curl_chatqna.log) != 0 ]]; then
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

   kubectl delete deployment client-test -n $CHATQNA_NAMESPACE
   kubectl delete gmc -n $CHATQNA_NAMESPACE 'chatqa'
   echo "sleep 10s for cleaning up"
   sleep 10
   check_resource_cleared $CHATQNA_NAMESPACE
}

function validate_chatqna_with_dataprep() {
   kubectl create ns $CHATQNA_DATAPREP_NAMESPACE
   sed -i "s|namespace: chatqa|namespace: $CHATQNA_DATAPREP_NAMESPACE|g"  $(pwd)/config/samples/ChatQnA/chatQnA_dataprep_xeon.yaml
   # workaround for issue #268
   #yq -i '(.spec.nodes.root.steps[] | select ( .name == "Tgi")).internalService.config.MODEL_ID = "bigscience/bloom-560m"' $(pwd)/config/samples/ChatQnA/chatQnA_dataprep_xeon.yaml
   kubectl apply -f $(pwd)/config/samples/ChatQnA/chatQnA_dataprep_xeon.yaml

   # Wait until the router service is ready
   echo "Waiting for the chatqa router service to be ready..."
   wait_until_pod_ready "chatqna router" $CHATQNA_DATAPREP_NAMESPACE "router-service"
   output=$(kubectl get pods -n $CHATQNA_DATAPREP_NAMESPACE)
   echo $output

   # deploy client pod for testing
   kubectl create deployment client-test -n $CHATQNA_DATAPREP_NAMESPACE --image=python:3.8.13 -- sleep infinity

   # Wait until all pods are ready
   wait_until_all_pod_ready $CHATQNA_DATAPREP_NAMESPACE 300s
   if [ $? -ne 0 ]; then
       echo "Error Some pods are not ready!"
       exit 1
   fi

    pods_count=$(kubectl get pods -n $CHATQNA_DATAPREP_NAMESPACE -o jsonpath='{.items[*].metadata.name}' | wc -w)
    # expected_ready_pods, expected_external_pods, expected_total_pods
    # pods_count-1 is to exclude the client pod in this namespace
    check_gmc_status $CHATQNA_DATAPREP_NAMESPACE 'chatqa' $((pods_count-1)) 0 10
    if [ $? -ne 0 ]; then
       echo "GMC status is not as expected"
       exit 1
    fi

   # giving time to populating data
   sleep 90

   kubectl get pods -n $CHATQNA_DATAPREP_NAMESPACE
   # send request to chatqnA
   export CLIENT_POD=$(kubectl get pod -n $CHATQNA_DATAPREP_NAMESPACE -l app=client-test -o jsonpath={.items..metadata.name})
   echo "$CLIENT_POD"
   accessUrl=$(kubectl get gmc -n $CHATQNA_DATAPREP_NAMESPACE -o jsonpath="{.items[?(@.metadata.name=='chatqa')].status.accessUrl}")

   kubectl exec "$CLIENT_POD" -n $CHATQNA_DATAPREP_NAMESPACE -- curl "$accessUrl/dataprep"  -X POST  -F 'link_list=["https://raw.githubusercontent.com/opea-project/GenAIInfra/main/microservices-connector/test/data/gaudi.txt"]' -H "Content-Type: multipart/form-data" > $LOG_PATH/curl_dataprep.log
   exit_code=$?
   if [ $exit_code -ne 0 ]; then
       echo "dataprep failed, please check the logs in ${LOG_PATH}!"
       exit 1
   fi
   echo "Checking response results, make sure the output is reasonable. "
   local status=false
   if [[ -f $LOG_PATH/curl_dataprep.log ]] && \
   [[ $(grep -c "Data preparation succeeded" $LOG_PATH/curl_dataprep.log) != 0 ]]; then
       status=true
   fi
   if [ $status == false ]; then
       if [[ -f $LOG_PATH/curl_dataprep.log ]]; then
           cat $LOG_PATH/curl_dataprep.log
       fi
       echo "Response check failed, please check the logs in artifacts!"
       exit 1
   else
       echo "Response check succeed!"
   fi

   kubectl exec "$CLIENT_POD" -n $CHATQNA_DATAPREP_NAMESPACE -- curl $accessUrl  -X POST  -d '{"text":"What are the key features of Intel Gaudi?","parameters":{"max_new_tokens":17, "do_sample": true}}' -H 'Content-Type: application/json' > $LOG_PATH/curl_chatqna_dataprep.log
   exit_code=$?
   if [ $exit_code -ne 0 ]; then
       echo "chatqna failed, please check the logs in ${LOG_PATH}!"
       exit 1
   fi

   echo "Checking response results, make sure the output is reasonable. "
   local status=false
   if [[ -f $LOG_PATH/curl_chatqna_dataprep.log ]] && \
   [[ $(grep -c "[DONE]" $LOG_PATH/curl_chatqna_dataprep.log) != 0 ]]; then
       status=true
   fi
   if [ $status == false ]; then
       if [[ -f $LOG_PATH/curl_chatqna_dataprep.log ]]; then
           cat $LOG_PATH/curl_chatqna_dataprep.log
       fi
       echo "Response check failed, please check the logs in artifacts!"
       exit 1
   else
       echo "Response check succeed!"
   fi

   kubectl delete deployment client-test -n $CHATQNA_DATAPREP_NAMESPACE
   kubectl delete gmc -n $CHATQNA_DATAPREP_NAMESPACE 'chatqa'
   echo "sleep 10s for cleaning up"
   sleep 10
   check_resource_cleared $CHATQNA_DATAPREP_NAMESPACE
}

function validate_chatqna_in_switch() {
   kubectl create ns $CHATQNA_SWITCH_NAMESPACE
   sed -i "s|namespace: switch|namespace: $CHATQNA_SWITCH_NAMESPACE|g"  $(pwd)/config/samples/ChatQnA/chatQnA_switch_xeon.yaml
   # workaround for issue #268
   #yq -i '(.spec.nodes.root.steps[] | select ( .name == "Tgi")).internalService.config.MODEL_ID = "bigscience/bloom-560m"' $(pwd)/config/samples/ChatQnA/chatQnA_switch_xeon.yaml
   kubectl apply -f $(pwd)/config/samples/ChatQnA/chatQnA_switch_xeon.yaml

   # Wait until the router service is ready
   echo "Waiting for the chatqa router service to be ready..."
   wait_until_pod_ready "chatqna router" $CHATQNA_SWITCH_NAMESPACE "router-service"
   output=$(kubectl get pods -n $CHATQNA_SWITCH_NAMESPACE)
   echo $output

   # deploy client pod for testing
   kubectl create deployment client-test -n $CHATQNA_SWITCH_NAMESPACE --image=python:3.8.13 -- sleep infinity

   # Wait until all pods are ready
   wait_until_all_pod_ready $CHATQNA_SWITCH_NAMESPACE 300s
   if [ $? -ne 0 ]; then
       echo "Error Some pods are not ready!"
       exit 1
   fi

    pods_count=$(kubectl get pods -n $CHATQNA_SWITCH_NAMESPACE -o jsonpath='{.items[*].metadata.name}' | wc -w)
    # expected_ready_pods, expected_external_pods, expected_total_pods
    # pods_count-1 is to exclude the client pod in this namespace
    check_gmc_status $CHATQNA_SWITCH_NAMESPACE 'switch' $((pods_count-1)) 0 15
    if [ $? -ne 0 ]; then
       echo "GMC status is not as expected"
       exit 1
    fi

   # giving time to populating data
   sleep 90

   kubectl get pods -n $CHATQNA_SWITCH_NAMESPACE
   # send request to chatqnA
   export CLIENT_POD=$(kubectl get pod -n $CHATQNA_SWITCH_NAMESPACE -l app=client-test -o jsonpath={.items..metadata.name})
   echo "$CLIENT_POD"
   accessUrl=$(kubectl get gmc -n $CHATQNA_SWITCH_NAMESPACE -o jsonpath="{.items[?(@.metadata.name=='switch')].status.accessUrl}")

   # test the chatqna with model condition: "model-id":"intel" and "embedding-model-id":"small"
   kubectl exec "$CLIENT_POD" -n $CHATQNA_SWITCH_NAMESPACE -- curl $accessUrl  -X POST  -d '{"text":"What is the revenue of Nike in 2023?", "model-id":"intel", "embedding-model-id":"small", "parameters":{"max_new_tokens":17, "do_sample": true}}' -H 'Content-Type: application/json' > $LOG_PATH/curl_chatqna_switch_intel.log
   exit_code=$?
   if [ $exit_code -ne 0 ]; then
       echo "chatqna failed, please check the logs in ${LOG_PATH}!"
       exit 1
   fi

   echo "Checking response results, make sure the output is reasonable. "
   local status=false
   if [[ -f $LOG_PATH/curl_chatqna_switch_intel.log ]] && \
   [[ $(grep -c "[DONE]" $LOG_PATH/curl_chatqna_switch_intel.log) != 0 ]]; then
       status=true
   fi
   if [ $status == false ]; then
       if [[ -f $LOG_PATH/curl_chatqna_switch_intel.log ]]; then
           cat $LOG_PATH/curl_chatqna_switch_intel.log
       fi
       echo "Response check failed, please check the logs in artifacts!"
       exit 1
   else
       echo "Response check succeed!"
   fi

   # test the chatqna with model condition: "model-id":"llama" and "embedding-model-id":"large"
   kubectl exec "$CLIENT_POD" -n $CHATQNA_SWITCH_NAMESPACE -- curl $accessUrl  -X POST  -d '{"text":"What is the revenue of Nike in 2023?", "model-id":"llama", "embedding-model-id":"large", "parameters":{"max_new_tokens":17, "do_sample": true}}' -H 'Content-Type: application/json' > $LOG_PATH/curl_chatqna_switch_llama.log
   exit_code=$?
   if [ $exit_code -ne 0 ]; then
       echo "chatqna failed, please check the logs in ${LOG_PATH}!"
       exit 1
   fi

   echo "Checking response results, make sure the output is reasonable. "
   local status=false
   if [[ -f $LOG_PATH/curl_chatqna_switch_llama.log ]] && \
   [[ $(grep -c "[DONE]" $LOG_PATH/curl_chatqna_switch_llama.log) != 0 ]]; then
       status=true
   fi
   if [ $status == false ]; then
       if [[ -f $LOG_PATH/curl_chatqna_switch_llama.log ]]; then
           cat $LOG_PATH/curl_chatqna_switch_llama.log
       fi
       echo "Response check failed, please check the logs in artifacts!"
       exit 1
   else
       echo "Response check succeed!"
   fi

   kubectl delete deployment client-test -n $CHATQNA_SWITCH_NAMESPACE
   kubectl delete gmc -n $CHATQNA_SWITCH_NAMESPACE 'switch'
   echo "sleep 10s for cleaning up"
   sleep 10
   check_resource_cleared $CHATQNA_SWITCH_NAMESPACE
}


function validate_modify_config() {
    kubectl create ns $MODIFY_STEP_NAMESPACE
    cp $(pwd)/config/samples/CodeGen/codegen_xeon.yaml $(pwd)/config/samples/CodeGen/codegen_xeon_mod.yaml
    sed -i "s|namespace: codegen|namespace: $MODIFY_STEP_NAMESPACE|g" $(pwd)/config/samples/CodeGen/codegen_xeon_mod.yaml
    kubectl apply -f $(pwd)/config/samples/CodeGen/codegen_xeon_mod.yaml

    # Wait until the router service is ready
    echo "Waiting for the router service to be ready..."
    wait_until_pod_ready "router" $MODIFY_STEP_NAMESPACE "router-service"
    output=$(kubectl get pods -n $MODIFY_STEP_NAMESPACE)
    echo $output

    # Wait until all pods are ready
    wait_until_all_pod_ready $MODIFY_STEP_NAMESPACE 300s
    if [ $? -ne 0 ]; then
         echo "Error Some pods are not ready!"
         exit 1
    fi

    pods_count=$(kubectl get pods -n $MODIFY_STEP_NAMESPACE -o jsonpath='{.items[*].metadata.name}' | wc -w)
    check_gmc_status $MODIFY_STEP_NAMESPACE 'codegen' $pods_count 0 3
    if [ $? -ne 0 ]; then
       echo "GMC status is not as expected"
       exit 1
    fi

    #change the model id of the step named "Tgi" in the codegen_xeon_mod.yaml
    yq -i '(.spec.nodes.root.steps[] | select ( .name == "Tgi")).internalService.config.MODEL_ID = "HuggingFaceH4/mistral-7b-grok"' $(pwd)/config/samples/CodeGen/codegen_xeon_mod.yaml
    kubectl apply -f $(pwd)/config/samples/CodeGen/codegen_xeon_mod.yaml

    # Wait until all pods are ready
    wait_until_all_pod_ready $MODIFY_STEP_NAMESPACE 300s
    if [ $? -ne 0 ]; then
         echo "Error Some pods are not ready!"
         exit 1
    fi

    pods_count=$(kubectl get pods -n $MODIFY_STEP_NAMESPACE -o jsonpath='{.items[*].metadata.name}' | wc -w)
    check_gmc_status $MODIFY_STEP_NAMESPACE 'codegen' $pods_count 0 3
    if [ $? -ne 0 ]; then
       echo "GMC status is not as expected"
       exit 1
    fi

   #revert the codegen yaml
   sed -i "s|namespace: $MODIFY_STEP_NAMESPACE|namespace: codegen|g"  $(pwd)/config/samples/CodeGen/codegen_xeon_mod.yaml
   kubectl delete gmc -n $MODIFY_STEP_NAMESPACE 'codegen'
   echo "sleep 10s for cleaning up"
   sleep 10
   check_resource_cleared $MODIFY_STEP_NAMESPACE
}

function validate_remove_step() {
    kubectl create ns $DELETE_STEP_NAMESPACE
    cp $(pwd)/config/samples/CodeGen/codegen_xeon.yaml $(pwd)/config/samples/CodeGen/codegen_xeon_del.yaml
    sed -i "s|namespace: codegen|namespace: $DELETE_STEP_NAMESPACE|g"  $(pwd)/config/samples/CodeGen/codegen_xeon_del.yaml
    kubectl apply -f $(pwd)/config/samples/CodeGen/codegen_xeon_del.yaml

    # Wait until the router service is ready
    echo "Waiting for the router service to be ready..."
    wait_until_pod_ready "router" $DELETE_STEP_NAMESPACE "router-service"
    output=$(kubectl get pods -n $DELETE_STEP_NAMESPACE)
    echo $output

    # Wait until all pods are ready
    wait_until_all_pod_ready $DELETE_STEP_NAMESPACE 300s
    if [ $? -ne 0 ]; then
         echo "Error Some pods are not ready!"
         exit 1
    fi

    pods_count=$(kubectl get pods -n $DELETE_STEP_NAMESPACE -o jsonpath='{.items[*].metadata.name}' | wc -w)
    check_gmc_status $DELETE_STEP_NAMESPACE 'codegen' $pods_count 0 3
    if [ $? -ne 0 ]; then
       echo "GMC status is not as expected"
       exit 1
    fi

    # remove the step named "llm" in the codegen_xeon.yaml
    yq -i 'del(.spec.nodes.root.steps[] | select ( .name == "Llm"))' $(pwd)/config/samples/CodeGen/codegen_xeon_del.yaml
    kubectl apply -f $(pwd)/config/samples/CodeGen/codegen_xeon_del.yaml

    sleep 10
    check_pod_terminated $DELETE_STEP_NAMESPACE

    check_gmc_status $DELETE_STEP_NAMESPACE 'codegen' 2 0 2
    if [ $? -ne 0 ]; then
       echo "GMC status is not as expected"
       exit 1
    fi

   #revert the codegen yaml
   sed -i "s|namespace: $DELETE_STEP_NAMESPACE|namespace: codegen|g"  $(pwd)/config/samples/CodeGen/codegen_xeon_del.yaml
   kubectl delete gmc -n $DELETE_STEP_NAMESPACE 'codegen'
   echo "sleep 10s for cleaning up"
   sleep 10
   check_resource_cleared $DELETE_STEP_NAMESPACE
}

function validate_codegen() {
   kubectl create ns $CODEGEN_NAMESPACE
   sed -i "s|namespace: codegen|namespace: $CODEGEN_NAMESPACE|g"  $(pwd)/config/samples/CodeGen/codegen_xeon.yaml
   kubectl apply -f $(pwd)/config/samples/CodeGen/codegen_xeon.yaml

   # Wait until the router service is ready
   echo "Waiting for the codegen router service to be ready..."
   wait_until_pod_ready "codegen router" $CODEGEN_NAMESPACE "router-service"
   output=$(kubectl get pods -n $CODEGEN_NAMESPACE)
   echo $output

   # deploy client pod for testing
   kubectl create deployment client-test -n $CODEGEN_NAMESPACE --image=python:3.8.13 -- sleep infinity

   # Wait until all pods are ready
   wait_until_all_pod_ready $CODEGEN_NAMESPACE 300s
   if [ $? -ne 0 ]; then
       echo "Error Some pods are not ready!"
       exit 1
   fi

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
   sed -i "s|namespace: codetrans|namespace: $CODETRANS_NAMESPACE|g"  $(pwd)/config/samples/CodeTrans/codetrans_xeon.yaml
   kubectl apply -f $(pwd)/config/samples/CodeTrans/codetrans_xeon.yaml

   # Wait until the router service is ready
   echo "Waiting for the codetrans router service to be ready..."
   wait_until_pod_ready "codetrans router" $CODETRANS_NAMESPACE "router-service"
   output=$(kubectl get pods -n $CODETRANS_NAMESPACE)
   echo $output

   # deploy client pod for testing
   kubectl create deployment client-test -n $CODETRANS_NAMESPACE --image=python:3.8.13 -- sleep infinity

   # Wait until all pods are ready
   wait_until_all_pod_ready $CODETRANS_NAMESPACE 300s
   if [ $? -ne 0 ]; then
       echo "Error Some pods are not ready!"
       exit 1
   fi

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
   sed -i "s|namespace: docsum|namespace: $DOCSUM_NAMESPACE|g"  $(pwd)/config/samples/DocSum/docsum_xeon.yaml
   kubectl apply -f $(pwd)/config/samples/DocSum/docsum_xeon.yaml

   # Wait until the router service is ready
   echo "Waiting for the docsum router service to be ready..."
   wait_until_pod_ready "docsum router" $DOCSUM_NAMESPACE "router-service"
   output=$(kubectl get pods -n $DOCSUM_NAMESPACE)
   echo $output

   # deploy client pod for testing
   kubectl create deployment client-test -n $DOCSUM_NAMESPACE --image=python:3.8.13 -- sleep infinity

   # Wait until all pods are ready
   wait_until_all_pod_ready $DOCSUM_NAMESPACE 300s
   if [ $? -ne 0 ]; then
       echo "Error Some pods are not ready!"
       exit 1
   fi

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
