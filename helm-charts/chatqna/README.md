# ChatQnA

Helm chart for deploying ChatQnA service.

## Installing the Chart

To install the chart, run the following:

```console
export HFTOKEN="insert-your-huggingface-token-here"
export MODELDIR="/mnt"
export MODELNAME="Intel/neural-chat-7b-v3-3"
helm install chatqna chatqna --set llm-uservice.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set llm-uservice.tgi.volume=${MODELDIR} --set llm-uservice.tgi.LLM_MODEL_ID=${MODELNAME}
# To use Gaudi device
# helm install chatqna chatqna --set llm-uservice.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --values chatqna/gaudi-values.yaml
```

## Values

| Key                                   | Type   | Default                       | Description                                                                                                                              |
| ------------------------------------- | ------ | ----------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| image.repository                      | string | `"opea/chatqna:latest"`       |                                                                                                                                          |
| service.port                          | string | `"8888"`                      |                                                                                                                                          |
| llm-uservice.HUGGINGFACEHUB_API_TOKEN | string | `""`                          | Your own Hugging Face API token                                                                                                          |
| llm-uservice.tgi.LLM_MODEL_ID         | string | `"Intel/neural-chat-7b-v3-3"` | Models id from https://huggingface.co/, or predownloaded model directory                                                                 |
| llm-uservice.tgi.volume               | string | `"/mnt"`                      | Cached models directory, tgi will not download if the model is cached here. The "volume" will be mounted to container as /data directory |
