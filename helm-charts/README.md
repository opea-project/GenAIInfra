# Helm charts for deploying GenAI Components and Examples

This directory contains Helm charts for [GenAIComps](https://github.com/opea-project/GenAIComps) and [GenAIExamples](https://github.com/opea-project/GenAIExamples) deployment on Kubernetes.

## Table of Contents

- [Helm Charts](#helm-charts)
  - [Examples](#examples)
  - [Components](#components)
- [Deploy with Helm charts](#deploy-with-helm-charts)
  - [From Source Code](#from-source-code)
  - [Using Helm Charts repository](#using-helm-charts-repository)
- [Helm Charts Options](#helm-charts-options)
- [Using HPA (autoscaling)](#using-hpa-autoscaling)
- [Using Persistent Volume](#using-persistent-volume)
- [Using Private Docker Hub](#using-private-docker-hub)
- [Generate manifests from Helm Charts](#generate-manifests-from-helm-charts)

## Helm Charts

List of supported workloads and components.

### Examples

AI application examples you can run directly on Xeon and Gaudi. You can also refer to these examples to develop your own customized AI application.

| Helm chart                         | Link to GenAIExamples                                                                              | Description                                                                                     |
| ---------------------------------- | -------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------- |
| [chatqna](./chatqna/README.md)     | [ChatQnA](https://github.com/opea-project/GenAIExamples/tree/main/ChatQnA/README.md)               | An example of chatbot for question and answering through retrieval argumented generation (RAG). |
| [agentqna](./agentqna/README.md)   | [Agent QnA](https://github.com/opea-project/GenAIExamples/tree/main/AgentQnA/README.md)            | A hierarchical multi-agent system for question-answering applications.                          |
| [audioqna](./audioqna/README.md)   | [Audio QnA](https://github.com/opea-project/GenAIExamples/tree/main/AudioQnA/README.md)            | An example of chatbot for question and answering with audio file support.                       |
| [codegen](./codegen/README.md)     | [Code Generation](https://github.com/opea-project/GenAIExamples/tree/main/CodeGen/README.md)       | An example of copilot designed for code generation in Visual Studio Code.                       |
| [codetrans](./codetrans/README.md) | [Code Translation](https://github.com/opea-project/GenAIExamples/tree/main/CodeTrans/README.md)    | An example of programming language code translation.                                            |
| [docsum](./docsum/README.md)       | [Document Summarization](https://github.com/opea-project/GenAIExamples/tree/main/DocSum/README.md) | An example of document summarization.                                                           |
| [faqgen](./faqgen/README.md)       | [FAQ generator](https://github.com/opea-project/GenAIExamples/tree/main/FaqGen/README.md)          | An example to generate FAQs.                                                                    |
| [visualqna](./audioqna/README.md)  | [Visual QnA](https://github.com/opea-project/GenAIExamples/tree/main/VisualQnA/README.md)          | An example of answering open-ended questions based on an image.                                 |

### Components

Components which are building blocks for AI application.
All components Helm charts are put in the ./common directory, and the support list is growing.
Refer to [GenAIComps](https://github.com/opea-project/GenAIComps) for details of each component.

## Deploy with Helm charts

### From Source Code

These Helm charts are designed to be easy to start, which means you can deploy a workload easily without further options.
However, `HUGGINGFACEHUB_API_TOKEN` should be set in most cases for a workload to start up correctly.
Examples of deploy a workload:

```
export myrelease=mytgi
export chartname=common/tgi
helm dependency update $chartname
helm install $myrelease $chartname --set global.HUGGINGFACEHUB_API_TOKEN="insert-your-huggingface-token-here"
```

Depending on your environment, you may want to customize some of the options, see [Helm Charts Options](#helm-charts-options) for further information.

### Using Helm Charts repository

The Helm charts are released to https://github.com/orgs/opea-project/packages. You can check the list there and deploy with

```
export chartname=chatqna
helm install myrelease oci://ghcr.io/opea-project/charts/${chartname}
```

## Helm Charts Options

Here is a list of a few important options that user may want to change.

For more options, read each Helm chart's `README.md` file and check its `values.yaml` or `gaudi-values.yaml` files (if applicable).

There are global options (which should be shared across all components of a workload) and specific options that only apply to one component.

| Helm chart | Options                         | Description                                                                                                                                                                                                                                                                    |
| ---------- | ------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| global     | HUGGINGFACEHUB_API_TOKEN        | Your own HuggingFace token, there is no default value. If not set, you might fail to start the component.                                                                                                                                                                      |
| global     | http_proxy https_proxy no_proxy | Proxy settings. If you are running the workloads behind the proxy, you'll have to add your proxy settings here.                                                                                                                                                                |
| global     | modelUsePVC                     | The PersistentVolumeClaim you want to use as HuggingFace hub cache. Default "" means not using PVC. Only one of modelUsePVC/modelUseHostPath can be set.                                                                                                                       |
| global     | modelUseHostPath                | If you don't have Persistent Volume in your k8s cluster and want to use local directory as HuggingFace hub cache, set modelUseHostPath to your local directory name. Note that this can't share across nodes. Default "". Only one of modelUsePVC/modelUseHostPath can be set. |
| global     | monitoring                      | Enable monitoring for (ChatQnA) service components. See [Pre-conditions](monitoring.md#pre-conditions) before enabling!                                                                                                                                                        |
| tgi        | LLM_MODEL_ID                    | The model id you want to use for tgi server. Default "Intel/neural-chat-7b-v3-3".                                                                                                                                                                                              |

## Deploy the Helm Charts on Intel® Xeon® Processors with Intel® Trust Domain Extensions (Intel® TDX)

See [TDX instructions](TDX.md) on how to deploy the Helm Charts on Intel® Xeon® processors with Intel® Trust Domain Extensions (Intel® TDX).

## Using HPA (autoscaling)

See [HPA instructions](HPA.md) on how to enable horizontal pod autoscaling for service components, based on their usage metrics.

## Using Persistent Volume

It's common to use Persistent Volume (PV) for model caches (HuggingFace hub cache) in a production k8s cluster. PersistentVolumeClaim (PVC) can be passed to containers, but it's the user's responsibility to create the PVC depending on your k8s cluster's capability.

This example setup uses NFS on Ubuntu 22.04.

- Export NFS directory from NFS server

```
sudo apt install nfs-kernel-server
sudo mkdir -p /data/nfspv && sudo chown nobody:nogroup /data/nfspv && sudo chmod 777 /data/nfspv
echo "/data/nfspv 192.168.0.0/24(rw,sync,no_subtree_check)" |sudo tee -a /etc/exports
sudo systemctl restart nfs-server

```

- Create a Persistent Volume

```
cat <<EOF >nfspv.yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: nfspv
spec:
  capacity:
    storage: 300Gi
  volumeMode: Filesystem
  accessModes:
    - ReadWriteMany
  persistentVolumeReclaimPolicy: Retain
  storageClassName: nfs
  nfs:
    path: "/data/nfspv"
    server: "192.168.0.184"
    readOnly: false
EOF
kubectl apply -f nfspv.yaml
```

- Create a PersistentVolumeClaim

```
cat << EOF > nfspvc.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: model-volume
spec:
  accessModes:
    - ReadWriteMany
  storageClassName: "nfs"
  resources:
    requests:
      storage: 100Gi
EOF
```

- Set `global.modelUsePVC` when doing Helm install, or modify the `values.yaml`

```
helm install tgi common/tgi --set global.modelUsePVC=model-volume
```

## Using Private Docker Hub

By default, we're using Docker images from [official Docker hub](https://hub.docker.com/u/opea), with Docker image version aligned with OPEA releases.
If you have private hub, see the following examples.

To use local Docker registry:

```
export OPEA_IMAGE_REPO=192.168.0.100:5000/
find . -name '*values.yaml' -type f -exec sed -i "s#repository: opea/*#repository: ${OPEA_IMAGE_REPO}opea/#g" {} \;
```

## Generate manifests from Helm Charts

Some users may want to use Kubernetes manifests (YAML files) for workload deployment, we do not maintain manifests itself, and will generate them using `helm template`.
See `update_genaiexamples.sh` for how the manifests are generated for supported _GenAIExamples_.
See `update_manifests.sh` for how the manifests are generated for supported _GenAIComps_.
Please note that the above scripts have hardcoded settings to reduce user configuration effort.
They are not supposed to be directly used by users.
