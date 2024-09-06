# OpenVINO vLLM

Helm chart for deploying OpenVINO optimized vLLM Inference service.

## Installing the Chart

To install the chart, run the following:

```bash
cd GenAIInfra/helm-charts/common
export MODELDIR=/mnt/opea-models
export MODELNAME="bigscience/bloom-560m"
export HFTOKEN="insert-your-huggingface-token-here"

# If proxy is required, please export the appropriate proxy values.
export http_proxy=<your_http_proxy>
export https_proxy=<your_https_proxy>

helm install vllm-openvino vllm-openvino --set global.modelUseHostPath=${MODELDIR} --set LLM_MODEL_ID=${MODELNAME} --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set global.http_proxy=${http_proxy} --set global.https_proxy=${https_proxy} --wait
```

`--wait` flag in the above helm installation command lets the shell wait till `vllm-openvino` is completely up and ready.

> **_NOTE:_** Make sure your `MODELDIR` exists on the node where your workload is scheduled so you can cache the downloaded model for next time use. Otherwise, set `global.modelUseHostPath` to 'null' if you don't want to cache the model.

If you already cached the model locally, you can pass it to container like this example:

MODELDIR=/mnt/opea-models

MODELNAME="/data/models--bigscience--bloom-560m"

> **_NOTE:_** By default, the vLLM service will be downloading **Intel/neural-chat-7b-v3-3** model from Huggingface, which is around 4GB in size. To use a smaller model, please set the LLM_MODEL_ID value to your desired model, as shown above, while installing the chart.

## Verify

To verify the installation, run the following command to make sure all pods are running. Please note that it may take a while to come the vLLM pod in ready state.

```bash
kubectl get pod
```

Once you see `vllm-openvino` pod in ready and running state, run the following command:

```bash
 kubectl port-forward svc/vllm-openvino 2080:80
```

This exposes the port 80, on which `vllm-openvino` service is running inside the pod, at port 2080 on the host.

Now, we can access the service from the host machine. Open another terminal and run the following command to verify whether `vllm-openvino` service is working:

```bash
curl http://localhost:2080/v1/completions -sS --fail-with-body \
    -X POST \
    -d '{"prompt":"What is Deep Learning?", "model": "bigscience/bloom-560m", "max_tokens":17, "temperature": 0.5}' \
    -H 'Content-Type: application/json'
```

## Values

| Key                             | Type   | Default                              | Description                                                                                                                                                                                                                |
| ------------------------------- | ------ | ------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| LLM_MODEL_ID                    | string | `"bigscience/bloom-560m"`            | Models id from https://huggingface.co/, or predownloaded model directory                                                                                                                                                   |
| global.HUGGINGFACEHUB_API_TOKEN | string | `insert-your-huggingface-token-here` | Hugging Face API token                                                                                                                                                                                                     |
| global.modelUseHostPath         | string | `"/mnt/opea-models"`                 | Cached models directory, vLLM will not download if the model is cached here. The host path "modelUseHostPath" will be mounted to container as /data directory. Setting this to null/empty will force it to download model. |
| image.repository                | string | `"vllm"`                             |                                                                                                                                                                                                                            |
| image.tag                       | string | `"openvino"`                         |                                                                                                                                                                                                                            |
| horizontalPodAutoscaler.enabled | bool   | false                                | Enable HPA autoscaling for the service deployment based on metrics it provides. See HPA section in ../../README.md before enabling!                                                                                        |
