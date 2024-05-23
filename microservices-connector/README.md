# genai-microservices-connector(GMC)
This repo defines the GenAI Microservice Connector for OPEA projects.

## Description
The GMC contains the CRD and its controller to bring up the needed services for a GenAI application.
Istio Service Mesh can also be leveraged to facilicate communication between microservices in the GenAI application.

## Architecture

![GMC Architecture](./architecture.png)

## Personas

![GMC Personas](./personas.png)

## Getting Started

**CRD** defines are at config/crd/bases/  
**API** is api/v1alpha3/  
**Controller** is at internal/controller  

`make manifests` to install the APIs  
`make run` to lanch the controllers   

### Prerequisites
- go version v1.21.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.


### Introduction

There are `2` components in this repo:

- 1. `manager`: controller manager to handle GMC CRD
- 2. `router`: route the traffic among the microservices defined in GMC connector

#### How to build these binaries?

```sh
make build
```

#### How to build docker images for these 2 components?

```sh
make docker
```

#### How to delete these components' binaries?

```sh
make clean 
```

### To Deploy on the cluster
**Build and push your image to the location specified by `CTR_IMG`:**

```sh
make docker-build docker-push CTR_IMG=<some-registry>/gmconnector:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `CTR_IMG`:**

```sh
make deploy CTR_IMG=<some-registry>/gmconnector:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall
**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Project Distribution

Following are the steps to build the installer and distribute this project to users.

1. Build the installer for the image built and published in the registry:

```sh
make build-installer CTR_IMG=<some-registry>/gmconnector:tag
```

NOTE: The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without
its dependencies.

2. Using the installer

Users can just run kubectl apply -f <URL for YAML BUNDLE> to install the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/gmconnector/<tag or branch>/dist/install.yaml
```


