# embedding-usvc

Helm chart for deploying embedding microservice.

embedding-usvc depends on TEI, refer to tei for more config details.

## Installing the Chart

To install the chart, run the following:

```console
$ export MODELDIR="/mnt"
$ export MODELNAME="BAAI/bge-base-en-v1.5"
$ helm install embedding embedding-usvc --set global.modelUseHostPath=${MODELDIR} --set tei.EMBEDDING_MODEL_ID=${MODELNAME}
```

## Values

| Key                     | Type   | Default                       | Description                                                                                                                                                  |
| ----------------------- | ------ | ----------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| image.repository        | string | `"opea/embedding-tei:latest"` |                                                                                                                                                              |
| service.port            | string | `"6000"`                      |                                                                                                                                                              |
| tei.EMBEDDING_MODEL_ID  | string | `"BAAI/bge-base-en-v1.5"`     | Models id from https://huggingface.co/, or predownloaded model directory                                                                                     |
| global.modelUseHostPath | string | `"/mnt"`                      | Cached models directory, tgi will not download if the model is cached here. The host path "modelUseHostPath" will be mounted to container as /data directory |
