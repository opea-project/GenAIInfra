# gpt-sovits

Helm chart for deploying gpt-sovits microservice.

## Install the chart

```console
cd GenAIInfra/helm-charts/common/
export MODELDIR=/mnt/opea-models
helm install gpt-sovits gpt-sovits --set global.modelUseHostPath=${MODELDIR}
```

The gpt-sovits service will download model `lj1995/GPT-SoVITS` which is about 2.8GB.

### Install the microservice in air gapped (offline) mode

To run `gpt-sovits` microservice in an air gapped environment, users are required to pre-download the model `lj1995/GPT-SoVITS` to a shared storage.

Below is an example for using node level local directory to download the model data:

Assuming the model data is shared using node-local directory `/mnt/opea-models`.

```
# On every K8s node, run the following command:
export MODEL_DIR=/mnt/opea-models
# Download model, assumes Python huggingface_hub[cli] module is already installed
huggingface-cli download --local-dir-use-symlinks False --local-dir "${MODEL_DIR}/lj1995/GPT-SoVITS" lj1995/GPT-SoVITS
# On K8s master node, run the following command:
# Install using Helm with the following additional parameters:
helm install ... ... --set global.offline=true,global.modelUseHostPath=${MODEL_DIR}
```

Assuming we share the offline data on cluster level using a persistent volume (PV), first we need to create the persistent volume claim (PVC) with name `opea-model-pvc` to store the model data.

```
# Download model openai/whisper-small at the root directory of the corresponding PV
# ... ...
# Install using Helm with the following additional parameters:
# export MODEL_PVC=opea-model-pvc
# helm install ... ... --set global.offline=true,global.modelUsePVC=${MODEL_PVC}
```

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

| Key                     | Type   | Default                              | Description                                                                                                                                                                                                                                                                                                                                 |
| ----------------------- | ------ | ------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| image.repository        | string | `"opea/gpt-sovits"`                  |                                                                                                                                                                                                                                                                                                                                             |
| service.port            | string | `"9880"`                             |                                                                                                                                                                                                                                                                                                                                             |
| global.HF_TOKEN         | string | `insert-your-huggingface-token-here` | Hugging Face API token                                                                                                                                                                                                                                                                                                                      |
| global.offline          | bool   | `false`                              | Whether to run the microservice in air gapped environment                                                                                                                                                                                                                                                                                   |
| global.modelUseHostPath | string | `""`                                 | Cached models directory on Kubernetes node, service will not download if the model is cached here. The host path "modelUseHostPath" will be mounted to the container as /data directory. Setting this to null/empty will force the pod to download the model every time during startup. May not be set if `global.modelUsePVC` is also set. |
| global.modelUsePVC      | string | `""`                                 | Name of Persistent Volume Claim to use for model cache. The Persistent Volume will be mounted to the container as /data directory. Setting this to null/empty will force the pod to download the model every time during startup. May not be set if `global.modelUseHostPath` is also set.                                                  |
