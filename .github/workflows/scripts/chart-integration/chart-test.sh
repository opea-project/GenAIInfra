#!/bin/bash
# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

set -x

# get the current user id
# get the current user name
export USER_ID=$(whoami)
export PATH=/home/$USER_ID/.local/bin:$PATH # add huggingface-cli to PATH

CHART_NAME=$1
ROLLOUT_TIMEOUT_SECONDS=600s
KUBECTL_TIMEOUT_SECONDS=60s
RELEASE_NAME=${CHART_NAME}$(date +%Y%m%d%H%M%S)
NAMESPACE=${CHART_NAME}-$(date +%Y%m%d%H%M%S)

SCRIPT_DIR=$(dirname -- $0)
source $SCRIPT_DIR/lib.sh

# step 1 Run init
echo "Testing chart $CHART_NAME, init environment..."
# call init_chart function
init_${CHART_NAME}
echo "Testing chart $CHART_NAME, init environment done!"

# step 2 Helm install
echo "Testing chart $CHART_NAME, installing chart with helm..."
if helm install --create-namespace --namespace $NAMESPACE --wait --timeout "$ROLLOUT_TIMEOUT_SECONDS" $RELEASE_NAME .; then
    echo "Testing chart $CHART_NAME, installing chart with helm done!"
    return_code=0
else
    echo "Failed to install chart $CHART_NAME!"
    return_code=1
fi

# step 3 Validate
# if return_code is 0, then validate, call validate_${CHART_NAME}
if [ $return_code -eq 0 ]; then
    echo "Testing chart $CHART_NAME, validating..."
    # check the output of validate_${CHART_NAME} to determine if the validation is successful
    # if the output of validate_${CHART_NAME} contains "Response check succeed!", then the validation is successful
    output=$(validate_${CHART_NAME})
    if [[ $output == *"succeed!"* ]]; then
        echo "Testing chart $CHART_NAME, validating done!"
    else
        echo "Testing chart $CHART_NAME, validating failed!"
        return_code=1
    fi
fi

# step 4 Clean up
echo "Testing chart $CHART_NAME, cleaning up..."
helm uninstall $RELEASE_NAME --namespace $NAMESPACE
if ! kubectl delete ns $NAMESPACE --timeout=$KUBECTL_TIMEOUT_SECONDS; then
    # if kubectl delete timeout, force delete
    kubectl delete pods --namespace $NAMESPACE --force --grace-period=0 --all
    kubectl delete ns $NAMESPACE --force --grace-period=0 --timeout=$KUBECTL_TIMEOUT_SECONDS
fi
echo "Testing chart $CHART_NAME, cleaning up done!"

exit $return_code
