# tgi

Helm chart for deploying Hugging Face Text Generation Inference service.

## Installing the Chart

To install the chart, run the following:

```console
cd GenAIInfra/helm-charts/common
export MODELDIR=/mnt/opea-models
export MODELNAME="bigscience/bloom-560m"
helm install tgi tgi --set global.modelUseHostPath=${MODELDIR} --set LLM_MODEL_ID=${MODELNAME}
# To deploy on Gaudi enabled kubernetes cluster
# helm install tgi tgi --set global.modelUseHostPath=${MODELDIR} --set LLM_MODEL_ID=${MODELNAME} --values gaudi-values.yaml
```

By default, the tgi service will downloading the "bigscience/bloom-560m" which is about 1.1GB.

If you already cached the model locally, you can pass it to container like this example:

MODELDIR=/home/ubuntu/hfmodels

MODELNAME="/data/models--bigscience--bloom-560m"

## Values

| Key                     | Type   | Default                                           | Description                                                                                                                                                  |
| ----------------------- | ------ | ------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| LLM_MODEL_ID            | string | `"bigscience/bloom-560m"`                         | Models id from https://huggingface.co/, or predownloaded model directory                                                                                     |
| port                    | string | `2080`                                            | Hugging Face Text Generation Inference service port                                                                                                          |
| global.modelUseHostPath | string | `"/mnt/opea-models"`                              | Cached models directory, tgi will not download if the model is cached here. The host path "modelUseHostPath" will be mounted to container as /data directory |
| image.repository        | string | `"ghcr.io/huggingface/text-generation-inference"` |                                                                                                                                                              |
| image.tag               | string | `"1.4"`                                           |                                                                                                                                                              |
