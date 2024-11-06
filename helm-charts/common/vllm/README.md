# vllm

Helm chart for deploying vLLM Inference service.

Refer to [Deploy with Helm Charts](../../README.md) for global guides.

## Installing the Chart

To install the chart, run the following:

Note that you cannot use vllm as the service release name due to [environment variables conflict](https://docs.vllm.ai/en/stable/serving/env_vars.html#environment-variables).

```bash
cd GenAIInfra/helm-charts/common
export MODELDIR=/mnt/opea-models
export MODELNAME="Intel/neural-chat-7b-v3-3"
export HFTOKEN="insert-your-huggingface-token-here"

# If you are behind a proxy, please export the appropriate proxy values.
export http_proxy=<your_http_proxy>
export https_proxy=<your_https_proxy>

```

- Deploy on XEON device:

```bash
helm install myvllm vllm --set global.modelUseHostPath=${MODELDIR} --set LLM_MODEL_ID=${MODELNAME} --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN}
```

- To deploy on Gaudi enabled Kubernetes cluster:

```bash
helm install myvllm vllm --set global.modelUseHostPath=${MODELDIR} --set LLM_MODEL_ID=${MODELNAME} --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --values vllm/gaudi-values.yaml
```

- To deploy OpenVINO optimized vLLM on XEON device:

```bash
helm -f vllm/openvino-values.yaml install myvllm vllm --set global.modelUseHostPath=${MODELDIR} --set LLM_MODEL_ID=${MODELNAME} --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set global.http_proxy=${http_proxy} --set global.https_proxy=${https_proxy}
```

By default, the vLLM service will download "Intel/neural-chat-7b-v3-3" model. If you already cached the model locally, you can pass it to container like this example:

```bash
MODELDIR=/mnt/opea-models

MODELNAME="facebook/opt-125m"
```

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are runinng.

Then run the command `kubectl port-forward svc/myvllm 2080:80` to expose the vllm service for access.

Open another terminal and run the following command to verify the service if working:

```console
curl http://localhost:2080/v1/completions \
  -H "Content-Type: application/json" \
  -d '{"model": "Intel/neural-chat-7b-v3-3", "prompt": "What is Deep Learning?", "max_tokens": 32, "temperature": 0}'
```

## Values

| Key                             | Type   | Default                              | Description                                                                                                                                                                                                            |
| ------------------------------- | ------ | ------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| LLM_MODEL_ID                    | string | `"Intel/neural-chat-7b-v3-3"`        | Models id from https://huggingface.co/, or predownloaded model directory                                                                                                                                               |
| global.HUGGINGFACEHUB_API_TOKEN | string | `insert-your-huggingface-token-here` | Hugging Face API token                                                                                                                                                                                                 |
| global.modelUseHostPath         | string | `""`                                 | Cached models directory, vllm will not download if the model is cached here. The host path "modelUseHostPath" will be mounted to container as /data directory. Set this to null/empty will force it to download model. |
| image.repository                | string | `"opea/vllm"`                        |                                                                                                                                                                                                                        |
| image.tag                       | string | `"latest"`                           |                                                                                                                                                                                                                        |
