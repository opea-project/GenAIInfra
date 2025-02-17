# ChatQnA

Helm chart for deploying ChatQnA service. ChatQnA depends on the following services:

- [data-prep](../common/data-prep/README.md)
- [embedding-usvc](../common/embedding-usvc/README.md)
- [tei](../common/tei/README.md)
- [retriever-usvc](../common/retriever-usvc/README.md)
- [redis-vector-db](../common/redis-vector-db/README.md)
- [reranking-usvc](../common/reranking-usvc/README.md)
- [teirerank](../common/teirerank/README.md)
- [llm-uservice](../common/llm-uservice/README.md)
- [tgi](../common/tgi/README.md)
- [vllm](../common/vllm/README.md)

## Installing the Chart

To install the chart, run the following:

```console
cd GenAIInfra/helm-charts/
./update_dependency.sh
helm dependency update chatqna
export HFTOKEN="insert-your-huggingface-token-here"
export MODELDIR="/mnt/opea-models"
export MODELNAME="meta-llama/Meta-Llama-3-8B-Instruct"
# To use CPU with vLLM
helm install chatqna chatqna --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} --set vllm.LLM_MODEL_ID=${MODELNAME}
# To use Gaudi device with vLLM
#helm install chatqna chatqna --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} --set vllm.LLM_MODEL_ID=${MODELNAME} -f chatqna/gaudi-vllm-values.yaml
# To use CPU with TGI
#helm install chatqna chatqna --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} --set tgi.LLM_MODEL_ID=${MODELNAME} -f chatqna/cpu-tgi-values.yaml
# To use Gaudi device with TGI
#helm install chatqna chatqna --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} --set tgi.LLM_MODEL_ID=${MODELNAME} -f chatqna/gaudi-tgi-values.yaml
# To use Nvidia GPU with TGI
#helm install chatqna chatqna --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} --set tgi.LLM_MODEL_ID=${MODELNAME} -f chatqna/nv-values.yaml
# To include guardrail component in chatqna on Gaudi with TGI
#helm install chatqna chatqna --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} -f chatqna/guardrails-gaudi-values.yaml
# To run chatqna with Intel TDX feature
#helm install chatqna chatqna --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set vllm.LLM_MODEL_ID=${MODELNAME} --set redis-vector-db.tdxEnabled=true --set redis-vector-db.resources.limits.memory=4Gi --set retriever-usvc.tdxEnabled=true --set retriever-usvc.resources.limits.memory=7Gi --set tei.tdxEnabled=true --set tei.resources.limits.memory=4Gi --set teirerank.tdxEnabled=true --set teirerank.resources.limits.memory=6Gi --set nginx.tdxEnabled=true --set chatqna-ui.tdxEnabled=true --set chatqna-ui.resources.limits.memory=2Gi --set data-prep.tdxEnabled=true --set data-prep.resources.limits.memory=11Gi --set vllm.tdxEnabled=true --set vllm.resources.limits.memory=80Gi

# To use CPU with vLLM with Qdrant DB
helm install chatqna chatqna --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} --set vllm.LLM_MODEL_ID=${MODELNAME} -f chatqna/cpu-qdrant-values.yaml
```

### IMPORTANT NOTE

1. Make sure your `MODELDIR` exists on the node where your workload is scheduled so you can cache the downloaded model for next time use. Otherwise, set `global.modelUseHostPath` to 'null' if you don't want to cache the model.

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are running.

Curl command and UI are the two options that can be leveraged to verify the result.

### Verify the workload through curl command

Run the command `kubectl port-forward svc/chatqna 8888:8888` to expose the service for access.

Open another terminal and run the following command to verify the service if working:

```console
curl http://localhost:8888/v1/chatqna \
    -H "Content-Type: application/json" \
    -d '{"messages": "What is the revenue of Nike in 2023?"}'
```

### Verify the workload through UI

The UI has already been installed via the Helm chart. To access it, use the external IP of one your Kubernetes node along with the NGINX port. You can find the NGINX port using the following command:

```bash
export port=$(kubectl get service chatqna-nginx --output='jsonpath={.spec.ports[0].nodePort}')
echo $port
```

Open a browser to access `http://<k8s-node-ip-address>:${port}` to play with the ChatQnA workload.

## Values

| Key               | Type   | Default                                 | Description                                                                            |
| ----------------- | ------ | --------------------------------------- | -------------------------------------------------------------------------------------- |
| image.repository  | string | `"opea/chatqna"`                        |                                                                                        |
| service.port      | string | `"8888"`                                |                                                                                        |
| tgi.LLM_MODEL_ID  | string | `"meta-llama/Meta-Llama-3-8B-Instruct"` | Inference models for TGI                                                               |
| vllm.LLM_MODEL_ID | string | `"meta-llama/Meta-Llama-3-8B-Instruct"` | Inference models for vLLM                                                              |
| global.monitoring | bool   | `false`                                 | Enable usage metrics for the service components. See ../monitoring.md before enabling! |

## Troubleshooting

If you encounter any issues, please refer to [ChatQnA Troubleshooting](troubleshooting.md)
