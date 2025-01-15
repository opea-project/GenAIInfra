# OPEA with Istio

## Introduction

Istio service mesh provides many features including 1) [mTLS between Kubernetes pods](#enforce-mtls-between-opea-pods) and 2) [TLS connection to Kubernetes ingress](#create-istio-gateway-with-tls-and-virtual-service). 

This document describes how to enable the above two Istio features with OPEA applications. We will use the new Istio ambient mode (a.k.a. sidecar-less mode)

## Deployment

In this document we use the following components:

- OPEA ChatQnA as an exmaple application
- Istio (in ambient mode) with ingress gateway using TLS and strict mTLS for ChatQnA application
- Cert-Manager for issuing TLS certificate to Istio ingress gateway

### Deploy Istio, ChatQnA and Cert-Manager

In this document we use [helmfile](https://helmfile.readthedocs.io/en/latest/) to do the deployment:

```bash
helmfile apply
```
> [!NOTE]
> The above deployment uses `model-volume` Persistent Volume Claim (PVC) for storing the ChatQnA models so ensure such PVC and corresponding PV are available in your cluster.

### Install Kubernetes Gateway CRDs

Note that the Kubernetes Gateway API CRDs do not come installed by default on most Kubernetes clusters, so make sure they are installed before using the Gateway API:

```bash
kubectl get crd gateways.gateway.networking.k8s.io &> /dev/null || \
  { kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.2.0/standard-install.yaml; }
```

## Create Istio gateway with TLS and virtual service

Istio gateway terminates the external TLS connections. Istio virtual service routes the traffic to services. In this example, all the traffic that matches host '*.intel.com' and path prefix '/' is routed to 'chatqna-nginx.chatqna' service. The Istio gateway needs certificate which is created via 'cert-manager' Issuer and Certificate.

Create Istio gateway and virtual service:

```bash
kubectl apply -f istio-gateway-and-virtual-service.yaml
```

Create cert-manager Issuer and Certificate:

```bash
kubectl apply -f istio-gateway-ca-and-cert.yaml
```

Now you are able to connect to OPEA engine services via TLS. You can test the connection with the command:

```bash
# Get Istio ingress loadbalancer (LB) address. If you don't use LB, you can set use `kubectl port-forward` command. 
IP=$(kubectl get svc -n istio-ingress -ojsonpath="{.items[0].status.loadBalancer.ingress[0].ip}")
# Resolve IP to DNS. DNS needs to match the dnsNames in istio-gateway certificate.
curl -ks https://${DNS}/v1/chatqna -H "Content-Type: application/json" -d '{"messages": "What is the TLS?"}'
```
> [!NOTE]
> `https` scheme (TLS) is used and in curl we ignore the server's self signed certificate with `-k` option.


## Enforce mTLS between OPEA pods

This task ensures the OPEA workloads only communicate using mutual TLS.

```bash
kubectl apply -f istio-mtls-strict.yaml -n chatqna
```

## Cleanup

Once you are done with the example you can cleanup yuor environment with the following commands:

```bash
kubectl delete -f istio-gateway-and-virtual-service.yaml
kubectl delete -f istio-gateway-ca-and-cert.yaml
kubectl apply -f istio-mtls-strict.yaml -n chatqna
helmfile delete
```
