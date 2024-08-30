# OPEA Pipeline Proxy

OPEA Pipeline Proxy is a replacement of the default Istio proxy with additional features designed specifically for OPEA RAG pipelines.

## Features

- Guardrails HTTP filter

## Build

OPEA Pipeline Proxy is based on Istio proxy and Envoy, you can refer to [Building Envoy with Bazel](https://github.com/envoyproxy/envoy/blob/main/bazel/README.md) for build locally. In addition, Clang 14 and OpenVINO is required to build OPEA Pipeline Proxy.

You can also build OPEA Pipeline Proxy in the build container.

```sh
# Build the binary. The binary will be generated to `bin/envoy`.
tools/build.sh

# Build the image. The image will be tagged as `opea/proxyv2:<ISTIO_TAG>`.
BUILD_IMAGE=1 tools/build.sh
```

## Deployment

Before deploying OPEA Pipeline Proxy, you have to install Istio. Please follow the steps [here](https://istio.io/latest/docs/setup/install/istioctl/) for Istio installation.

During the installation, you have to assign the OPEA Pipeline Proxy to deploy instead of the default one delivered by Istio.

```sh
istioctl install --set hub=docker.io/opea --set components.pilot.hub=docker.io/istio
```

You can also use the annotation [here](https://istio.io/latest/docs/reference/config/annotations/#SidecarProxyImage) to inject OPEA Pipeline Proxy as sidecars.

## Tasks

- [Guardrails](./deployments/guardrails/README.md)
