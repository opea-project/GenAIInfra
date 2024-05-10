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
