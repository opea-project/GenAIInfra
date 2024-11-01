#!/bin/bash
# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

saved_errexit=False

function turnoff_and_save_errexit() {
    if [[ "$-" =~ e ]]; then
        saved_errexit=True
        set +e
    fi
}


function resume_errexit() {
    if [ "$saved_errexit" = "True" ]; then
        saved_errexit=False
        set -e
    fi
}


function wait_until_pod_ready() {
    echo "Waiting for the $1 to be ready..."
    max_retries=60
    retry_count=0
    while ! is_pod_ready $2 $3; do
        if [ $retry_count -ge $max_retries ]; then
            echo "$1 is not ready after waiting for a significant amount of time"
            get_gmc_controller_logs
            exit 1
        fi
        echo "$1 is not ready yet. Retrying in 30 seconds..."
        sleep 15
        output=$(kubectl get pods -n $2)
        echo $output
        retry_count=$((retry_count + 1))
    done
}

function is_pod_ready() {
    if [ "$2" == "gmc-controller" ]; then
      pod_status=$(kubectl get pods -n $1 -o jsonpath='{.items[].status.conditions[?(@.type=="Ready")].status}')
    else
      pod_status=$(kubectl get pods -n $1 -l app=$2 -o jsonpath='{.items[].status.conditions[?(@.type=="Ready")].status}')
    fi
    if [ "$pod_status" == "True" ]; then
        return 0
    else
        return 1
    fi
}

function get_gmc_controller_logs() {
    # Fetch the name of the pod with the app-name gmc-controller in the specified namespace
    pod_name=$(kubectl get pods -n $SYSTEM_NAMESPACE -l control-plane=gmc-controller -o jsonpath='{.items[0].metadata.name}')

    # Check if the pod name was found
    if [ -z "$pod_name" ]; then
        echo "No pod found with app-name gmc-controller in namespace $SYSTEM_NAMESPACE"
        return 1
    fi

    # Get the logs of the found pod
    echo "Fetching logs for pod $pod_name in namespace $SYSTEM_NAMESPACE..."
    kubectl logs $pod_name -n $SYSTEM_NAMESPACE
}


function wait_until_all_pod_ready() {
  namespace=$1
  timeout=$2

  turnoff_and_save_errexit
  echo "Wait for all pods in NS $namespace to be ready..."
  pods=$(kubectl get pods -n $namespace --no-headers | grep -v "Terminating" | awk '{print $1}')
  # Loop through each pod
  echo "$pods" | while read -r line; do
    pod_name=$line
    kubectl wait --for=condition=Ready pod/${pod_name} -n $namespace --timeout=${timeout}
    if [ $? -ne 0 ]; then
      echo "Pod $pod_name is not ready after waiting for ${timeout}"
      echo "Pod $pod_name status:"
      kubectl describe pod $pod_name -n $namespace
      echo "Pod $pod_name logs:"
      kubectl logs $pod_name -n $namespace
      exit 1
    fi
  done
  resume_errexit
}

function check_gmc_status() {
  namespace=$1
  gmc_name=$2
  expected_ready_pods=$3
  expected_external_pods=$4
  expected_total_pods=$5

  # pods*3 is because 1 pod has 1 configmap + 1 deployment + 1 service
  # minus 1 is because router and redis don't have the configmap
  expected_total_records=$((3* $3 - 2))

  if [ $((expected_ready_pods + expected_external_pods)) -ne $expected_total_pods ]; then
    return 1
  fi

  gmc_status=$(kubectl get gmc -n $namespace -o jsonpath="{.items[?(@.metadata.name=='$gmc_name')].status.status}")
  echo $gmc_status
  if [[ "$gmc_status" == "$expected_ready_pods/$expected_external_pods/$expected_total_pods" ]]; then
    return 0
  else
    return 1
  fi
  annotation=$(kubectl get gmc -n $namespace -o json | jq ".items[] | select(.metadata.name==\"$gmc_name\") | .status.annotations | length")
  echo $annotation
  if [ $annotation -eq $expected_total_records ]; then
    return 0
  else
    return 1
  fi
}

function check_resource_cleared() {
  namespace=$1

  actual_count=$(kubectl get all -n $namespace --no-headers | wc -l)
  if [ $actual_count -eq 0 ]; then
    return 0
  else
    #check every line of kubectl get all status is Terminating
    remaining=$(kubectl get pods -n $namespace --no-headers)
    echo $remaining
    status=$(echo $remaining | awk '{print $3}')
    for i in $status; do
      if [[ "$i" != "Terminating" ]]; then
        return 1
      fi
    done
    return 0
  fi
}

function check_pod_terminated() {
  namespace=$1

  #check every line of kubectl get all status is Terminating
  remaining=$(kubectl get pods -n $namespace --no-headers)
  echo $remaining
  status=$(echo $remaining | awk '{print $3}')
  for i in $status; do
    if [[ "$i" == "Terminating" ]]; then
      return 0
    fi
  done
  return 1
}
