# OPEA ipexllm microservice

Helm chart for deploying OPEA ipexllm service, using vllm backend on Intel GPU.

Refer to [Deploy with Helm Charts](../../README.md) for global guides.

## Installing the Chart

To install the chart, make sure you have installed run the following:

```console
cd GenAIInfra/helm-charts/common
export MODELDIR=/mnt/opea-models
export HFTOKEN="insert-your-huggingface-token-here"
export MODELNAME="deepseek-ai/DeepSeek-R1-Distill-Qwen-1.5B"

# To deploy ipexvllm microserice on Intel GPU
helm install ipexllm ipexllm --set global.modelUseHostPath=${MODELDIR} --set LLM_MODEL_ID=${MODELNAME} --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN}

```

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are running and in ready state.

Then run the command `kubectl port-forward svc/ipexllm 2080:80` to expose the corresponding K8s service for access.

Open another terminal and run the following command to verify the service if working:

```console
curl http://localhost:2080/v1/completions \
  -H "Content-Type: application/json" \
  -d '{"model": "deepseek-ai/DeepSeek-R1-Distill-Qwen-1.5B", "prompt": "What is Deep Learning?", "max_tokens": 32, "temperature": 0}'
```

## Values

| Key                             | Type    | Default                                       | Description                                                                                                                                                                                                            |
| ------------------------------- | ------- | --------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| LLM_MODEL_ID                    | string  | `"deepseek-ai/DeepSeek-R1-Distill-Qwen-1.5B"` | Models id from https://huggingface.co/, or predownloaded model directory                                                                                                                                               |
| DTYPE                           | string  | float16                                       | Data type for model weights and activations, possible values: auto, half, float16, bfloat16, float, float32                                                                                                            |
| QUANTIZATION                    | string  | fp8                                           | Model quantization accuracy, possible values: sym_int4, asym_int4, fp6, fp8, fp8_e4m3, fp8_35m2, fp16                                                                                                                  |
| PIPELINE_PARALLEL_SIZE          | integer | 1                                             | pipeline parallel size                                                                                                                                                                                                 |
| global.HUGGINGFACEHUB_API_TOKEN | string  | `insert-your-huggingface-token-here`          | HuggingFace API token                                                                                                                                                                                                  |
| global.modelUseHostPath         | string  | `""`                                          | Cached models directory, vllm will not download if the model is cached here. The host path "modelUseHostPath" will be mounted to container as /data directory. Set this to null/empty will force it to download model. |
| autoscaling.enabled             | bool    | `false`                                       | Enable HPA autoscaling for the service deployment based on metrics it provides. See [HPA instructions](../../HPA.md) before enabling!                                                                                  |
| global.monitoring               | bool    | `false`                                       | Enable usage metrics for the service. Required for HPA. See [monitoring instructions](../../monitoring.md) before enabling!                                                                                            |
