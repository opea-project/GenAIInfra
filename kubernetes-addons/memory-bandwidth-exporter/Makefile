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

lint: golangci-lint ## Run golangci-lint linter & yamllint
	@echo "Running linters...${GOLANGCI_LINT}"
	$(GOLANGCI_LINT) run ./...

lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes
	$(GOLANGCI_LINT) run --fix ./...

##@ Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint-$(GOLANGCI_LINT_VERSION)

## Tool Versions
GOLANGCI_LINT_VERSION ?= v1.59.1

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/cmd/golangci-lint,${GOLANGCI_LINT_VERSION})

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary (ideally with version)
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f $(1) ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
mv "$$(echo "$(1)" | sed "s/-$(3)$$//")" $(1) ;\
}
endef