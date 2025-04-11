# retriever-usvc

Helm chart for deploying OPEA retriever-usvc microservice.

## Installing the chart

`retriever-usvc` will use TEI for embedding service, and support different vector DB backends.

- TEI: please refer to the [tei](../tei) for more information.

- Redis vector DB: please refer to [redis-vector-db](../redis-vector-db/) for more information.

- Milvus DB: please refer to [milvus-helm](https://github.com/zilliztech/milvus-helm/tree/milvus-4.2.12) for more information.

- Qdrant DB: please refer to [qdrant-helm](https://github.com/qdrant/qdrant-helm/tree/qdrant-1.13.1/charts/qdrant) for more information.

First, you need to install the `tei` helm chart and one of the vector DB service, i.e. `redis-vector-db` chart.

After you've deployed dependency charts successfully, please run `kubectl get svc` to get the service endpoint URL respectively, i.e. `http://tei:80`, `redis://redis-vector-db:6379`.

To install `retriever-usvc` chart, run the following:

```console
cd GenAIInfra/helm-charts/common/retriever-usvc
helm dependency update
export HFTOKEN="insert-your-huggingface-token-here"
export TEI_EMBEDDING_ENDPOINT="http://tei"

# Install retriever-usvc with Redis DB backend
export RETRIEVER_BACKEND="REDIS"
export DB_HOST="redis-vector-db"
helm install retriever-usvc . --set TEI_EMBEDDING_ENDPOINT=${TEI_EMBEDDING_ENDPOINT} --set global.HUGGINGFACEHUB_API_TOKEN=${HF_TOKEN} --set RETRIEVER_BACKEND=${RETRIEVER_BACKEND} --set REDIS_HOST=${DB_HOST}

# Install retriever-usvc with Milvus DB backend
# export RETRIEVER_BACKEND="MILVUS"
# export DB_HOST="milvus"
# helm install retriever-usvc . --set TEI_EMBEDDING_ENDPOINT=${TEI_EMBEDDING_ENDPOINT} --set global.HUGGINGFACEHUB_API_TOKEN=${HF_TOKEN} --set RETRIEVER_BACKEND=${RETRIEVER_BACKEND} --set MILVUS_HOST=${DB_HOST}

# Install retriever-usvc with Qdrant DB backend
# export RETRIEVER_BACKEND="QDRANT"
# export DB_HOST="qdrant"
# helm install retriever-usvc . --set TEI_EMBEDDING_ENDPOINT=${TEI_EMBEDDING_ENDPOINT} --set global.HUGGINGFACEHUB_API_TOKEN=${HF_TOKEN} --set RETRIEVER_BACKEND=${RETRIEVER_BACKEND} --set QDRANT_HOST=${DB_HOST}
```

### Install the microservice in airgap offline mode

To support running this microservice in airgap offline environment, users are required to manually download the offline data including the nltk data in a shared storage, below is the example to use node level local directory to download the offline data:

Suppose we share the model data on node-level using local directory `/mnt/opea-models`

```
# On every K8s node, run the following command:
export MODELDIR=/mnt/opea-models
# Download nltk data, suppose you have installed python nltk modules
python -m nltk.downloader -d ${MODELDIR}/nltk_data all && chmod -R a+r ${MODELDIR}/nltk_data

# Install using helm with the following additional parameters:
# helm install ... ... --set global.offline=true,global.modelUseHostPath=${MODELDIR}
```

Suppose we share the model data on cluster level using a persistent volume(PV), first we need to create a persistent volume claim(PVC) with name `opea-model-pvc`:

```
# Download nltk data and model into the `nltk_data` directory at the root of the corresponding PV
# Install using helm with the following additional parameters:
# export PVC=opea-model-pvc
# helm install ... ... --set global.offline=true,global.modelUsePVC=${PVC}
```

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are running.

Then run the command `kubectl port-forward svc/retriever-usvc 7000:7000` to expose the retriever-usvc service for access.

Open another terminal and run the following command to verify the service if working:

```console
export your_embedding=$(python3 -c "import random; embedding = [random.uniform(-1, 1) for _ in range(768)]; print(embedding)")
curl http://localhost:7000/v1/retrieval  \
    -X POST \
    -d "{\"text\":\"What is the revenue of Nike in 2023?\",\"embedding\":${your_embedding}}" \
    -H 'Content-Type: application/json'
```

## Values

| Key                             | Type   | Default   | Description                                                                                             |
| ------------------------------- | ------ | --------- | ------------------------------------------------------------------------------------------------------- |
| global.HUGGINGFACEHUB_API_TOKEN | string | `""`      | Your own Hugging Face API token                                                                         |
| service.port                    | string | `"7000"`  |                                                                                                         |
| RETRIEVER_BACKEND               | string | `"REDIS"` | vector DB backend to use, one of "REDIS", "MILVUS", "QDRANT"                                            |
| REDIS_HOST                      | string | `""`      | Redis service URL host, only valid for Redis, please see `values.yaml` for other Redis configuration    |
| MILVUS_HOST                     | string | `""`      | Milvus service URL host, only valid for Milvus, please see `values.yaml` for other Milvus configuration |
| QDRANT_HOST                     | string | `""`      | Qdrant service URL host, only valid for Qdrant, please see `values.yaml` for other Qdrant configuration |
| TEI_EMBEDDING_ENDPOINT          | string | `""`      |                                                                                                         |
| global.monitoring               | bool   | `false`   |                                                                                                         |

## Milvus support

Refer to the milvus-values.yaml for Milvus configurations.
