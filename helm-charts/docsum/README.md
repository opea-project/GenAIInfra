# DocSum

Helm chart for deploying DocSum service.

DocSum depends on LLM microservice, refer to llm-uservice for more config details.

## Installing the Chart

To install the chart, run the following:

```console
cd GenAIInfra/helm-charts/
./update_dependency.sh
helm dependency update docsum
export HFTOKEN="insert-your-huggingface-token-here"
export MODELDIR="/mnt/opea-models"
export MODELNAME="Intel/neural-chat-7b-v3-3"
helm install docsum docsum --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} --set tgi.LLM_MODEL_ID=${MODELNAME}
# To use Gaudi device
# helm install docsum docsum --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --values docsum/gaudi-values.yaml
```

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are running.

Then run the command `kubectl port-forward svc/docsum 8888:8888` to expose the DocSum service for access.

Open another terminal and run the following command to verify the service if working:

```console
curl http://localhost:8888/v1/docsum \
    -H 'Content-Type: application/json' \
    -d '{"messages": "Text Embeddings Inference (TEI) is a toolkit for deploying and serving open source text embeddings and sequence classification models. TEI enables high-performance extraction for the most popular models, including FlagEmbedding, Ember, GTE and E5."}'
```

## Values

| Key                             | Type   | Default                       | Description                                                                                                                                                  |
| ------------------------------- | ------ | ----------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| image.repository                | string | `"opea/docsum:latest"`        |                                                                                                                                                              |
| service.port                    | string | `"8888"`                      |                                                                                                                                                              |
| global.HUGGINGFACEHUB_API_TOKEN | string | `""`                          | Your own Hugging Face API token                                                                                                                              |
| global.modelUseHostPath         | string | `"/mnt/opea-models"`          | Cached models directory, tgi will not download if the model is cached here. The host path "modelUseHostPath" will be mounted to container as /data directory |
| tgi.LLM_MODEL_ID                | string | `"Intel/neural-chat-7b-v3-3"` | Models id from https://huggingface.co/, or predownloaded model directory                                                                                     |
