# CodeTrans

Helm chart for deploying CodeTrans service.

CodeTrans depends on LLM microservice, refer to llm-uservice for more config details.

## Installing the Chart

To install the chart, run the following:

```console
cd GenAIInfra/helm-charts/
./update_dependency.sh
helm dependency update codetrans
export HFTOKEN="insert-your-huggingface-token-here"
export MODELDIR="/mnt/opea-models"
export MODELNAME="HuggingFaceH4/mistral-7b-grok"
helm install codetrans codetrans --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} --set tgi.LLM_MODEL_ID=${MODELNAME}
# To use Gaudi device
# helm install codetrans codetrans --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --values codetrans/gaudi-values.yaml
```

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are running.

Then run the command `kubectl port-forward svc/codetrans 7777:7777` to expose the CodeTrans service for access.

Open another terminal and run the following command to verify the service if working:

```console
curl http://localhost:7777/v1/codetrans \
    -H 'Content-Type: application/json' \
    -d '{"language_from": "Golang","language_to": "Python","source_code": "package main\n\nimport \"fmt\"\nfunc main() {\n    fmt.Println(\"Hello, World!\");\n}"}'
```

## Values

| Key                             | Type   | Default                           | Description                                                                                                                                                  |
| ------------------------------- | ------ | --------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| image.repository                | string | `"opea/codetrans:latest"`         |                                                                                                                                                              |
| service.port                    | string | `"7777"`                          |                                                                                                                                                              |
| global.HUGGINGFACEHUB_API_TOKEN | string | `""`                              | Your own Hugging Face API token                                                                                                                              |
| global.modelUseHostPath         | string | `"/mnt/opea-models"`              | Cached models directory, tgi will not download if the model is cached here. The host path "modelUseHostPath" will be mounted to container as /data directory |
| tgi.LLM_MODEL_ID                | string | `"HuggingFaceH4/mistral-7b-grok"` | Models id from https://huggingface.co/, or predownloaded model directory                                                                                     |
