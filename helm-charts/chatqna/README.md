# ChatQnA

Helm chart for deploying ChatQnA service. ChatQnA depends on the following services:

- [data-prep](../common/data-prep/README.md)
- [embedding-usvc](../common/embedding-usvc/README.md)
- [tei](../common/tei/README.md)
- [retriever-usvc](../common/retriever-usvc/README.md)
- [redis-vector-db](../common/redis-vector-db/README.md)
- [reranking-usvc](../common/reranking-usvc/README.md)
- [teirerank](../common/teirerank/README.md)

For LLM inference, two more microservices will be required. We can either use [TGI](https://github.com/huggingface/text-generation-inference) or [vLLM](https://github.com/vllm-project/vllm) as our LLM backend. Depending on that, we will have following microservices as part of dependencies for ChatQnA application.

1. For using **TGI** as an inference service, following 2 microservices will be required:

   - [llm-uservice](../common/llm-uservice/README.md)
   - [tgi](../common/tgi/README.md)

2. For using **vLLM** as an inference service, following 2 microservices would be required:

   - [llm-ctrl-uservice](../common/llm-ctrl-uservice/README.md)
   - [vllm](../common/vllm/README.md)

> **_NOTE :_** We shouldn't have both inference engine deployed. It is required to only setup either of them. To achieve this, conditional flags are added in the chart dependency. We will be switching off flag corresponding to one service and switching on the other, in order to have a proper setup of all ChatQnA dependencies.

## Installing the Chart

Please follow the following steps to install the ChatQnA Chart:

1. Clone the GenAIInfra repository:

```bash
git clone https://github.com/opea-project/GenAIInfra.git
```

2. Setup the dependencies and required environment variables:

```bash
cd GenAIInfra/helm-charts/
./update_dependency.sh
helm dependency update chatqna
export HFTOKEN="insert-your-huggingface-token-here"
export MODELDIR="/mnt/opea-models"
export MODELNAME="Intel/neural-chat-7b-v3-3"
```

3. Depending on the device which we are targeting for running ChatQnA, please use one the following installation commands:

```bash
# Install the chart on a Xeon machine

# If you would like to use the traditional UI, please change the image as well as the containerport within the values
# append these at the end of the command "--set chatqna-ui.image.repository=opea/chatqna-ui,chatqna-ui.image.tag=latest,chatqna-ui.containerPort=5173"

helm install chatqna chatqna --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} --set tgi.LLM_MODEL_ID=${MODELNAME}
```

```bash
# To use Gaudi device
helm install chatqna chatqna --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} --set tgi.LLM_MODEL_ID=${MODELNAME} -f chatqna/gaudi-values.yaml
```

```bash
# To use Nvidia GPU
helm install chatqna chatqna --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} --set tgi.LLM_MODEL_ID=${MODELNAME} -f chatqna/nv-values.yaml
```

```bash
# To include guardrail component in chatqna on Xeon
helm install chatqna chatqna --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} -f chatqna/guardrails-values.yaml
```

```bash
# To include guardrail component in chatqna on Gaudi
helm install chatqna chatqna --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} -f chatqna/guardrails-gaudi-values.yaml
```

> **_NOTE :_** Default installation will use [TGI (Text Generation Inference)](https://github.com/huggingface/text-generation-inference) as inference engine. To use vLLM as inference engine, please see below.

```bash
# To use vLLM inference engine on XEON device

helm install chatqna chatqna --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} --set llm-ctrl-uservice.LLM_MODEL_ID=${MODELNAME} --set vllm.LLM_MODEL_ID=${MODELNAME} --set tgi.enabled=false --set vllm.enabled=true

# To use OpenVINO optimized vLLM inference engine on XEON device

helm -f ./chatqna/vllm-openvino-values.yaml install chatqna chatqna --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} --set llm-ctrl-uservice.LLM_MODEL_ID=${MODELNAME} --set vllm.LLM_MODEL_ID=${MODELNAME}
```

### IMPORTANT NOTE

1. Make sure your `MODELDIR` exists on the node where your workload is scheduled so you can cache the downloaded model for next time use. Otherwise, set `global.modelUseHostPath` to 'null' if you don't want to cache the model.

2. Please set `http_proxy`, `https_proxy` and `no_proxy` values while installing chart, if you are behind a proxy.

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are running.

Curl command and UI are the two options that can be leveraged to verify the result.

### Verify the workload through curl command

Run the command `kubectl port-forward svc/chatqna 8888:8888` to expose the service for access.

Open another terminal and run the following command to verify the service if working:

```bash
curl http://localhost:8888/v1/chatqna \
    -X POST \
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

| Key                        | Type   | Default                       | Description                                                                            |
| -------------------------- | ------ | ----------------------------- | -------------------------------------------------------------------------------------- |
| image.repository           | string | `"opea/chatqna"`              |                                                                                        |
| service.port               | string | `"8888"`                      |                                                                                        |
| tgi.LLM_MODEL_ID           | string | `"Intel/neural-chat-7b-v3-3"` | Models id from https://huggingface.co/, or predownloaded model directory               |
| vllm-openvino.LLM_MODEL_ID | string | `"Intel/neural-chat-7b-v3-3"` | Models id from https://huggingface.co/, or predownloaded model directory               |
| global.monitoring          | bop;   | false                         | Enable usage metrics for the service components. See ../monitoring.md before enabling! |

## Troubleshooting

If you encounter any issues, please refer to [ChatQnA Troubleshooting](troubleshooting.md)
