#!/bin/bash

# Copyright (C) 2025 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ -n "$HF_TOKEN" ]; then
    echo "Using HF_TOKEN env var"
elif [ -n "$1" ] && [ -f "$1" ]; then
    echo "Using given HF token file"
    HF_TOKEN=$(cat $1)
else
    echo "Usage: [HF_TOKEN=<token>] ${0##*/} [hf-token file]"
    echo "ERROR: HF token missing!"
    exit 1
fi

helm upgrade --install opea -n kubeai kubeai/kubeai \
    --create-namespace \
    --set secrets.huggingface.token=$HF_TOKEN \
    -f $DIR/opea-values.yaml
