# reranking-usvc

Helm chart for deploying reranking microservice.

reranking-usvc depends on TEI, refer to teirerank for more config details.

## Installing the Chart

To install the chart, run the following:

```console
$ export MODELDIR="/mnt"
$ helm install reranking reranking-usvc --set global.modelUseHostPath=${MODELDIR}
```

## Values

| Key                       | Type   | Default                       | Description                                                                                                                                                  |
| ------------------------- | ------ | ----------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| image.repository          | string | `"opea/reranking-tgi:latest"` |                                                                                                                                                              |
| service.port              | string | `"8000"`                      |                                                                                                                                                              |
| teirerank.RERANK_MODEL_ID | string | `"BAAI/bge-reranker-base"`    | Models id from https://huggingface.co/, or predownloaded model directory                                                                                     |
| global.modelUseHostPath   | string | `"/mnt"`                      | Cached models directory, tgi will not download if the model is cached here. The host path "modelUseHostPath" will be mounted to container as /data directory |
