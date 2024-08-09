# CodeGen

Helm chart for deploying CodeGen service.

CodeGen depends on LLM and tgi microservice, refer to [llm-uservice](../common/llm-uservice) and [tgi](../common/tgi) for more config details.

## Installing the Chart

To install the chart, run the following:

```console
cd GenAIInfra/helm-charts/
./update_dependency.sh
helm dependency update codegen
export HFTOKEN="insert-your-huggingface-token-here"
export MODELDIR="/mnt/opea-models"
export MODELNAME="meta-llama/CodeLlama-7b-hf"
# To run on Xeon
helm install codegen codegen --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} --set tgi.LLM_MODEL_ID=${MODELNAME}
# To run on Gaudi
#helm install codegen codegen --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} --set tgi.LLM_MODEL_ID=${MODELNAME} -f codegen/gaudi-values.yaml
```

### IMPORTANT NOTE

1. To use model `meta-llama/CodeLlama-7b-hf`, you should first goto the [huggingface model card](https://huggingface.co/meta-llama/CodeLlama-7b-hf) to apply for the model access first. You need to make sure your huggingface token has at least read access to that model.

2. Make sure your `MODELDIR` exists on the node where your workload is schedueled so you can cache the downloaded model for next time use. Otherwise, set `global.modelUseHostPath` to 'null' if you don't want to cache the model.

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are running.

Then run the command `kubectl port-forward svc/codegen 7778:7778` to expose the service for access.

Open another terminal and run the following command to verify the service if working:

```console
curl http://localhost:7778/v1/codegen \
    -H "Content-Type: application/json" \
    -d '{"messages": "Implement a high-level API for a TODO list application. The API takes as input an operation request and updates the TODO list in place. If the request is invalid, raise an exception."}'
```

## Values

| Key                             | Type   | Default                        | Description                                                                                                                                                                                                                       |
| ------------------------------- | ------ | ------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| image.repository                | string | `"opea/codegen"`               |                                                                                                                                                                                                                                   |
| service.port                    | string | `"7778"`                       |                                                                                                                                                                                                                                   |
| global.HUGGINGFACEHUB_API_TOKEN | string | `""`                           | Your own Hugging Face API token                                                                                                                                                                                                   |
| global.modelUseHostPath         | string | `"/mnt/opea-models"`           | Cached models directory, tgi will not download if the model is cached here. The host path "modelUseHostPath" will be mounted to container as /data directory. Set this to empty/null will force it to download models every time. |
| tgi.LLM_MODEL_ID                | string | `"meta-llama/CodeLlama-7b-hf"` | Models id from https://huggingface.co/, or predownloaded model directory                                                                                                                                                          |
