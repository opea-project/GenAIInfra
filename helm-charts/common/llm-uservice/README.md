# llm-uservice

Helm chart for deploying LLM microservice.

llm-uservice depends on TGI, refer to tgi for more config details.

## Installing the Chart

To install the chart, run the following:

```console
cd GenAIInfra/helm-charts/common
helm dependency update llm-uservice
export HFTOKEN="insert-your-huggingface-token-here"
export MODELDIR="/mnt/opea-models"
export MODELNAME="m-a-p/OpenCodeInterpreter-DS-6.7B"
helm install llm llm-uservice --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} --set tgi.LLM_MODEL_ID=${MODELNAME} --wait
# To deploy on Gaudi enabled k8s cluster
# helm install llm llm-uservice --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} --set tgi.LLM_MODEL_ID=${MODELNAME} --values llm-uservice/gaudi-values.yaml --wait
```

## Values

| Key                             | Type   | Default                               | Description                                                                                                                                                  |
| ------------------------------- | ------ | ------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| global.HUGGINGFACEHUB_API_TOKEN | string | `""`                                  | Your own Hugging Face API token                                                                                                                              |
| global.modelUseHostPath         | string | `"/mnt/opea-models"`                  | Cached models directory, tgi will not download if the model is cached here. The host path "modelUseHostPath" will be mounted to container as /data directory |
| image.repository                | string | `"opea/llm-tgi:latest"`               |                                                                                                                                                              |
| service.port                    | string | `"9000"`                              |                                                                                                                                                              |
| tgi.LLM_MODEL_ID                | string | `"m-a-p/OpenCodeInterpreter-DS-6.7B"` | Models id from https://huggingface.co/, or predownloaded model directory                                                                                     |
