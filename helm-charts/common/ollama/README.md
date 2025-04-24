# ollama

Helm chart for deploying Ollama model server.

## Installing the Chart

To install the chart, run the following:

```console
cd GenAIInfra/helm-charts/common
export MODELNAME="llama3.2"

helm install ollama-release ollama --set OLLAMA_MODEL=${MODELNAME}
```

By default, the ollama container will download the "llama3.2:1b" model, which is about 1.3GB.

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are running.

Then run the command `kubectl port-forward svc/ollama-release 11434:80` to expose the ollama service for access.

Open another terminal and run the following command to verify the service is working:

```console
curl http://localhost:11434/api/generate -d '{
  "model": "llama3.2:1b",
  "prompt": "What is Deep Learning?",
  "options": {
    "num_predict": 40
  }
}'
```

## Values

| Key                     | Type   | Default         | Description                                                                                                                                                                                                                                                                                                                |
| ----------------------- | ------ | --------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| LLM_MODEL_ID            | String | `"llama3.2:1b"` | The model ID to use. Must be one of the models listed in the [Ollama Library](https://ollama.com/library)                                                                                                                                                                                                                  |
| global.modelUseHostPath | String | `""`            | Cached models directory on Kubernetes node, service will not download if the model is cached here. The host path "modelUseHostPath" will be mounted to the container as /.ollama directory. Setting this to null/empty will force the container to download the model. May not be set if "global.modelUsePVC" is also set. |
| global.modelUsePVC      | String | `""`            | Name of Persistent Volume Claim to use for model cache. The Persistent Volume will be mounted to the container as /.ollama directory. Setting this to null/empty will force the container to download the model. May not be set if "global.modelUseHostPath" is also set.                                                  |
