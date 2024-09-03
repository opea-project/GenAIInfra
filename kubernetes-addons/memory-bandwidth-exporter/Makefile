GO_CMD     := go
DEBUG ?= 0
DOCKER_REGISTRY ?= docker.io/opea
CONTAINER_TOOL ?= docker
VERSION ?= latest

ifeq ($(DEBUG),0)
	GOFLAGS=-ldflags="all=-s -w"
endif

MBE_IMG_NAME = memory-bandwidth-exporter:$(VERSION)
MBE_IMG = ${DOCKER_REGISTRY}/${MBE_IMG_NAME}

build: 
	@mkdir -p bin
	@echo "Building memory-bandwidth-exporter binary..."
	$(GO_CMD) build -o bin/memory-bandwidth-exporter  $(GOFLAGS) cmd/main.go

docker.build:
	@echo "Building memory-bandwidth-exporter Docker image..."
	$(CONTAINER_TOOL) build -t ${MBE_IMG} -f Dockerfile .

docker.push:
	@echo "Push memory-bandwidth-exporter Docker image..."
	$(CONTAINER_TOOL) push ${MBE_IMG}

clean:
	@echo "Cleaning up..."
	rm -rf bin

change_img:
	sed -i "s\MBE_IMG\${MBE_IMG}\g" config/manifests/memory-bandwidth-exporter.yaml

test:
	@echo "Running tests..."
	$(GO_CMD) test ./...