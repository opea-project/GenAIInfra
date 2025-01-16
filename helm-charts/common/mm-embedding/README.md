# OPEA mm-embedding microservice

Helm chart for deploying OPEA multimodal embedding service.

## Installing the Chart

To install the chart, run the following:

```console
cd GenAIInfra/helm-charts/common
export MODELDIR=/mnt/opea-models
export HFTOKEN="insert-your-huggingface-token-here"
# To deploy embedding-multimodal-bridgetower microserice on CPU
helm install mm-embedding mm-embedding --set global.modelUseHostPath=${MODELDIR} --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN}
# To deploy embedding-multimodal-bridgetower microserice on Gaudi
# helm install mm-embedding mm-embedding --set global.modelUseHostPath=${MODELDIR} --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --values mm-embedding/gaudi-values.yaml
# To deploy embedding-multimodal-clip microserice on CPU
# helm install mm-embedding mm-embedding --set global.modelUseHostPath=${MODELDIR} --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --values mm-embedding/variant_clip-values.yaml
```

By default, the embedding-multimodal-bridgetower service will download the "BridgeTower/bridgetower-large-itm-mlm-itc" model which is about 3.5GB, and the embedding-multimodal-clip service will download the "openai/clip-vit-base-patch32" model which is about 1.7GB.

If you already cached the model locally, you can pass it to container like this example:

MODELDIR=/mnt/opea-models

MODELNAME="/data/models--BridgeTower--bridgetower-large-itm-mlm-itc"

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are runinng and in ready state.

Then run the command `kubectl port-forward svc/mm-embedding 6990:6990` to expose the mm-embedding service for access.

Open another terminal and run the following command to verify the service if working:

```console
# Verify with embedding-multimodal-bridgetower
curl http://localhost:6990/v1/encode \
    -XPOST \
    -d '{"text":"This is example"}' \
    -H 'Content-Type: application/json'

# Verify with embedding-multimodal-clip
curl http://localhost:6990/v1/embeddings \
    -XPOST \
    -d '{"text":"This is example"}' \
    -H 'Content-Type: application/json'
```

## Values

| Key                             | Type   | Default                              | Description                                                                                                                                                                                                               |
| ------------------------------- | ------ | ------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| global.HUGGINGFACEHUB_API_TOKEN | string | `insert-your-huggingface-token-here` | Hugging Face API token                                                                                                                                                                                                    |
| global.modelUseHostPath         | string | `""`                                 | Cached models directory, service will not download if the model is cached here. The host path "modelUseHostPath" will be mounted to container as /data directory. Set this to null/empty will force it to download model. |
| autoscaling.enabled             | bool   | `false`                              | Enable HPA autoscaling for the service deployment based on metrics it provides. See [HPA instructions](../../HPA.md) before enabling!                                                                                     |
| global.monitoring               | bool   | `false`                              | Enable usage metrics for the service. Required for HPA. See [monitoring instructions](../../monitoring.md) before enabling!                                                                                               |
