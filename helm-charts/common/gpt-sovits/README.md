# gpt-sovits

Helm chart for deploying gpt-sovits microservice.

## Install the chart

```console
cd GenAIInfra/helm-charts/common/
export MODELDIR=/mnt/opea-models
helm install gpt-sovits gpt-sovits --set global.modelUseHostPath=${MODELDIR}
```

The gpt-sovits service will download model `lj1995/GPT-SoVITS` which is about 2.8GB.

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are running.

Then run the command `kubectl port-forward svc/gpt-sovits 9880:9880` to expose the gpt-sovits service for access.

Open another terminal and run the following command to verify the service if working:

- Chinese only

```bash
curl localhost:9880/ -XPOST -d '{
    "text": "先帝创业未半而中道崩殂，今天下三分，益州疲弊，此诚危急存亡之秋也。",
    "text_language": "zh"
}' --output out.wav
```

- English only

```bash
curl localhost:9880/ -XPOST -d '{
    "text": "Discuss the evolution of text-to-speech (TTS) technology from its early beginnings to the present day. Highlight the advancements in natural language processing that have contributed to more realistic and human-like speech synthesis. Also, explore the various applications of TTS in education, accessibility, and customer service, and predict future trends in this field. Write a comprehensive overview of text-to-speech (TTS) technology.",
    "text_language": "en"
}' --output out.wav
```

## Values

| Key                             | Type   | Default                              | Description                                                                                                                                                                                                                                                                        |
| ------------------------------- | ------ | ------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| image.repository                | string | `"opea/gpt-sovits"`                  |                                                                                                                                                                                                                                                                                    |
| service.port                    | string | `"9880"`                             |                                                                                                                                                                                                                                                                                    |
| global.HUGGINGFACEHUB_API_TOKEN | string | `insert-your-huggingface-token-here` | Hugging Face API token                                                                                                                                                                                                                                                             |
| global.modelUseHostPath         | string | `""`                                 | Cached models directory, service will not download if the model is cached here. The host path "modelUseHostPath" will be mounted to the container and the downloaded model will be saved to directory `lj1995/GPT-SoVITS`. Set this to null/empty will force it to download model. |
