# llm-ctrl Microservice

Helm chart for deploying LLM controller microservice which facilitates connections and handles responses from OpenVINO vLLM microservice.

`llm-ctrl-uservice` depends on vLLM microservice. You should properly set `vLLM_ENDPOINT` as the HOST URI of vLLM microservice. If not set, it will consider the default value : `http://<helm-release-name>-vllm:80`

As this service depends on vLLM microservice, we can proceed in either of 2 ways:

- Install both microservices individually.
- Install the vLLM microservice as dependency for `llm-ctrl-uservice` microservice.

## (Option 1): Installing the charts individually:

First, you need to install the `vllm` chart, please refer to the [vllm](../vllm) chart for more information.

After you've deployed the `vllm` chart successfully, please run `kubectl get svc` to get the vLLM service name with port. We need to provide this to `llm-ctrl-uservice` as a value for vLLM_ENDPOINT for letting it discover and connect to the vLLM microservice.

> **_NOTE:_** While installing charts separately, if you don't provide any vLLM endpoint explicitly, it will take the default endpoint as `http://<helm-release-name>-vllm:80`. So, if you are not providing the vLLM endpoint explicitly, please make sure to provide same helm release name to both the charts while installing.

Get the service name for vLLM deployment by running: `kubectl get svc`. In the current case, service name would be `myvllm`.

> **_NOTE:_** Please add the service name for vLLM to the value of no_proxy env var, if you are behind a proxy.

To install the chart, run the following:

```bash
cd GenAIInfra/helm-charts/common/llm-ctrl-uservice
export HFTOKEN="insert-your-huggingface-token-here"
export vLLM_ENDPOINT="http://myvllm"
export MODELNAME="Intel/neural-chat-7b-v3-3"

# If proxy is required, please export the appropriate proxy values.
export http_proxy=<your_http_proxy>
export https_proxy=<your_https_proxy>

helm dependency update
helm install llm-ctrl-uservice . --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set vLLM_ENDPOINT=${vLLM_ENDPOINT} --set LLM_MODEL_ID=${MODELNAME} --set global.http_proxy=${http_proxy} --set global.https_proxy=${https_proxy} --wait
```

## (Option 2): Installing the chart with automatic installation of dependency:

```bash
cd GenAIInfra/helm-charts/common/llm-ctrl-uservice
export HFTOKEN="insert-your-huggingface-token-here"
export MODELDIR="/mnt/opea-models"
export MODELNAME="Intel/neural-chat-7b-v3-3"

# If proxy is required, please export the appropriate proxy values.
export http_proxy=<your_http_proxy>
export https_proxy=<your_https_proxy>

helm dependency update
helm install llm-ctrl-uservice . --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} --set LLM_MODEL_ID=${MODELNAME} --set vllm.LLM_MODEL_ID=${MODELNAME} --set autodependency.enabled=true --set global.http_proxy=${http_proxy} --set global.https_proxy=${https_proxy} --wait
```

`--wait` flag in above installation command will make sure that all the dependencies are resolved and all services are deployed.

## Verify

To verify the installation, run the following command to make sure all pods are running.

```bash
kubectl get pod
```

Once you see `llm-ctrl-uservice` pod and `llm-ctrl-uservice-vllm` pod in ready and running state, run the following command:

```bash
kubectl port-forward svc/llm-ctrl-uservice 9000:9000
```

This exposes the port 9000, on which `llm-ctrl-uservice` is running inside the pod, at port 9000 on the host.

Now, we can access the service from the host machine. Open another terminal and run the following command to verify whether `llm-ctrl-uservice` is working:

```bash
curl http://localhost:9000/v1/chat/completions \
    -X POST \
    -d '{"query":"What is Deep Learning?","max_new_tokens":17,"top_k":10,"top_p":0.95,"typical_p":0.95,"temperature":0.01,"repetition_penalty":1.03,"streaming":true}' \
    -H 'Content-Type: application/json'
```

## Values

| Key                             | Type   | Default              | Description                                                                                                                                                   |
| ------------------------------- | ------ | -------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| global.HUGGINGFACEHUB_API_TOKEN | string | `""`                 | Your own Hugging Face API token                                                                                                                               |
| global.modelUseHostPath         | string | `"/mnt/opea-models"` | Cached models directory, vLLM will not download if the model is cached here. The host path "modelUseHostPath" will be mounted to container as /data directory |
| image.repository                | string | `"opea/llm-vllm"`    |                                                                                                                                                               |
| service.port                    | string | `"9000"`             |                                                                                                                                                               |
| vLLM_ENDPOINT                   | string | `""`                 | OpenVINO vLLM service endpoint                                                                                                                                |
