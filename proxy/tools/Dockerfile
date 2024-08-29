ARG BUILD_TAG=75238004b0fcfd8a7f71d380d7a774dda5c39622
ARG ISTIO_TAG=1.23.0

FROM opea/envoy-build-ubuntu:${BUILD_TAG} AS build

FROM istio/proxyv2:${ISTIO_TAG}

# Install libtbb.
RUN apt update && apt install -y libtbb12

# Install OpenVINO.
COPY --from=build --chown=0:0 --chmod=644 /usr/lib/libopenvino.so* /usr/lib
COPY --from=build --chown=0:0 --chmod=644 /usr/lib/libopenvino_c.so* /usr/lib
COPY --from=build --chown=0:0 --chmod=644 /usr/lib/libopenvino_ir_frontend.so* /usr/lib
COPY --from=build --chown=0:0 --chmod=644 /usr/lib/libopenvino_intel_cpu_plugin.so /usr/lib

# Install OpenVINO Tokenizers.
COPY --from=build --chown=0:0 --chmod=644 /usr/lib/libcore_tokenizers.so /usr/lib
COPY --from=build --chown=0:0 --chmod=644 /usr/lib/libopenvino_tokenizers.so /usr/lib

# Install Envoy.
COPY bin/envoy /usr/local/bin/envoy
