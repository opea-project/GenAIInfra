FROM golang:1.22 AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
COPY . /workspace/
RUN go mod download

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o memory-bandwidth-exporter cmd/main.go

FROM ubuntu:22.04
USER root
WORKDIR /
COPY --from=builder /workspace/memory-bandwidth-exporter .

ENTRYPOINT  ["bash", "-c"]
CMD ["/memory-bandwidth-exporter --collector.node.name=${NODE_NAME} --collector.container.namespaceWhiteList=${NAMESPACE_WHITELIST}"]