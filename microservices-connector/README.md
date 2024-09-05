# genai-microservices-connector(GMC)

This repo defines the GenAI Microservice Connector(GMC) for OPEA projects. GMC can be used to compose and adjust GenAI pipelines dynamically
on kubernetes. It can leverage the microservices provided by [GenAIComps](https://github.com/opea-project/GenAIComps) and external services to compose GenAI pipelines. External services might be running in a public cloud or on-prem by providing an URL and access details such as an API key and ensuring there is network connectivity. It also allows users to adjust the pipeline on the fly like switching to a different Large language Model(LLM), adding new functions into the chain(like adding guardrails),etc. GMC supports different types of steps in the pipeline, like sequential, parallel and conditional.

Please refer to [usage_guide](./usage_guide.md) for sample use cases.
Please refer to [chatqna_use_cases](./config/samples/ChatQnA/use_cases.md) for more ChatQnA use cases.

## Description

The GenAI Microservice Connector(GMC) contains the CustomResourceDefinition(CRD) and its controller to bring up the services needed for a GenAI application.
Istio Service Mesh can also be leveraged to facilicate communication between microservices in the GenAI application.

## Architecture

![GMC Architecture](./architecture.png)

## Personas

![GMC Personas](./personas.png)

## Getting Started

**CRD** defines are at config/crd/bases/
**API** is api/v1alpha3/
**Controller** is at internal/controller

### Prerequisites

- Access to a Kubernetes v1.11.3+ cluster.

### Introduction

There are `2` components in this repo:

- 1. `manager`: controller manager to handle GMC CRD
- 2. `router`: route the traffic among the microservices defined in GMC

### GMC build

#### Binaries building

```sh
make build
```

#### Docker images building

```sh
make docker.build
```

#### Binaries deleting

```sh
make clean
```

### GMC Deployment on K8s cluster

**GMC Images NOTE:** This image ought to be published in [dockhub](https://hub.docker.com/u/opea), including [gmcmanager](https://hub.docker.com/r/opea/gmcmanager) and [gmcrouter](https://hub.docker.com/r/opea/gmcrouter). Make sure you have the proper permission to the registry and use the latest images.

There are 2 methods for deploying GMC on K8s cluster:

- Deploy via native kubectl
- Deploy via helm chart

#### Deploy using native kubectl

There are 3 steps for deploying GMC on K8s cluster as below:

- Install GMC CRD
- Prepare GenAI Components and GMC Router manifests
- Deploy GMC Manager

**Deploy GMC NOTE:**

- Before installting the manifests, please replace your own huggingface tokens
- Please set the path to the manifests if you have pre-defined directory to save the models on you cluster hosts
- The `SYSTEM_NAMESPACE` should keep the same with the namespace defined in `gmc-manager.yaml` and `gmc-manager-rbac.yaml`
- The configmap name `gmcyaml` is defined in gmcmanager deployment Spec. Please modify accordingly if you want
  use a different name for the configmap

```sh
# Install GMC CRD
kubectl apply -f config/crd/bases/gmc.opea.io_gmconnectors.yaml
# Prepare GenAI Components and GMC Router manifests
cp $(pwd)/config/gmcrouter/gmc-router.yaml -p $(pwd)/config/manifests/
export YOUR_HF_TOKEN=<your hugging facetoken>
export MOUNT_DIR=<your model path>
find . -name '*.yaml' -type f -exec sed -i "s#insert-your-huggingface-token-here#$YOURTOKEN#g" {} \;
find . -name '*.yaml' -type f -exec sed -i "s#path: /mnt/opea-models#path: $MOUNT_DIR#g" {} \;
# Deploy GMC Manager
export SYSTEM_NAMESPACE=system
kubectl create namespace $SYSTEM_NAMESPACE
kubectl create configmap gmcyaml -n $SYSTEM_NAMESPACE --from-file $(pwd)/config/manifests
kubectl apply -f $(pwd)/config/rbac/gmc-manager-rbac.yaml
kubectl apply -f $(pwd)/config/manager/gmc-manager.yaml
```

**Check the installation result**

```sh
kubectl get pods -n system
NAME                              READY   STATUS    RESTARTS   AGE
gmc-controller-78f9c748cb-ltcdv   1/1     Running   0          3m
```

**To Uninstall via kubectl**

```sh
# Delete the instances (CRs) from the cluster
kubectl delete -k config/samples/
# Delete the APIs(CRDs) from the cluster
make uninstall
# UnDeploy the controller from the cluster
make undeploy
```

#### Deploy via helm chart

Please refer to [helm chart README](./helm/README.md) for deploying GMC using helm chart.

### Next Step

Please refer to [usage_guide](./usage_guide.md) for sample use cases.
Please refer to [advanced_use_cases](./config/samples/ChatQnA/use_cases.md) for more use cases based on ChatQnA example.

### Troubleshooting guide

Please refer to this [troubleshooting_guide](./troubleshooting_guide.md) for identifying GMC Custom Resource issues.
