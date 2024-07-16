# GenAIInfra

GenAIInfra is the containerization and cloud native suite for OPEA, including artifacts to deploy [GenAIExamples](https://github.com/opea-project/GenAIExamples) in a cloud native way, which can be used by enterprise users to deploy to their own cloud.

## Overview

The GenAIInfra repository is organized under four main directories, which include artifacts for OPEA deploying:

| Directory           | Purpose                                                                                                                     |
| ------------------- | --------------------------------------------------------------------------------------------------------------------------- |
| `microservices-connector`       | GenAI Microservices Connector(GMC) for deploying and adjusting[GenAIExamples](https://github.com/opea-project/GenAIExamples) on Kubernetes. |
| `helm-charts`       | Helm charts for deploying [GenAIComponents](https://github.com/opea-project/GenAIComps) on Kubernetes.                     |
| `kubernetes-addons` | Deploy Kubernetes add-ons for OPEA.                                                                                         |
| `manifests`         | Manifests for deploying [GenAIComponents](https://github.com/opea-project/GenAIComps) on Kubernetes and on Docker Compose. |
| `scripts`           | Scripts for testing, tools for facilitate OPEA deployment, and etc.                                                         |

## Prerequisite

GenAIInfra uses Kubernetes as the cloud native infrastructure. Please follow the steps below to prepare the Kubernetes environment.

### Setup Kubernetes cluster

Please follow [Kubernetes official setup guide](https://kubernetes.io/docs/setup/) to setup Kubernetes. We recommend to use Kubernetes with version >= 1.27.

There are different methods to setup Kubernetes production cluster, such as [kubeadm](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/), [kubespray](https://kubespray.io/), and [more](https://kubernetes.io/docs/setup/production-environment/tools/).

NOTE: We recommend to use containerd when choosing the container runtime during Kubernetes setup. Docker engine is also verified on Ubuntu 22.04 and above.

### (Optional) To run GenAIInfra on [Intel Gaudi](https://habana.ai/products/) product:

The following steps are optional. They're only required if you want to run the workloads on Intel Gaudi product.

1. Please check the [support matrix](https://docs.habana.ai/en/latest/Support_Matrix/Support_Matrix.html) to make sure that environment meets the requirements.

2. [Install Intel Gaudi software stack](https://docs.habana.ai/en/latest/Installation_Guide/Bare_Metal_Fresh_OS.html#driver-fw-install-bare).

3. [Install and setup container runtime](https://docs.habana.ai/en/latest/Installation_Guide/Bare_Metal_Fresh_OS.html#set-up-container-usage), based on the container runtime used by Kubernetes.

NOTE: Please make sure you configure the appropriate container runtime based on the type of container runtime you installed during Kubernetes setup.

4. [Install Intel Gaudi device plugin for Kubernetes](https://docs.habana.ai/en/latest/Orchestration/Gaudi_Kubernetes/Device_Plugin_for_Kubernetes.html).

## Usages

### Use GenAI Microservices Connector(GMC) to deploy and adjust GenAIExamples

Follow [GMC README](https://github.com/opea-project/GenAIInfra/blob/main/microservices-connector/README.md)
to install GMC into your kubernetes cluster. GMC provides sample config for GenAIExamples under `microservices-connector/config/samples`. Select the example workflow you want to deploy, deploy the example (e.g. DocSum) using `kubectl`:

```shell
kubectl apply -f ./microservices-connector/config/samples/docsum_xeon.yaml
```

### Use helm charts to deploy

To deploy GenAIExamples to Kubernetes using helm charts, you need [Helm](https://helm.sh/docs/intro/install/) installed on your machine.

Clone the GenAIInfra repository and change into the `helm-charts` directory:

```shell
git clone https://github.com/opea-project/GenAIInfra.git
cd GenAIInfra/helm-charts
```

Select the example workflow you want to deploy, set the customized values in `values.yaml` and deploy the example (e.g. codegen) using `helm`:

```shell
helm install codegen ./codegen
```

### Use manifests to deploy

GenAIInfra also supports deploy GenAIExamples using manifests, you need [kubectl](https://kubernetes.io/docs/tasks/tools) installed on your machine.

Clone the GenAIInfra repository and change into the `manifests` directory:

```shell
git clone https://github.com/opea-project/GenAIInfra.git
cd GenAIInfra/manifests
```

Select the example workflow you want to deploy, deploy the example (e.g. DocSum) using `kubectl`:

```shell
kubectl apply -f ./DocSum/manifests/
```

## Additional Content

- [Code of Conduct](https://github.com/opea-project/docs/tree/main/community/CODE_OF_CONDUCT.md)
- [Contribution](https://github.com/opea-project/docs/tree/main/community/CONTRIBUTING.md)
- [Security Policy](https://github.com/opea-project/docs/tree/main/community/SECURITY.md)
- [Legal Information](/LEGAL_INFORMATION.md)
