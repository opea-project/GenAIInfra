# tgi

Helm chart for deploying Hugging Face Text Generation Inference service.

## Installing the Chart

Before installation, you need to export `HFTOKEN`.
```console
export HFTOKEN="insert-your-huggingface-token-here"
helm repo add opea https://opea-project.github.io/GenAIInfra
helm install my-tgi opea/tgi --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN}
```

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are runinng.

Then run the command `kubectl port-forward svc/tgi 2080:80` to expose the tgi service for access.

Open another terminal and run the following command to verify the service if working:

```console
curl http://localhost:2080/generate \
    -X POST \
    -d '{"inputs":"What is Deep Learning?","parameters":{"max_new_tokens":17, "do_sample": true}}' \
    -H 'Content-Type: application/json'
```

## Values

| Key                             | Type   | Default                                           | Description                                                                                                                                                                                                           |
| ------------------------------- | ------ | ------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| LLM_MODEL_ID                    | string | `"bigscience/bloom-560m"`                         | Models id from https://huggingface.co/, or predownloaded model directory                                                                                                                                              |
| global.HUGGINGFACEHUB_API_TOKEN | string | `insert-your-huggingface-token-here`              | Hugging Face API token                                                                                                                                                                                                |
| global.modelUseHostPath         | string | `"/mnt/opea-models"`                              | Cached models directory, tgi will not download if the model is cached here. The host path "modelUseHostPath" will be mounted to container as /data directory. Set this to null/empty will force it to download model. |
| image.repository                | string | `"ghcr.io/huggingface/text-generation-inference"` |                                                                                                                                                                                                                       |
| image.tag                       | string | `"1.4"`                                           |                                                                                                                                                                                                                       |
| horizontalPodAutoscaler.enabled | bool   | false                                             | Enable HPA autoscaling for the service deployment based on metrics it provides. See [HPA instructions](../../HPA.md) before enabling!                                                                                 |
| global.monitoring               | bool   | false                                             | Enable usage metrics for the service. Required for HPA. See [monitoring instructions](../../monitoring.md) before enabling!                                                                                           |
