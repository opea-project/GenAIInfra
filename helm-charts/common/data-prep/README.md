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

### Install the microservice in air gapped(offline) mode

To support running this microservice in an air gapped environment, users are required to download the offline data including the `nltk` data and model `unstructuredio/yolo_x_layout` to a shared storage. Below is an example for using node level local directory to download the offline data:

Assuming the `nltk` data is shared using node-local directory `/mnt/nltk_data`, and the model data is shared using node-local directory `/mnt/opea-models`.

```
# On every K8s node, run the following command:
export MODELDIR=/mnt/opea-models
export NLTKDATA=/mnt/nltk_data
# Download nltk data, assumes Python nltk module(s) are already installed
python -m nltk.downloader -d $NLTKDATA all && chmod -R a+r $NLTKDATA
# Download model,  assumes Python huggingface_hub[cli] module are already installed
huggingface-cli download unstructuredio/yolo_x_layout --local-dir ${MODELDIR}/unstructuredio/yolo_x_layout && chmod -R a+r ${MODELDIR}/unstructuredio/yolo_x_layout

# Install using Helm with the following additional parameters:
# helm install ... ... --set global.offline=true,global.modelUseHostPath=${MODELDIR},global.nltkDataUseHostPath=${NLTKDATA}
```

Assuming we share the offline data on cluster level using a persistent volume(PV), first we need to create the persistent volume claim(PVC) with name `opea-model-pvc` to store model data, the PVC with name `opea-nltk-pvc` to store nltk data:

```
# Download nltk data and model into the root and `unstructuredio/yolo_x_layout` directory at the root of the corresponding PVs respectively
# Install using Helm with the following additional parameters:
# export MODELPVC=opea-model-pvc
# export NLTKPVC=opea-nltk-pvc
# helm install ... ... --set global.offline=true,global.modelUsePVC=${MOELPVC},global.nltkDataUsePVC=${NLTKPVC}
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
