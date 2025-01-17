# SearchQnA

Helm chart for deploying SearchQnA service.

SearchQnA depends on the following helm charts(micro services):

- [tei](../common/tei)
- [embedding-usvc](../common/embedding-usvc)
- [web-retriever](../common/web-retriever)
- [teirerank](../common/teirerank)
- [reranking-usvc](../common/reranking-usvc)
- [tgi](../common/tgi)
- [llm-uservice](../common/llm-uservice)
- [ui](../common/ui)
- [nginx](../common/nginx)

## Installing the Chart

To install the chart, run the following:

```console
cd GenAIInfra/helm-charts/
./update_dependency.sh
helm dependency update searchqna
export MODELDIR="/mnt/opea-models"
export MODEL="Intel/neural-chat-7b-v3-3"
export HFTOKEN="insert-your-huggingface-token-here"
export GOOGLE_API_KEY="insert-your-google-api-key-here"
export GOOGLE_CSE_ID="insert-your-google-search-engine-id-here"

# To run on Xeon
helm install searchqna searchqna --set global.modelUseHostPath=${MODELDIR} --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set web-retriever.GOOGLE_API_KEY=${GOOGLE_API_KEY} --web-retriever.GOOGLE_CSE_ID=${GOOGLE_CSE_ID} --set tgi.LLM_MODEL_ID=${MODEL} --set llm-uservice.LLM_MODEL_ID=${MODEL}

# To run on Gaudi
# helm install searchqna searchqna --set global.modelUseHostPath=${MODELDIR} --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set web-retriever.GOOGLE_API_KEY=${GOOGLE_API_KEY} --web-retriever.GOOGLE_CSE_ID=${GOOGLE_CSE_ID} --set tgi.LLM_MODEL_ID=${MODEL} --set llm-uservice.LLM_MODEL_ID=${MODEL} -f gaudi-values.yaml
```

### IMPORTANT NOTE

1. Make sure your `MODELDIR` exists on the node where your workload is scheduled so you can cache the downloaded model for next time use. Otherwise, set `global.modelUseHostPath` to 'null' if you don't want to cache the model. This workload by default will download model `Intel/neural-chat-7b-v3-3`, `BAAI/bge-base-en-v1.5`, `BAAI/bge-reranker-base` for inferencing, embedding, reranking respectively.

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are running.

### Verify the workload through curl command

Then run the command `kubectl port-forward svc/searchqna 3008:3008` to expose the service for access.

Open another terminal and run the following command to verify the service if working:

```console
curl http://localhost:3008/v1/searchqna \
  -X POST \
  -d '{"messages": "What is the latest news? Give me also the source link.", "stream": "True"}' \
  -H 'Content-Type: application/json'
```

### Verify the workload through UI

The UI has already been installed via the Helm chart. To access it, use the external IP of one your Kubernetes node along with the NGINX port. You can find the NGINX port using the following command:

```bash
export port=$(kubectl get service searchqna-nginx --output='jsonpath={.spec.ports[0].nodePort}')
echo $port
```

Open a browser to access `http://<k8s-node-ip-address>:${port}` to play with.

## Values

| Key                       | Type   | Default                     | Description                                                                            |
| ------------------------- | ------ | --------------------------- | -------------------------------------------------------------------------------------- |
| image.repository          | string | `"opea/searchqna"`          |                                                                                        |
| service.port              | string | `"3008"`                    |                                                                                        |
| tgi.LLM_MODEL_ID          | string | `Intel/neural-chat-7b-v3-3` | inference model                                                                        |
| llm_uservice.LLM_MODEL_ID | string | `Intel/neural-chat-7b-v3-3` | should be the same as `tgi.LLM_MODEL_ID`                                               |
| global.monitoring         | bool   | `false`                     | Enable usage metrics for the service components. See ../monitoring.md before enabling! |
