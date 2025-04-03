#!/bin/bash

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

helm upgrade --install opea -n kubeai kubeai/kubeai \
    --create-namespace \
    --set secrets.huggingface.token=$HF_TOKEN \
    -f $DIR/opea-values.yaml