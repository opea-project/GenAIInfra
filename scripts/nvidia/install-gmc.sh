#!/usr/bin/env bash
# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

set -e

SCRIPT_DIR=$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)
cd $SCRIPT_DIR && cd ../../
GenAIInfra_DIR=$(pwd)
cd $GenAIInfra_DIR/microservices-connector/helm

if [ -n "$YOUR_HF_TOKEN" ]; then
    find manifests_common/ -name '*.yaml' -type f -exec sed -i "s#insert-your-huggingface-token-here#$YOUR_HF_TOKEN#g" {} \;
fi

if [ -n "$YOUR_GOOGLE_API_KEY" ]; then
    find manifests_common/ -name '*.yaml' -type f -exec sed -i "s#GOOGLE_API_KEY:.*#GOOGLE_API_KEY: "$YOUR_GOOGLE_API_KEY"#g" {} \;
fi

if [ -n "$YOUR_GOOGLE_CSE_ID" ]; then
    find manifests_common/ -name '*.yaml' -type f -exec sed -i "s#GOOGLE_CSE_ID:.*#GOOGLE_CSE_ID: "$YOUR_GOOGLE_CSE_ID"#g" {} \;
fi


if [ -n "$MOUNT_DIR" ]; then
    find manifests_common/ -name '*.yaml' -type f -exec sed -i "s#path: /mnt/opea-models#path: $MOUNT_DIR#g" {} \;
fi

# install GMC helm chart
helm install -n system --create-namespace gmc .
sleep 2
kubectl get pod -n system
