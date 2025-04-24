# OPEA text2image microservice

Helm chart for deploying OPEA text2image service.

## Installing the Chart

To install the chart, run the following:

```console
cd GenAIInfra/helm-charts/common
export MODELDIR=/mnt/opea-models
export MODELNAME=stable-diffusion-v1-5/stable-diffusion-v1-5
export HFTOKEN="insert-your-huggingface-token-here"
helm install text2image text2image --set global.modelUseHostPath=${MODELDIR} --set MODEL=${MODELNAME} --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN}
# To deploy on Gaudi enabled kubernetes cluster
# helm install text2image text2image --set global.modelUseHostPath=${MODELDIR} --set MODEL=${MODELNAME} --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --values gaudi-values.yaml
```

By default, the text2image service will downloading the "stable-diffusion-v1-5/stable-diffusion-v1-5" which is about 45GB.

If you already cached the model locally, you can pass it to container like this example:

MODELDIR=/mnt/opea-models

MODELNAME="/data/models--stable-diffusion-v1-5--stable-diffusion-v1-5"

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are runinng and in ready state.

Then run the command `kubectl port-forward svc/text2image 9379:9379` to expose the text2image service for access.

Open another terminal and run the following command to verify the service if working:

```console
curl http://localhost:9379/v1/text2image \
    -XPOST \
    -d '{"prompt":"An astronaut riding a green horse", "num_images_per_prompt":1}' \
    -H 'Content-Type: application/json'
```

## Values

| Key                             | Type   | Default                                         | Description                                                                                                                                                                                                                  |
| ------------------------------- | ------ | ----------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| MODEL                           | string | `"stable-diffusion-v1-5/stable-diffusion-v1-5"` | Models id from https://huggingface.co/, or predownloaded model directory                                                                                                                                                     |
| global.HUGGINGFACEHUB_API_TOKEN | string | `insert-your-huggingface-token-here`            | Hugging Face API token                                                                                                                                                                                                       |
| global.modelUseHostPath         | string | `""`                                            | Cached models directory, text2image will not download if the model is cached here. The host path "modelUseHostPath" will be mounted to container as /data directory. Set this to null/empty will force it to download model. |
| autoscaling.enabled             | bool   | `false`                                         | Enable HPA autoscaling for the service deployment based on metrics it provides. See [HPA instructions](../../HPA.md) before enabling!                                                                                        |
| global.monitoring               | bool   | `false`                                         | Enable usage metrics for the service. Required for HPA. See [monitoring instructions](../../monitoring.md) before enabling!                                                                                                  |
