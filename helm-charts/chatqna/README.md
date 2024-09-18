# ChatQnA

Helm chart for deploying ChatQnA service. ChatQnA depends on the following services:

- [data-prep](../common/data-prep)
- [embedding-usvc](../common/embedding-usvc)
- [tei](../common/tei)
- [retriever-usvc](../common/retriever-usvc)
- [redis-vector-db](../common/redis-vector-db)
- [reranking-usvc](../common/reranking-usvc)
- [teirerank](../common/teirerank)

Apart from above mentioned services, there are following conditional dependencies (out of which, one are required):

1. If we want to use TGI as our inference service, following 2 services will be required:

   - [llm-uservice](../common/llm-uservice)
   - [tgi](../common/tgi)

2. If we want to use OpenVINO vLLM inference service, following 2 services would be required:
   - [llm-vllm-uservice](../common/llm-vllm-uservice)
   - [vllm-openvino](../common/vllm-openvino)

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

> **_NOTE:_** Default installation will use [TGI (Text Generation Inference)](https://github.com/huggingface/text-generation-inference) as inference engine. To use vLLM as inference engine, please see below.

```bash
# To use OpenVINO vLLM inference engine on Xeon device
helm install chatqna chatqna --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} --set llm-vllm-uservice.LLM_MODEL_ID=${MODELNAME} --set vllm.LLM_MODEL_ID=${MODELNAME} --set tags.tgi=false --set vllm.enabled=true
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

| Key                                    | Type   | Default                       | Description                                                                                                                             |
| -------------------------------------- | ------ | ----------------------------- | --------------------------------------------------------------------------------------------------------------------------------------- |
| image.repository                       | string | `"opea/chatqna"`              |                                                                                                                                         |
| service.port                           | string | `"8888"`                      |                                                                                                                                         |
| tgi.LLM_MODEL_ID                       | string | `"Intel/neural-chat-7b-v3-3"` | Models id from https://huggingface.co/, or predownloaded model directory                                                                |
| vllm-openvino.LLM_MODEL_ID             | string | `"Intel/neural-chat-7b-v3-3"` | Models id from https://huggingface.co/, or predownloaded model directory                                                                |
| global.horizontalPodAutoscaler.enabled | bool   | false                         | HPA autoscaling for the TGI and TEI service deployments based on metrics they provide. See HPA section in ../README.md before enabling! |
