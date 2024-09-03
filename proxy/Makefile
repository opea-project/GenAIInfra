# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

BUILD_IMG ?= envoy-build-ubuntu
BUILD_TAG ?= 75238004b0fcfd8a7f71d380d7a774dda5c39622
ISTIO_IMG ?= proxyv2
ISTIO_TAG ?= 1.23.0
DOCKER_REGISTRY ?= docker.io/opea

BUILD_WITH_CONTAINER ?= 0
BAZEL_ARG ?=

ifeq ($(BUILD_WITH_CONTAINER),0)

.PHONY: build
build:
	bazel build -c opt $(BAZEL_ARG) envoy
	mkdir -p bin
	cp -f bazel-bin/envoy bin/envoy

.PHONY: test
test:
	bazel test $(BAZEL_ARG) //test/...

else

.PHONY: build
build:
	docker run -it \
		--rm \
		-v /tmp/proxy-docker-build:/root/.cache \
		-v $(shell pwd):/source \
		$(DOCKER_REGISTRY)/$(BUILD_IMG):$(BUILD_TAG) \
		"/bin/bash" "-c" "cd /source && export PATH=/opt/llvm/bin:$$PATH && BAZEL_ARG=\"--config=docker\" make"

.PHONY: test
test:
	docker run -it \
		--rm \
		-v /tmp/proxy-docker-build:/root/.cache \
		-v $(shell pwd):/source \
		$(DOCKER_REGISTRY)/$(BUILD_IMG):$(BUILD_TAG) \
		"/bin/bash" "-c" "cd /source && export PATH=/opt/llvm/bin:$$PATH && BAZEL_ARG=\"--config=docker\" make test"

endif

.PHONY: image
image:
	docker build -f tools/Dockerfile -t $(DOCKER_REGISTRY)/$(ISITO_IMG):$(ISTIO_TAG)

.PHONY: build-image
build-image:
	docker build -f tools/Dockerfile-build -t $(DOCKER_REGISTRY)/$(BUILD_IMG):$(BUILD_TAG) .

.PHONY: compilation-database
compilation-database:
	$(shell bazel info output_base)/external/envoy/tools/gen_compilation_database.py --vscode //source/... //test/...
