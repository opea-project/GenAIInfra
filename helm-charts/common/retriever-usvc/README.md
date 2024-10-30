# Retriever Microservice (retriever-usvc)

**Helm chart for deploying Retriever microservice.**

There are two versions of Retriever microservice. First one is based on redis-vector-db and TEI. It does retrieval for textual data. An alternative multimodal version based on `vdms-values.yaml` file, performs retrieval for visual data. Follow along to select and install the version which suites your use case.

retriever-usvc depends on redis-vector-db, tei. The multimodal version depends on vdms-vector-db. Endpoints for these dependencies should be set properly before installing the chart.

## Install Retriever Microservice based on redis-vector-db and TEI

### (Option1): Installing the chart separately

First, you need to install the tei and redis-vector-db chart, refer to the [tei](../tei/README.md) and [redis-vector-db](../redis-vector-db/README.md) for more information.

After you've deployed the tei and redis-vector-db chart successfully, run `kubectl get svc` to get the service endpoint and URL respectively, i.e. `http://tei`, `redis://redis-vector-db:6379`.

To install retriever-usvc chart, run the following:

```console
cd GenAIInfra/helm-charts/common/retriever-usvc
export REDIS_URL="redis://redis-vector-db:6379"
export TEI_EMBEDDING_ENDPOINT="http://tei"
helm dependency update
helm install retriever-usvc . --set REDIS_URL=${REDIS_URL} --set TEI_EMBEDDING_ENDPOINT=${TEI_EMBEDDING_ENDPOINT}
```

### (Option2): Installing the chart with dependencies automatically

```console
cd GenAIInfra/helm-charts/common/retriever-usvc
helm dependency update
helm install retriever-usvc . --set tei.enabled=true --set redis-vector-db.enabled=true
```

## Install Retriever microservice based on vdms-vector-db

### (Option1): Installing the chart separately

First, you need to install the `vdms-vector-db`. Refer to [vdms-vector-db](../vdms-vector-db) chart guide for more information.

After you've deployed `vdms-vector-db` chart successfully, run `kubectl get svc` to get the service endpoint and port for vdms-vector-db service.

To install retriever-usvc chart, run the following:

```bash
cd GenAIInfra/helm-charts/common/retriever-usvc

# Use the host and port received in previous step as VDMS_HOST and VDMS_PORT.
export VDMS_HOST="vdms-vector-db"
export VDMS_PORT="8001"
export INDEX_NAME="mega-videoqna"
export HFTOKEN=<your huggingface token>
# Set a directory to cache emdedding models
export CACHEDIR="/home/$USER/.cache"

# Export the proxy variables. Assign empty string if no proxy setup required.
export https_proxy="your_http_proxy"
export http_proxy="your_https_proxy"

helm dependency update
helm install retriever-usvc . -f ./variant_videoqna-values.yaml --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set indexName=${INDEX_NAME} --set global.cacheUseHostPath=${CACHEDIR} --set global.https_proxy=${https_proxy} --set global.http_proxy=${http_proxy}

```

### (Option2): Installing the chart with dependencies automatically

```bash
cd GenAIInfra/helm-charts/common/retriever-usvc
export INDEX_NAME="mega-videoqna"
export HFTOKEN=<your huggingface token>
# Set a directory to cache emdedding models
export CACHEDIR="/home/$USER/.cache"

# Export the proxy variables. Assign empty string if no proxy setup required.
export https_proxy="your_http_proxy"
export http_proxy="your_https_proxy"

helm dependency update
helm install retriever-usvc . -f ./variant_videoqna-values.yaml --set vdms-vector-db.enabled=true --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set indexName=${INDEX_NAME} --set global.cacheUseHostPath=${CACHEDIR} --set global.https_proxy=${https_proxy} --set global.http_proxy=${http_proxy}
```

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are running.

Then run the command `kubectl port-forward svc/retriever-usvc 7000:7000` to expose the retriever-usvc service for access.

Open another terminal and run the following command to verify the service if working:

```bash
export your_embedding=$(python3 -c "import random; embedding = [random.uniform(-1, 1) for _ in range(768)]; print(embedding)")
curl http://localhost:7000/v1/retrieval  \
    -X POST \
    -d "{\"text\":\"What is the revenue of Nike in 2023?\",\"embedding\":${your_embedding}}" \
    -H 'Content-Type: application/json'
```

## Values

| Key                          | Type   | Default                 | Description |
| ---------------------------- | ------ | ----------------------- | ----------- |
| image.repository             | string | `"opea/retriever-tgi"`  |             |
| service.port                 | string | `"7000"`                |             |
| REDIS_URL                    | string | `""`                    |             |
| TEI_EMBEDDING_ENDPOINT       | string | `""`                    |             |
| global.monitoring            | bop;   | `false`                 |             |
| vdms-values:image.repository | string | `"opea/retriever-vdms"` |             |

## Milvus support

Refer to the milvus-values.yaml for milvus configurations.
