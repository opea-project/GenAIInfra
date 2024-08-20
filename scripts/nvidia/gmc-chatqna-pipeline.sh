#!/usr/bin/env bash
# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

set -e

SCRIPT_DIR=$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)
cd $SCRIPT_DIR && cd ../../
GenAIInfra_DIR=$(pwd)
cd $GenAIInfra_DIR/microservices-connector/

# TODO: to support more examples
kubectl create ns chatqa
kubectl apply -f $(pwd)/config/samples/chatQnA_nv.yaml

sleep 2
kubectl get service -n chatqa
kubectl create deployment client-test -n chatqa --image=python:3.8.13 -- sleep infinity