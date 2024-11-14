#!/bin/bash

# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

registry=registry:5000
hf_token=""
http_proxy=""
https_proxy=""

# Function to run a Docker container
run_docker_container() {
  local name=$1
  local port=$2
  local volume=$3
  local image=$4
  local model_id=$5

  # Run the Docker container
  docker stop "$name" 2>/dev/null
  docker rm "$name" 2>/dev/null
  docker run --name "$name" -d -p "$port":80 -e HF_TOKEN=$hf_token -e http_proxy=$http_proxy -e https_proxy=$https_proxy -v "$volume" "$image" --model-id "$model_id"
}

# Function to check if the service is up
check_service() {
  local name=$1
  local check_url=$2
  local check_data=$3

  # Check if the service is up
  until curl -s -o /dev/null -w "%{http_code}" -X POST -d "$check_data" -H 'Content-Type: application/json' "$check_url" | grep -q "200"; do
    echo "curl -X POST -d '$check_data' -H 'Content-Type: application/json' '$check_url'"
    echo "Waiting for $name service to be up..."
    docker logs --tail 5 "$name"
    sleep 5
  done

  echo "$name service is up"

  # Destroy the Docker container after a successful check
  docker stop "$name"
  docker rm "$name"
}

# Define the ports to use
port1=9001
port2=9002
port3=9003

# Run the Docker containers
run_docker_container "tei" "$port1" "$PWD/data:/data" "ghcr.io/huggingface/text-embeddings-inference:cpu-1.5" "BAAI/bge-base-en-v1.5"
run_docker_container "teirank" "$port2" "$PWD/data:/data" "ghcr.io/huggingface/text-embeddings-inference:cpu-1.5" "BAAI/bge-reranker-base"
run_docker_container "tgi" "$port3" "$PWD/data:/data" "ghcr.io/huggingface/text-generation-inference:2.4.0-intel-cpu" "Intel/neural-chat-7b-v3-3"

# Check the services
check_service "tei" "http://localhost:$port1/v1/embeddings" '{"input": "This is a test"}'
check_service "teirank" "http://localhost:$port2/rerank" '{"query": "What is Deep Learning?", "texts": ["Deep Learning is " ]}'
check_service "tgi" "http://localhost:$port3/v1/chat/completions" '{"messages": [{"role": "user", "content": "Say this is a test!"}]}'


docker build . -t $registry/opea/models:latest
docker push $registry/opea/models:latest
