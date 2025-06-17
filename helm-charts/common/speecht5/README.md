# speecht5

Helm chart for deploying speecht5 service.

## Installing the Chart

To install the chart, run the following:

```console
export MODELDIR=/mnt/opea-models
helm install speecht5 speecht5 --set global.modelUseHostPath=${MODELDIR}
```

### Install the microservice in air gapped (offline) mode

To run `speecht5` microservice in an air gapped environment, users are required to pre-download the models `microsoft/speecht5_tts` and `microsoft/speecht5_hifigan` to a shared storage.

Below is an example for using node level local directory to download the model data:

Assuming the model data is shared using node-local directory `/mnt/opea-models`.

```
# On every K8s node, run the following command:
export MODEL_DIR=/mnt/opea-models
# Download model, assumes Python huggingface_hub[cli] module is already installed
huggingface-cli download --cache-dir "${MODEL_DIR}" microsoft/speecht5_tts
huggingface-cli download --cache-dir "${MODEL_DIR}" microsoft/speecht5_hifigan
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

Use port-forward to access it from localhost.

```console
kubectl port-forward service/speecht5 1234:7055 &
curl http://localhost:1234/v1/tts \
  -XPOST \
  -d '{"text": "Who are you?"}' \
  -H 'Content-Type: application/json'
```

## Values

| Key              | Type   | Default           | Description |
| ---------------- | ------ | ----------------- | ----------- |
| image.repository | string | `"opea/speecht5"` |             |
| service.port     | string | `"7055"`          |             |
