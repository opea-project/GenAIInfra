# Helm charts for deploying GenAI Components and Examples

This directory contains helm charts for [GenAIComps](https://github.com/opea-project/GenAIComps) and [GenAIExamples](https://github.com/opea-project/GenAIExamples) deployment on Kubernetes.

## Table of Contents

- [Helm Charts](#helm-charts)
  - [Examples](#examples)
  - [Components](#components)
- [How to deploy with helm charts](#deploy-with-helm-charts)
- [Helm Charts Options](#helm-charts-options)
- [HorizontalPodAutoscaler (HPA) support](#horizontalpodautoscaler-hpa-support)
  - [Pre-conditions](#pre-conditions)
  - [Gotchas](#gotchas)
  - [Verify HPA metrics](#verify-hpa-metrics)
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
| chatqna    | horizontalPodAutoscaler.enabled | Enable HPA autoscaling for TGI and TEI service deployments based on metrics they provide. See [Pre-conditions](#pre-conditions) and [Gotchas](#gotchas) before enabling!                                                                                                       |
| tgi        | LLM_MODEL_ID                    | The model id you want to use for tgi server. Default "Intel/neural-chat-7b-v3-3".                                                                                                                                                                                              |

## HorizontalPodAutoscaler (HPA) support

`horizontalPodAutoscaler` option enables HPA scaling for the TGI and TEI inferencing deployments:
https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/

Autoscaling is based on custom application metrics provided through [Prometheus](https://prometheus.io/).

### Pre-conditions

If cluster does not run [Prometheus operator](https://github.com/prometheus-operator/kube-prometheus)
yet, it SHOULD be be installed before enabling HPA, e.g. by using:
https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack

Enabling HPA in top-level Helm chart (e.g. `chatqna`), overwrites cluster's current _PrometheusAdapter_
configuration with relevant custom metric queries. If that has existing queries that should be retained,
relevant queries need to be added to existing _PrometheusAdapter_ configuration _manually_ from the
custom metrics Helm template (in top-level Helm chart).

### Gotchas

Why HPA is opt-in:

- Enabling (top level) chart `horizontalPodAutoscaler` option will _overwrite_ cluster's current
  `PrometheusAdapter` configuration with its own custom metrics configuration.
  Take copy of the existing one before install, if that matters:
  `kubectl -n monitoring get cm/adapter-config -o yaml > adapter-config.yaml`
- `PrometheusAdapter` needs to be restarted after install, for it to read the new configuration:
  `ns=monitoring; kubectl -n $ns delete $(kubectl -n $ns get pod --selector app.kubernetes.io/name=prometheus-adapter -o name)`
- By default Prometheus adds [k8s RBAC rules](https://github.com/prometheus-operator/kube-prometheus/blob/main/manifests/prometheus-roleBindingSpecificNamespaces.yaml)
  for accessing metrics from `default`, `kube-system` and `monitoring` namespaces. If Helm is
  asked to install OPEA services to some other namespace, those rules need to be updated accordingly
- Current HPA rules are examples for Xeon, for efficient scaling they need to be fine-tuned for given setup
  performance (underlying HW, used models and data types, OPEA version etc)

### Verify HPA metrics

To verify that metrics required by horizontalPodAutoscaler option work, check following...

Prometheus has found the metric endpoints, i.e. last number on `curl` output is non-zero:

```console
chart=chatqna; # OPEA services prefix
ns=monitoring; # Prometheus namespace
prom_url=http://$(kubectl -n $ns get -o jsonpath="{.spec.clusterIP}:{.spec.ports[0].port}" svc/prometheus-k8s);
curl --no-progress-meter $prom_url/metrics | grep scrape_pool_targets.*$chart
```

**NOTE**: TGI and TEI inferencing services provide metrics endpoint only after they've processed their first request!

PrometheusAdapter lists TGI and/or TGI custom metrics (`te_*` / `tgi_*`):

```console
kubectl get --raw /apis/custom.metrics.k8s.io/v1beta1 | jq .resources[].name
```

HPA rules list valid (not `<unknown>`) TARGET values for service deployments:

```console
ns=default;  # OPEA namespace
kubectl -n $ns get hpa
```

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
