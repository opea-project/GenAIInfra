# Deployment of the Helm Charts on Intel® Xeon® Processors with Intel® Trust Domain Extensions (Intel® TDX)

This document outlines the deployment process of Helm Charts on Intel® Xeon® Processors where the microservices are protected by [Intel TDX](https://www.intel.com/content/www/us/en/developer/tools/trust-domain-extensions/overview.html) with the help of Confidential Containers.

## Technical background

[Intel Trust Domain Extensions (Intel TDX)](https://www.intel.com/content/www/us/en/developer/tools/trust-domain-extensions/overview.html) is hardware-based trusted execution environment (TEE) that allows the deployment of hardware-isolated virtual machines (VM) designed to protect sensitive data and applications from unauthorized access.

[Confidential Containers](https://confidentialcontainers.org/docs/overview/) encapsulates pods inside confidential virtual machines, allowing Cloud Native workloads to leverage confidential computing hardware with minimal modification.

## Prerequisites

### System Requirements

| Category           | Details                                                                                  |
| ------------------ | ---------------------------------------------------------------------------------------- |
| Operating System   | Ubuntu 24.04                                                                             |
| Hardware Platforms | 4th Gen Intel® Xeon® Scalable processors<br>5th Gen Intel® Xeon® Scalable processors |
| Kubernetes Version | 1.29+                                                                                    |

This guide assumes that:

1. you are familiar with the regular deployment of the GenAIExamples using [Helm Charts](../README.md),
2. you have prepared a server with 4th Gen Intel® Xeon® Scalable Processor or later,
3. you have a single-node Kubernetes cluster already set up on the server for the regular deployment of the GenAIExamples.

## Getting Started

### Prepare Intel Xeon node

Follow the below steps on the server node with Intel Xeon Processor:

1. [Install Ubuntu 24.04 and enable Intel TDX](https://github.com/canonical/tdx/blob/noble-24.04/README.md#setup-host-os)
2. Check, if Intel TDX is enabled:

   ```bash
   sudo dmesg | grep -i tdx
   ```

   The output should show the Intel TDX module version and initialization status:

   ```text
   virt/tdx: TDX module: attributes 0x0, vendor_id 0x8086, major_version 1, minor_version 5, build_date 20240129, build_num 698
   (...)
   virt/tdx: module initialized
   ```

   In case the module version or `build_num` is lower than shown above, please refer to the [Intel TDX documentation](https://cc-enabling.trustedservices.intel.com/intel-tdx-enabling-guide/04/hardware_setup/#deploy-specific-intel-tdx-module-version) for update instructions.

3. Depending on the location of your kubelet config file, increase the kubelet timeout and wait until the node is `Ready`:

   ```bash
   # Kubespray installation
   echo "runtimeRequestTimeout: 30m" | sudo tee -a /etc/kubernetes/kubelet-config.yaml > /dev/null 2>&1
   # Vanilla Kubernetes installation
   sudo sed -i 's/runtimeRequestTimeout: .*/runtimeRequestTimeout: 30m/' /var/lib/kubelet/config.yaml > /dev/null 2>&1
   # Restart kubelet and wait for the node to be ready
   sudo systemctl daemon-reload && sudo systemctl restart kubelet
   kubectl wait --for=condition=Ready node --all --timeout=2m
   ```

### Prepare the cluster

Follow the steps below on the Kubernetes cluster:

1. [Install Confidential Containers Operator](https://cc-enabling.trustedservices.intel.com/intel-confidential-containers-guide/02/infrastructure_setup/#install-confidential-containers-operator)

### Deploy the ChatQnA

Follow the steps below to deploy ChatQnA:

1. Set the environment variables and update the dependencies:

   ```
   export HFTOKEN="insert-your-huggingface-token-here"
   export MODELNAME="Intel/neural-chat-7b-v3-3"
   export myrelease=chatqna
   export chartname=chatqna
   ./update_dependency.sh
   helm dependency update $chartname
   ```

2. Deploy the Helm Chart setting the `tdxEnabled` flag for each microservice you want to run using Intel TDX, for example:

   ```
   helm install $myrelease $chartname \
      --set global.HUGGINGFACEHUB_API_TOKEN="${HFTOKEN}" --set vllm.LLM_MODEL_ID="${MODELNAME}" \
      --set redis-vector-db.tdxEnabled=true --set redis-vector-db.resources.limits.memory=4Gi \
      --set retriever-usvc.tdxEnabled=true --set retriever-usvc.resources.limits.memory=7Gi \
      --set tei.tdxEnabled=true --set tei.resources.limits.memory=4Gi \
      --set teirerank.tdxEnabled=true --set teirerank.resources.limits.memory=6Gi \
      --set nginx.tdxEnabled=true \
      --set chatqna-ui.tdxEnabled=true --set chatqna-ui.resources.limits.memory=2Gi \
      --set data-prep.tdxEnabled=true --set data-prep.resources.limits.memory=11Gi \
      --set vllm.tdxEnabled=true --set vllm.resources.limits.memory=80Gi
   ```

> [!NOTE]
> The `resources.limits` needs to be set when the Intel TDX is used.
>
> By default, each Kubernetes pod will be assigned `1` CPU and `2Gi` of memory, but half of it will be used for filesystem.
>
> If the pods fail to start due to lack of disk space, increase the memory limits.
