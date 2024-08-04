# ChatQnA

Helm chart for deploying ChatQnA service. ChatQnA depends on the following services:

- [data-prep](../common/data-prep)
- [embedding-usvc](../common/embedding-usvc)
- [tei](../common/tei)
- [retriever-usvc](../common/retriever-usvc)
- [redis-vector-db](../common/redis-vector-db)
- [reranking-usvc](../common/reranking-usvc)
- [teirerank](../common/teirerank)
- [llm-uservice](../common/llm-uservice)
- [tgi](../common/tgi)

## Installing the Chart

To install the chart, run the following:

```console
cd GenAIInfra/helm-charts/
./update_dependency.sh
helm dependency update chatqna
export HFTOKEN="insert-your-huggingface-token-here"
export MODELDIR="/mnt/opea-models"
export MODELNAME="Intel/neural-chat-7b-v3-3"
helm install chatqna chatqna --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} --set tgi.LLM_MODEL_ID=${MODELNAME}
# To use Gaudi device
#helm install chatqna chatqna --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} --set tgi.LLM_MODEL_ID=${MODELNAME} -f chatqna/gaudi-values.yaml
# To use Nvidia GPU
#helm install chatqna chatqna --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} --set tgi.LLM_MODEL_ID=${MODELNAME} -f chatqna/nv-values.yaml
```

### IMPORTANT NOTE

1. Make sure your `MODELDIR` exists on the node where your workload is schedueled so you can cache the downloaded model for next time use. Otherwise, set `global.modelUseHostPath` to 'null' if you don't want to cache the model.

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are running.

Then run the command `kubectl port-forward svc/chatqna 8888:8888` to expose the service for access.

Open another terminal and run the following command to verify the service if working:

```console
curl http://localhost:8888/v1/chatqna \
    -H "Content-Type: application/json" \
    -d '{"messages": "What is the revenue of Nike in 2023?"}'
```

## Values

| Key                             | Type   | Default                       | Description                                                                                                                                                                                                                       |
| ------------------------------- | ------ | ----------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| image.repository                | string | `"opea/chatqna:latest"`       |                                                                                                                                                                                                                                   |
| service.port                    | string | `"8888"`                      |                                                                                                                                                                                                                                   |
| global.HUGGINGFACEHUB_API_TOKEN | string | `""`                          | Your own Hugging Face API token                                                                                                                                                                                                   |
| global.modelUseHostPath         | string | `"/mnt/opea-models"`          | Cached models directory, tgi will not download if the model is cached here. The host path "modelUseHostPath" will be mounted to container as /data directory. Set this to empty/null will force it to download models every time. |
| tgi.LLM_MODEL_ID                | string | `"Intel/neural-chat-7b-v3-3"` | Models id from https://huggingface.co/, or predownloaded model directory                                                                                                                                                          |
