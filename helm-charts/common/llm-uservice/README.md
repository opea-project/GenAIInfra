# llm-uservice

Helm chart for deploying llm microservice.

llm-uservice depends on tgi, refer to tgi for more config details.

## Installing the Chart

To install the chart, run the following:

```console
$ export HFTOKEN="insert-your-huggingface-token-here"
$ export MODELDIR="/mnt"
$ export MODELNAME="m-a-p/OpenCodeInterpreter-DS-6.7B"
$ helm install codegen codegen --set hfToken=${HFTOKEN} --set tgi.hftgi.volume=${MODELDIR} --set tgi.hftgi.modelId=${MODELNAME}
```

## Values

| Key               | Type   | Default                               | Description                                                                                                                              |
| ----------------- | ------ | ------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| hfToken           | string | `""`                                  | Your own Hugging Face API token                                                                                                          |
| image.repository  | string | `"intel/gen-ai-examples"`             |                                                                                                                                          |
| image.tag         | string | `"copilot"`                           |                                                                                                                                          |
| service.port      | string | `"9000"`                                |                                                                                                                                          |
| tgi.hftgi.modelId | string | `"m-a-p/OpenCodeInterpreter-DS-6.7B"` | Models id from https://huggingface.co/, or predownloaded model directory                                                                 |
| tgi.hftgi.port    | string | `"80"`                                | Hugging Face Text Generation Inference service port                                                                                      |
| tgi.hftgi.volume  | string | `"/mnt"`                              | Cached models directory, tgi will not download if the model is cached here. The "volume" will be mounted to container as /data directory |
