# CodeGen

Helm chart for deploying CodeGen service.

CodeGen depends on LLM and tgi microservice, refer to [llm-uservice](../common/llm-uservice/README.md) and [tgi](../common/tgi/README.md) for more config details.

## Installing the Chart

To install the chart, run the following:

```console
cd GenAIInfra/helm-charts/
./update_dependency.sh
helm dependency update codegen
export HFTOKEN="insert-your-huggingface-token-here"
export MODELDIR="/mnt/opea-models"
export MODELNAME="Qwen/Qwen2.5-Coder-7B-Instruct"
# To run on Xeon
helm install codegen codegen --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} --set tgi.LLM_MODEL_ID=${MODELNAME}
# To run on Gaudi
#helm install codegen codegen --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} --set tgi.LLM_MODEL_ID=${MODELNAME} -f codegen/gaudi-values.yaml
```

### IMPORTANT NOTE

1. Make sure your `MODELDIR` exists on the node where your workload is schedueled so you can cache the downloaded model for next time use. Otherwise, set `global.modelUseHostPath` to 'null' if you don't want to cache the model.

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are running.

Curl command and UI are the two options that can be leveraged to verify the result.

### Verify the workload through curl command

Then run the command `kubectl port-forward svc/codegen 7778:7778` to expose the service for access.

Open another terminal and run the following command to verify the service if working:

```console
curl http://localhost:7778/v1/codegen \
    -H "Content-Type: application/json" \
    -d '{"messages": "Implement a high-level API for a TODO list application. The API takes as input an operation request and updates the TODO list in place. If the request is invalid, raise an exception."}'
```

### Verify the workload through UI

The UI has already been installed via the Helm chart. To access it, use the external IP of one your Kubernetes node along with the NGINX port. You can find the NGINX port using the following command:

```bash
export port=$(kubectl get service codegen-nginx --output='jsonpath={.spec.ports[0].nodePort}')
echo $port
```

Open a browser to access `http://<k8s-node-ip-address>:${port}` to play with the ChatQnA workload.

## Values

| Key               | Type   | Default                            | Description                                                                            |
| ----------------- | ------ | ---------------------------------- | -------------------------------------------------------------------------------------- |
| image.repository  | string | `"opea/codegen"`                   |                                                                                        |
| service.port      | string | `"7778"`                           |                                                                                        |
| tgi.LLM_MODEL_ID  | string | `"Qwen/Qwen2.5-Coder-7B-Instruct"` | Models id from https://huggingface.co/, or predownloaded model directory               |
| global.monitoring | bool   | `false`                            | Enable usage metrics for the service components. See ../monitoring.md before enabling! |
