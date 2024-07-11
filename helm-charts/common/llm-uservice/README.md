# llm-uservice

Helm chart for deploying LLM microservice.

llm-uservice depends on TGI, you should set TGI_LLM_ENDPOINT as tgi endpoint.

## Installing the Chart

To install the chart, run the following:

```console
cd GenAIInfra/helm-charts/common
export HFTOKEN="insert-your-huggingface-token-here"
export TGI_LLM_ENDPOINT="http://tgi"
helm install llm llm-uservice --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set TGI_LLM_ENDPOINT=${TGI_LLM_ENDPOINT} --wait
```

## Values

| Key                             | Type   | Default                 | Description                                                                                                                                                  |
| ------------------------------- | ------ | ----------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| global.HUGGINGFACEHUB_API_TOKEN | string | `""`                    | Your own Hugging Face API token                                                                                                                              |
| global.modelUseHostPath         | string | `"/mnt/opea-models"`    | Cached models directory, tgi will not download if the model is cached here. The host path "modelUseHostPath" will be mounted to container as /data directory |
| image.repository                | string | `"opea/llm-tgi:latest"` |                                                                                                                                                              |
| service.port                    | string | `"9000"`                |                                                                                                                                                              |
| TGI_LLM_ENDPOINT                | string | `""`                    | LLM endpoint                                                                                                                                                 |
