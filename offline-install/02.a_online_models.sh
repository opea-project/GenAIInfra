#!/bin/bash

# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

registry=registry:5000
hf_token=""
http_proxy=""
https_proxy=""

python3 -m venv venv
source venv/bin/activate
pip install -U "huggingface_hub[cli]"
export HF_ENDPOINT=https://hf-mirror.com
huggingface-cli login --token $hf_token
huggingface-cli download BAAI/bge-base-en-v1.5 --local-dir cli.data/models--BAAI--bge-base-en-v1.5
huggingface-cli download BAAI/bge-reranker-base --local-dir cli.data/models--BAAI--bge-reranker-base
huggingface-cli download Intel/neural-chat-7b-v3-3 --local-dir cli.data/models--Intel--neural-chat-7b-v3-3

#docker build . -t $registry/opea/models:latest
#docker push $registry/opea/models:latest
