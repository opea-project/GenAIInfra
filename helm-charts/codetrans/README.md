# CodeTrans

Helm chart for deploying CodeTrans service.

CodeTrans depends on LLM microservice, refer to llm-uservice for more config details.

## Installing the Chart

To install the chart, run the following:

```console
cd GenAIInfra/helm-charts/
scripts/update_dependency.sh
helm dependency update codetrans
export HFTOKEN="insert-your-huggingface-token-here"
export MODELDIR="/mnt/opea-models"
export MODELNAME="Qwen/Qwen2.5-Coder-7B-Instruct"
# To use CPU with vLLM
helm install codetrans codetrans --set global.HF_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} --set llm-uservcie.LLM_MODEL_ID=${MODELNAME} --set vllm.LLM_MODEL_ID=${MODELNAME} -f codetrans/cpu-values.yaml
# To use CPU with TGI
# helm install codetrans codetrans --set global.HF_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} --set llm-uservcie.LLM_MODEL_ID=${MODELNAME} --set tgi.LLM_MODEL_ID=${MODELNAME} -f codetrans/cpu-tgi-values.yaml
# To use Gaudi device with vLLM
# helm install codetrans codetrans --set global.HF_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} --set llm-uservcie.LLM_MODEL_ID=${MODELNAME} --set vllm.LLM_MODEL_ID=${MODELNAME} -f codetrans/gaudi-values.yaml
# To use Gaudi device with TGI
# helm install codetrans codetrans --set global.HF_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} --set llm-uservcie.LLM_MODEL_ID=${MODELNAME} --set tgi.LLM_MODEL_ID=${MODELNAME} -f codetrans/gaudi-tgi-values.yaml
# To use AMD ROCm device with vLLM
# helm install codetrans codetrans --set global.HF_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} --set llm-uservcie.LLM_MODEL_ID=${MODELNAME} --set vllm.LLM_MODEL_ID=${MODELNAME} -f codetrans/rocm-values.yaml
# To use AMD ROCm device with TGI
# helm install codetrans codetrans --set global.HF_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} --set llm-uservcie.LLM_MODEL_ID=${MODELNAME} --set tgi.LLM_MODEL_ID=${MODELNAME} -f codetrans/rocm-tgi-values.yaml
```

### IMPORTANT NOTE

1. To use model `Qwen/Qwen2.5-Coder-7B-Instruct`, you should first goto the [huggingface model card](https://huggingface.co/mistralai/Mistral-7B-Instruct-v0.3) to apply for the model access first. You need to make sure your huggingface token has at least read access to that model.

2. Make sure your `MODELDIR` exists on the node where your workload is schedueled so you can cache the downloaded model for next time use. Otherwise, set `global.modelUseHostPath` to 'null' if you don't want to cache the model.

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are running.

Curl command and UI are the two options that can be leveraged to verify the result.

### Verify the workload through curl command

Then run the command `kubectl port-forward svc/codetrans 7777:7777` to expose the CodeTrans service for access.

Open another terminal and run the following command to verify the service if working:

```console
curl http://localhost:7777/v1/codetrans \
    -H 'Content-Type: application/json' \
    -d '{"language_from": "Golang","language_to": "Python","source_code": "package main\n\nimport \"fmt\"\nfunc main() {\n    fmt.Println(\"Hello, World!\");\n}"}'
```

### Verify the workload through UI

The UI has already been installed via the Helm chart. To access it, use the external IP of one your Kubernetes node along with the NGINX port. You can find the NGINX port using the following command:

```bash
export port=$(kubectl get service codetrans-nginx --output='jsonpath={.spec.ports[0].nodePort}')
echo $port
```

Open a browser to access `http://<k8s-node-ip-address>:${port}` to play with the ChatQnA workload.

## Values

| Key               | Type   | Default                            | Description                                                                            |
| ----------------- | ------ | ---------------------------------- | -------------------------------------------------------------------------------------- |
| image.repository  | string | `"opea/codetrans"`                 |                                                                                        |
| service.port      | string | `"7777"`                           |                                                                                        |
| tgi.LLM_MODEL_ID  | string | `"Qwen/Qwen2.5-Coder-7B-Instruct"` | Models id from https://huggingface.co/, or predownloaded model directory               |
| global.monitoring | bool   | `false`                            | Enable usage metrics for the service components. See ../monitoring.md before enabling! |
