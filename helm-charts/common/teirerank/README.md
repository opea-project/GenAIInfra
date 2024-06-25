# teirerank

Helm chart for deploying Hugging Face Text Generation Inference service.

## Installing the Chart

To install the chart, run the following:

```console
$ cd ${GenAIInfro_repo}/helm-charts/common
$ export MODELDIR=/mnt/model
$ export MODELNAME="BAAI/bge-reranker-base"
$ helm install teirerank teirerank --set global.modelUseHostPath=${MODELDIR} --set RERANK_MODEL_ID=${MODELNAME}
```

By default, the teirerank service will downloading the "BAAI/bge-base-en-v1.5" which is about 1.1GB.

If you already cached the model locally, you can pass it to container like this example:

MODELDIR=/mnt/model

MODELNAME="/data/BAAI/bge-base-en-v1.5"

## Values

| Key                     | Type   | Default                                           | Description                                                                                                                                                        |
| ----------------------- | ------ | ------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| RERANK_MODEL_ID         | string | `"BAAI/bge-reranker-base"`                        | Models id from https://huggingface.co/, or predownloaded model directory                                                                                           |
| global.modelUseHostPath | string | `"/mnt"`                                          | Cached models directory, teirerank will not download if the model is cached here. The host path "modelUseHostPath" will be mounted to container as /data directory |
| image.repository        | string | `"ghcr.io/huggingface/text-embeddings-inference"` |                                                                                                                                                                    |
| image.tag               | string | `"cpu-1.2"`                                       |                                                                                                                                                                    |
