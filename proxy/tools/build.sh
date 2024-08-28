#! /bin/bash

# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

BUILD_TAG=${BUILD_TAG:-75238004b0fcfd8a7f71d380d7a774dda5c39622}

# Build build image.
pushd tools
docker build . --build-arg="BUILD_TAG=${BUILD_TAG}" -f Dockerfile-build -t opea/envoy-build-ubuntu:${BUILD_TAG}
popd

# Build proxy in the container.
mkdir -p /tmp/proxy-docker-build
docker run -it \
           --rm \
           -v /tmp/proxy-docker-build:/root/.cache \
           -v $PWD:/source \
           opea/envoy-build-ubuntu:${BUILD_TAG} \
           "/bin/bash" "-c" "cd /source && export PATH=/opt/llvm/bin:\$PATH && bazel build --config docker -c opt envoy && mkdir -p bin && cp -f bazel-bin/envoy bin/envoy"
