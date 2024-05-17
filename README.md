# GenAIInfra

GenAIInfra is the containerization and cloud native suite for OPEA, including artifacts to deploy [GenAIExamples](https://github.com/opea-project/GenAIExamples) in a cloud native way, which can be used by enterprise users to deploy to their own cloud.

## Overview

The GenAIInfra repository is organized under four main directories, which include artifacts for OPEA deploying:

| Directory           | Purpose                                                                                                                     |
| ------------------- | --------------------------------------------------------------------------------------------------------------------------- |
| `helm-charts`       | Helm charts for deploying [GenAIExamples](https://github.com/opea-project/GenAIExamples) on Kubernetes.                     |
| `kubernetes-addons` | Deploy Kubernetes add-ons for OPEA.                                                                                         |
| `manifests`         | Manifests for deploying [GenAIExamples](https://github.com/opea-project/GenAIExamples) on Kubernetes and on Docker Compose. |
| `scripts`           | Scripts for testing, tools for facilitate OPEA deployment, and etc.                                                         |

## Usages

### Prerequisites

GenAIInfra uses [Kubernetes](https://kubernetes.io/) as the cloud native infrastructure. You will need the access to either managed kubernetes services (e.g., EKS,ASK,GKE and etc) or self managed kubernetes.

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
