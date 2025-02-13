# data-prep

Helm chart for deploying OPEA data-prep microservice.

## Installing the chart

`data-prep` will use TEI for embedding service, and support different vector DB backends.

- TEI: please refer to the [tei](../tei) for more information.

- Redis vector DB: please refer to [redis-vector-db](../redis-vector-db/) for more information.

- Milvus DB: please refer to [milvus-helm](https://github.com/zilliztech/milvus-helm/tree/milvus-4.2.12) for more information.

- Qdrant DB: please refer to [qdrant-helm](https://github.com/qdrant/qdrant-helm/tree/qdrant-1.13.1/charts/qdrant) for more information.

First, you need to install the `tei` helm chart and one of the vector DB service, i.e. `redis-vector-db` chart.

After you've deployed dependency charts successfully, please run `kubectl get svc` to get the service endpoint URL respectively, i.e. `http://tei:80`, `redis://redis-vector-db:6379`.

To install `data-prep` chart, run the following:

```console
cd GenAIInfra/helm-charts/common/data-prep
helm dependency update
export HFTOKEN="insert-your-huggingface-token-here"
export TEI_EMBEDDING_ENDPOINT="http://tei"

# Install data-prep with Redis DB backend
export DATAPREP_BACKEND="REDIS"
export DB_HOST="redis-vector-db"
helm install data-prep . --set TEI_EMBEDDING_ENDPOINT=${TEI_EMBEDDING_ENDPOINT} --set global.HUGGINGFACEHUB_API_TOKEN=${HF_TOKEN} --set DATAPREP_BACKEND=${DATAPREP_BACKEND} --set REDIS_HOST=${DB_HOST}

# Install data-prep with Milvus DB backend
# export DATAPREP_BACKEND="MILVUS"
# export DB_HOST="milvus"
# helm install data-prep . --set TEI_EMBEDDING_ENDPOINT=${TEI_EMBEDDING_ENDPOINT} --set global.HUGGINGFACEHUB_API_TOKEN=${HF_TOKEN} --set DATAPREP_BACKEND=${DATAPREP_BACKEND} --set MILVUS_HOST=${DB_HOST},MILVUS_PORT=19530,COLLECTION_NAME=rag__milvus

# Install data-prep with Qdrant DB backend
# export DATAPREP_BACKEND="QDRANT"
# export DB_HOST="qdrant"
# helm install data-prep . --set TEI_EMBEDDING_ENDPOINT=${TEI_EMBEDDING_ENDPOINT} --set global.HUGGINGFACEHUB_API_TOKEN=${HF_TOKEN} --set DATAPREP_BACKEND=${DATAPREP_BACKEND} --set QDRANT_HOST=${DB_HOST},QDRANT_PORT=6333,COLLECTION_NAME=rag_qdrant
```

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are running.

Then run the command `kubectl port-forward svc/data-prep 6007:6007` to expose the data-prep service for access.

Open another terminal and run the following command to verify the service if working:

```console
curl http://localhost:6007/v1/dataprep/ingest  \
    -X POST \
    -H "Content-Type: multipart/form-data" \
    -F "files=@./README.md"
```

## Values

| Key                             | Type   | Default   | Description                                                                                             |
| ------------------------------- | ------ | --------- | ------------------------------------------------------------------------------------------------------- |
| service.port                    | string | `"6007"`  |                                                                                                         |
| global.HUGGINGFACEHUB_API_TOKEN | string | `""`      | Your own Hugging Face API token                                                                         |
| DATAPREP_BACKEND                | string | `"REDIS"` | vector DB backend to use, one of "REDIS", "MILVUS", "QDRANT"                                            |
| REDIS_HOST                      | string | `""`      | Redis service URL host, only valid for Redis, please see `values.yaml` for other Redis configuration    |
| MILVUS_HOST                     | string | `""`      | Milvus service URL host, only valid for Milvus, please see `values.yaml` for other Milvus configuration |
| QDRANT_HOST                     | string | `""`      | Qdrant service URL host, only valid for Qdrant, please see `values.yaml` for other Qdrant configuration |
| TEI_EMBEDDING_ENDPOINT          | string | `""`      |                                                                                                         |
| global.monitoring               | bool   | `false`   | See ../../monitoring.md before enabling!                                                                |

## Milvus support

Refer to the milvus-values.yaml for milvus configurations.
