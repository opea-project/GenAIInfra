# NVIDIA GPU Quick-Start Guide

Ver: 1.0  
Last Update: 2024-Aug-21  
Author: [PeterYang12](https://github.com/PeterYang12)
E-mail: yuhan.yang@intel.com

This document is a quick-start guide for GenAIInfra deployment and test on NVIDIA GPU platform.

## Prerequisite

GenAIInfra uses Kubernetes as the cloud native infrastructure. Follow these steps to prepare the Kubernetes environment.

### Setup Kubernetes cluster

Follow the [Kubernetes official setup guide](https://kubernetes.io/docs/setup/) to setup Kubernetes. We recommend you use Kubernetes with version >= 1.27.

### To run GenAIInfra on NVIDIA GPUs

To run the workloads on NVIDIA GPUs, follow these steps.

1. Check the [support matrix](https://docs.nvidia.com/ai-enterprise/latest/product-support-matrix/index.html) to make sure your environment meets the requirements.

2. [Install the NVIDIA GPU CUDA driver and software stack](https://developer.nvidia.com/cuda-downloads).

3. [Install the NVIDIA Container Toolkit](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/latest/install-guide.html)

4. [Install the NVIDIA GPU device plugin for Kubernetes](https://github.com/NVIDIA/k8s-device-plugin).
5. [Install helm](https://helm.sh/docs/intro/install/)

NOTE: Make sure you configure the appropriate container runtime based on the type of container runtime you installed during Kubernetes setup.

## Usages

### Use GenAI Microservices Connector (GMC) to deploy and adjust GenAIExamples on NVIDIA GPUs

#### 1. Install the GMC Helm Chart

**_NOTE_**: Before installing GMC, export your own huggingface tokens, Google API KEY, and Google CSE ID. If you have a pre-defined directory to save the models on you cluster hosts, also set the path.

```
export YOUR_HF_TOKEN=<your hugging facetoken>
export YOUR_GOOGLE_API_KEY=<your google api key>
export YOUR_GOOGLE_CSE_ID=<your google cse id>
export MOUNT_DIR=<your model path>
```

Here is a simple way to install GMC using helm chart `./install-gmc.sh`

> WARNING: the install-gmc.sh may fail due to OS distributions.

For more details, refer to [GMC installation](../../microservices-connector/README.md) to get more details.

#### 2.Use GMC to compose a ChatQnA Pipeline

Refer to [Usage guide for GMC](../../microservices-connector/usage_guide.md) for more details.

Here provides a simple script to use GMC to compose ChatQnA pipeline.

#### 3. Test ChatQnA service

Refer to [GMC ChatQnA test](../../microservices-connector/usage_guide.md#use-gmc-to-compose-a-chatqna-pipeline)
Here provides a simple way to test the service. `./gmc-chatqna-test.sh`

#### 4. Delete ChatQnA and GMC

```
kubectl delete ns chatqa
./delete-gmc.sh
```

## FAQ and Troubleshooting

The scripts are only tested on bare metal **Ubuntu 22.04** with **NVIDIA H100**. Report an issue if you meet any issue.
