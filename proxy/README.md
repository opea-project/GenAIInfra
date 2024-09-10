# OPEA Pipeline Proxy

OPEA Pipeline Proxy is an enhancement of the default Istio proxy with additional features designed specifically for OPEA RAG pipelines.

## Features

- [Guardrails](./deployments/guardrails/README.md)

## Build

OPEA Pipeline Proxy is based on Istio proxy and Envoy, you can refer to [Building Envoy with Bazel](https://github.com/envoyproxy/envoy/blob/main/bazel/README.md) for build locally. In addition, Clang and OpenVINO is required to build OPEA Pipeline Proxy.

```sh
# Build OPEA Pipeline Proxy. The binary will be generated to `bin/envoy`.
make

# Build the image. The image will be tagged as `opea/proxyv2:<ISTIO_TAG>` by default.
make image
```

You can also build OPEA Pipeline Proxy in the build container.

```sh
# Build the build image.
make build-image

# Build OPEA Pipeline Proxy with the build container. The binary will be generated to `bin/envoy`.
BUILD_WITH_CONTAINER=1 make

# Build the image. The image will be tagged as `opea/proxyv2:<ISTIO_TAG>` by default.
make image
```

## Deployment

Before deploying OPEA Pipeline Proxy, you have to install Istio. Please follow the steps [here](https://istio.io/latest/docs/setup/install/istioctl/) for Istio installation.

During the installation, you have to assign the OPEA Pipeline Proxy to deploy instead of the default one delivered by Istio.

```sh
# Use the default Istio pilot and the proxyv2 delivered by OPEA.
istioctl install --set hub=docker.io/opea --set components.pilot.hub=docker.io/istio
```

You can also use the annotation [here](https://istio.io/latest/docs/reference/config/annotations/#SidecarProxyImage) to inject OPEA Pipeline Proxy as sidecars.

## Development

You can generate the [JSON Compilation Database](https://clang.llvm.org/docs/JSONCompilationDatabase.html) for Visual Studio Code with [clangd](https://marketplace.visualstudio.com/items?itemName=llvm-vs-code-extensions.vscode-clangd) extension and other compatible tools.

```sh
make compilation-database
```

You can test OPEA Pipeline Proxy with the following command.

```sh
make test
```
