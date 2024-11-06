# Data-Prep Microservice

Helm chart for deploying data-prep microservice. Data-Prep is consumed by several reference applications present in [GenAIExample](https://github.com/opea-project/GenAIExamples/tree/main).

There are 2 versions of Data-Prep microservice. First version is unimodal based on redis-vector-db and TEI. It performs data preparation for textual data. An alternative multimodal version based on `vdms-values.yaml` file, performs data preparation for visual data input. Follow along to select and install the version which suites your use case.

Data-Prep uses redis-vector-db and tei. The multimodal version uses vdms-vector-db service. Endpoints for these dependencies should be set properly before installing the chart.

## Install the chart for data preparation using Redis Vector DB

### (Option1): Installing the chart separately

First, you need to install the tei and redis-vector-db chart, please refer to the [tei](../tei/README.md) and [redis-vector-db](../redis-vector-db/README.md) for more information.

After you've deployed the tei and redis-vector-db chart successfully, please run `kubectl get svc` to get the service endpoint and URL respectively, i.e. `http://tei`, `redis://redis-vector-db:6379`.

To install data-prep chart, run the following:

```console
cd GenAIInfra/helm-charts/common/data-prep
export REDIS_URL="redis://redis-vector-db:6379"
export TEI_EMBEDDING_ENDPOINT="http://tei"
helm dependency update
helm install data-prep . --set REDIS_URL=${REDIS_URL} --set TEI_EMBEDDING_ENDPOINT=${TEI_EMBEDDING_ENDPOINT}
```

### (Option2): Installing the chart with dependencies automatically

```console
cd GenAIInfra/helm-charts/common/data-prep
helm dependency update
helm install data-prep . --set redis-vector-db.enabled=true --set tei.enabled=true

```

## Install the chart for multimodal data preparation using VDMS Vector DB

### (Option1): Installing the chart separately

First, you need to install the `vdms-vector-db` chart. Please refer to the [vdms-vector-db](../vdms-vector-db/README.md) for more information.

After you've deployed the `vdms-vector-db` chart successfully, please run `kubectl get svc` to get the service host and port respectively, for example: `http://vdms-vector-db:8001`.

Next, Run the following commands to install data-prep chart:

```bash
cd GenAIInfra/helm-charts/common/data-prep

# Use the host and port received in previous step as VDMS_HOST and VDMS_PORT.
export VDMS_HOST="vdms-vector-db"
export VDMS_PORT="8001"
export INDEX_NAME="mega-videoqna"
export HFTOKEN=<your huggingface token>
# Set a directory to cache emdedding models
export MODELDIR=/mnt/opea-models

# Export the proxy variables. Assign empty string if no proxy setup required.
export https_proxy="your_http_proxy"
export http_proxy="your_https_proxy"

helm dependency update
helm install data-prep . -f ../variant_videoqna-values.yaml --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set indexName=${INDEX_NAME} --set global.modelUseHostPath=${MODELDIR} --set vdmsHost=${VDMS_HOST} --set vdmsPort=${VDMS_PORT} --set global.https_proxy=${https_proxy} --set global.http_proxy=${http_proxy}
```

### (Option2): Installing the chart with dependencies automatically

```bash
cd GenAIInfra/helm-charts/common/data-prep
export INDEX_NAME="mega-videoqna"
export HFTOKEN=<your huggingface token>
# Set a directory to cache emdedding models
export MODELDIR=/mnt/opea-models

# Export the proxy variables. Assign empty string if no proxy setup required.
export https_proxy="your_http_proxy"
export http_proxy="your_https_proxy"

helm dependency update
helm install data-prep . -f ./variant_videoqna-values.yaml --set vdms-vector-db.enabled=true --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set indexName=${INDEX_NAME} --set global.modelUseHostPath=${MODELDIR} --set global.https_proxy=${https_proxy} --set global.http_proxy=${http_proxy}
```

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are running.

Then run the command `kubectl port-forward svc/data-prep 6007:6007` to expose the data-prep service for access.

Open another terminal and run the following command to verify the service if working:

### 1. For Data-prep service using redis-vector-db:

```bash

curl http://localhost:6007/v1/dataprep  \
    -X POST \
    -H "Content-Type: multipart/form-data" \
    -F "files=@./README.md"
```

### 2. For multimodal data prep service using vdms-vector-db:

```bash
# 1) Download a sample video in current dir:
curl -svLO "https://github.com/opea-project/GenAIExamples/raw/refs/heads/main/VideoQnA/docker_compose/intel/cpu/xeon/data/op_1_0320241830.mp4"

# 2) Verify using above video
curl -X POST http://localhost:6007/v1/dataprep \
      -H "Content-Type: multipart/form-data" \
      -F "files=@./op_1_0320241830.mp4"
```

## Values

| Key                          | Type   | Default                           | Description |
| ---------------------------- | ------ | --------------------------------- | ----------- |
| image.repository             | string | `"opea/dataprep-redis"`           |             |
| service.port                 | string | `"6007"`                          |             |
| REDIS_URL                    | string | `""`                              |             |
| TEI_EMBEDDING_ENDPOINT       | string | `""`                              |             |
| vdms-values:image.repository | string | `"opea/dataprep-multimodal-vdms"` |             |
| vdms-values:vdmsHost         | string | `""`                              |             |
| vdms-values:vdmsPort         | string | `"8001"`                          |             |
| vdms-values:indexName        | string | `"mega-videoqna"`                 |             |

## Milvus support

Refer to the milvus-values.yaml for milvus configurations.
