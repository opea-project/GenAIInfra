# tei

Helm chart for deploying Hugging Face Text Generation Inference service.

## Installing the Chart

To install the chart, run the following:

```console
$ cd ${GenAIInfro_repo}/helm-charts/common
$ export MODELDIR=/mnt/model
$ export MODELNAME="BAAI/bge-base-en-v1.5"
$ helm install tei tei --set hftei.volume=${MODELDIR} --set global.qnaConfig.EMBEDDING_MODEL_ID=${MODELNAME}
```

By default, the tei service will downloading the "BAAI/bge-base-en-v1.5" which is about 1.1GB.

If you already cached the model locally, you can pass it to container like this example:

MODELDIR=/mnt/model

MODELNAME="/data/BAAI/bge-base-en-v1.5"

## Values

| Key           | Type   | Default                                           | Description                                                                                                                              |
| ------------- | ------ | ------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| global.qnaConfig.EMBEDDING_MODEL_ID | string | `"BAAI/bge-base-en-v1.5"`                         | Models id from https://huggingface.co/, or predownloaded model directory                                                                 |
| hftei.port    | string | `"80"`                                            | Hugging Face Text Generation Inference service port                                                                                      |
| hftei.volume  | string | `"/mnt/model"`                                          | Cached models directory, tei will not download if the model is cached here. The "volume" will be mounted to container as /data directory |
| hftei.image   | string | `"ghcr.io/huggingface/text-embeddings-inference"` |                                                                                                                                          |
| hftei.tag     | string | `"cpu-1.2"`                                           |                                                                                                                                          |
| service.port  | string | `"80"`                                            | The service port                                                                                                                         |
