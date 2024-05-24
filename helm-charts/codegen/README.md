# CodeGen

Helm chart for deploying CodeGen service.

CodeGen depends on LLM microservice, refer to llm-uservice for more config details.

## Installing the Chart

To install the chart, run the following:

```console
$ export HFTOKEN="insert-your-huggingface-token-here"
$ export MODELDIR="/mnt"
$ export MODELNAME="m-a-p/OpenCodeInterpreter-DS-6.7B"
$ helm install codegen codegen --set llm-uservice.HUGGINGFACE_API_TOKEN=${HFTOKEN} --set llm-uservice.tgi.volume=${MODELDIR} --set llm-uservice.tgi.LLM_MODEL_ID=${MODELNAME}
```

## Values

| Key                                   | Type   | Default                                          | Description                                                                                                                              |
| ------------------------------------- | ------ | ------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------- |
| image.repository                      | string | `"opea/gen-ai-comps:codegen-megaservice-server"` |                                                                                                                                          |
| service.port                          | string | `"6666"`                                         |                                                                                                                                          |
| llm-uservice.HUGGINGFACEHUB_API_TOKEN | string | `""`                                             | Your own Hugging Face API token                                                                                                          |
| llm-uservice.tgi.LLM_MODEL_ID         | string | `"ise-uiuc/Magicoder-S-DS-6.7B"`                 | Models id from https://huggingface.co/, or predownloaded model directory                                                                 |
| llm-uservice.tgi.volume               | string | `"/mnt"`                                         | Cached models directory, tgi will not download if the model is cached here. The "volume" will be mounted to container as /data directory |
