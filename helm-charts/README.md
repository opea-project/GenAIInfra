# Helm charts for deploying GenAI Components and Examples

This directory contains helm charts for [GenAIComps](https://github.com/opea-project/GenAIComps) and [GenAIExamples](https://github.com/opea-project/GenAIExamples) deployment on Kubernetes.

## Table of Contents

- [Helm Charts](#helm-charts)
  - [Examples](#examples)
  - [Components](#components)
- [How to deploy with helm charts](#deploy-with-helm-charts)
- [Helm Charts Options](#helm-charts-options)
- [Using Persistent Volume](#using-persistent-volume)
- [Using Private Docker Hub](#using-private-docker-hub)
- [Helm Charts repository](#helm-chart-repository)
- [Generate manifests from Helm Charts](#generate-manifests-from-helm-charts)

## Helm Charts

List of supported workloads and components.

### Examples

AI application examples you can run directly on Xeon and Gaudi. You can also refer to these examples to develop your own customized AI application.

| Helm chart               | Link to GenAIExamples                                                                    | Description                                                                                     |
| ------------------------ | ---------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------- |
| [codegen](./codegen)     | [Code Generation](https://github.com/opea-project/GenAIExamples/tree/main/CodeGen)       | An example of copilot designed for code generation in Visual Studio Code.                       |
| [codetrans](./codetrans) | [Code Translation](https://github.com/opea-project/GenAIExamples/tree/main/CodeTrans)    | An example of programming language code translation.                                            |
| [chatqna](./chatqna)     | [ChatQnA](https://github.com/opea-project/GenAIExamples/tree/main/ChatQnA)               | An example of chatbot for question and answering through retrieval argumented generation (RAG). |
| [docsum](./docsum)       | [Document Summarization](https://github.com/opea-project/GenAIExamples/tree/main/DocSum) | An example of document summarization.                                                           |

### Components

Components which are building blocks for AI application.  
All components helm charts are put in the ./common directory, and the support list is growing.  
Refer to [GenAIComps](https://github.com/opea-project/GenAIComps) for details of each component.

## How to deploy with helm charts

These helm charts are designed to be easy to start, which means you can deploy a workload easily without further options.  
However, `HUGGINGFACEHUB_API_TOKEN` should be set in most cases for a workload to start up correctly.  
Examples of deploy a workload:

```
export myrelease=mytgi
export chartname=common/tgi
helm dependency update $chartname
helm install $myrelease $chartname --set global.HUGGINGFACEHUB_API_TOKEN="insert-your-huggingface-token-here"
```

Depends on your environment, you might want to customize some of the options, see [Helm Charts Options](#helm-charts-options) for further information.

## Helm Charts Options

Here we list a few important options that user might want to change, for more options, you can read each helm chart's README file and check the values.yaml and gaudi-values.yaml(If applicable).

There are global options(which should be shared across all components of a workload) and specific options that only apply to one component.

| Helm chart | Options                         | Description                                                                                                                                                                                                                                                                    |
| ---------- | ------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| global     | HUGGINGFACEHUB_API_TOKEN        | Your own huggingface token, there is no default value. If not set, you might fail to start the component.                                                                                                                                                                      |
| global     | http_proxy https_proxy no_proxy | Proxy settings. If you are running the workloads behind the proxy, you'll have to add your proxy settings here.                                                                                                                                                                |
| global     | modelUsePVC                     | The PersistentVolumeClaim you want to use as huggingface hub cache. Default "" means not using PVC. Only one of modelUsePVC/modelUseHostPath can be set.                                                                                                                       |
| global     | modelUseHostPath                | If you don't have Persistent Volume in your k8s cluster and want to use local directory as huggingface hub cache, set modelUseHostPath to your local directory name. Note that this can't share across nodes. Default "". Only one of modelUsePVC/modelUseHostPath can be set. |
| tgi        | LLM_MODEL_ID                    | The model id you want to use for tgi server. Default "Intel/neural-chat-7b-v3-3".                                                                                                                                                                                              |

## Using Persistent Volume

It's common to use Persistent Volume(PV) for model caches(huggingface hub cache) in a production k8s cluster. We support to pass the PersistentVolumeClaim(PVC) to containers, but it's the user's responsibility to create the PVC depending on your k8s cluster's capability.  
Here is an setup example using NFS on Ubuntu 22.04.

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

- Set global.modelUsePVC when doing helm install, or modify the values.yaml

```
helm install tgi common/tgi --set global.modelUsePVC=model-volume
```

## Using Private Docker Hub

By default, we're using docker images from [official docker hub](https://hub.docker.com/u/opea), with docker image version aligned with OPEA releases.  
If you have private hub or would like to use different docker image versions, see the following examples.

To use the latest tag for all images:  
`find . -name '*values.yaml' -type f -exec sed -i 's#tag: ""#tag: latest#g' {} \;`

To use local docker registry:

```
export OPEA_IMAGE_REPO=192.168.0.100:5000/
find . -name '*values.yaml' -type f -exec sed -i "s#repository: opea/*#repository: ${OPEA_IMAGE_REPO}opea/#g" {} \;
```

## Helm Charts repository (Experimental)

https://opea-project.github.io/GenAIInfra

## Generate manifests from Helm Charts

Some users may want to use kubernetes manifests(yaml files) for workload deployment, we do not maintain manifests itself, and will generate them using `helm template`.  
See update_genaiexamples.sh for how the manifests are generated for supported GenAIExamples.  
See update_manifests.sh for how the manifests are generated for supported GenAIComps.  
Please note that the above scripts have hardcoded settings to reduce user configuration effort.  
They are not supposed to be directly used by users.
