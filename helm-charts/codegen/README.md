# CodeGen

Helm chart for deploying CodeGen service.

CodeGen depends on LLM microservice, refer to llm-uservice for more config details.

## Installing the Chart

To install the chart, run the following:

```console
cd GenAIInfra/helm-charts/
./update_dependency.sh
helm dependency update codegen
export HFTOKEN="insert-your-huggingface-token-here"
export MODELDIR="/mnt"
export MODELNAME="m-a-p/OpenCodeInterpreter-DS-6.7B"
helm install codegen codegen --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} --set llm-uservice.tgi.LLM_MODEL_ID=${MODELNAME}
# To use Gaudi device
# helm install codegen codegen --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --values codegen/gaudi-values.yaml
```

## Values

| Key                             | Type   | Default                          | Description                                                                                                                                                  |
| ------------------------------- | ------ | -------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| image.repository                | string | `"opea/codegen:latest"`          |                                                                                                                                                              |
| service.port                    | string | `"7778"`                         |                                                                                                                                                              |
| global.HUGGINGFACEHUB_API_TOKEN | string | `""`                             | Your own Hugging Face API token                                                                                                                              |
| global.modelUseHostPath         | string | `"/mnt"`                         | Cached models directory, tgi will not download if the model is cached here. The host path "modelUseHostPath" will be mounted to container as /data directory |
| llm-uservice.tgi.LLM_MODEL_ID   | string | `"ise-uiuc/Magicoder-S-DS-6.7B"` | Models id from https://huggingface.co/, or predownloaded model directory                                                                                     |
