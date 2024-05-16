# tgi

Helm chart for deploying Hugging Face Text Generation Inference service.

## Installing the Chart

To install the chart, run the following:

```console
$ export MODELDIR=/mnt
$ export MODELNAME="bigscience/bloom-560m"
$ helm install tgi tgi --set hftgi.volume=${MODELDIR} --set hftgi.modelId=${MODELNAME}
```

By default, the tgi service will downloading the "bigscience/bloom-560m" which is about 1.1GB.

If you already cached the model locally, you can pass it to container like this example:

MODELDIR=/home/ubuntu/hfmodels

MODELNAME="/data/models--bigscience--bloom-560m"

## Values

| Key           | Type   | Default                                           | Description                                                                                                                              |
| ------------- | ------ | ------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| hftgi.modelId | string | `"bigscience/bloom-560m"`                         | Models id from https://huggingface.co/, or predownloaded model directory                                                                 |
| hftgi.port    | string | `"80"`                                            | Hugging Face Text Generation Inference service port                                                                                      |
| hftgi.volume  | string | `"/mnt"`                                          | Cached models directory, tgi will not download if the model is cached here. The "volume" will be mounted to container as /data directory |
| hftgi.image   | string | `"ghcr.io/huggingface/text-generation-inference"` |                                                                                                                                          |
| hftgi.tag     | string | `"1.4"`                                           |                                                                                                                                          |
| service.port  | string | `"80"`                                            | The service port                                                                                                                         |
