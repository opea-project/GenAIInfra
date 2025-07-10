# whisper

Helm chart for deploying whisper service.

## Installing the Chart

To install the chart, run the following:

```console
export MODELDIR=/mnt/opea-models
export ASR_MODEL_PATH="openai/whisper-small"
helm install whisper whisper --set global.modelUseHostPath=${MODELDIR} --set ASR_MODEL_PATH=${ASR_MODEL_PATH}
```

### Install the microservice in air gapped (offline) mode

To run `whisper` microservice in an air gapped environment, users are required to pre-download the model `openai/whisper-small` to a shared storage.

#### Use node-local directory

Assuming the model data is shared using node-local directory `/mnt/opea-models`.

```
# On every K8s node, run the following command:
export MODEL_DIR=/mnt/opea-models
# Download model, assumes Python huggingface_hub[cli] module is already installed
huggingface-cli download --cache-dir "${MODEL_DIR}" openai/whisper-small

# On K8s master node, run the following command:
# Install using Helm with the following additional parameters:
helm install ... --set global.offline=true,global.modelUseHostPath=${MODEL_DIR}
```

#### Use persistent volume

Assuming we share the offline data on cluster level using a persistent volume (PV), first we need to create the persistent volume claim (PVC) with name `opea-model-pvc` to store the model data.

```
# Download model openai/whisper-small at the root directory of the corresponding PV
# ...
# Install using Helm with the following additional parameters:
# export MODEL_PVC=opea-model-pvc
# helm install ... --set global.offline=true,global.modelUsePVC=${MODEL_PVC}
```

## Verify

Use port-forward to access it from localhost.

```console
kubectl port-forward service/whisper 1234:7066 &
curl http://localhost:1234/v1/asr \
  -XPOST \
  -d '{"audio": "UklGRigAAABXQVZFZm10IBIAAAABAAEARKwAAIhYAQACABAAAABkYXRhAgAAAAEA"}' \
  -H 'Content-Type: application/json'
```

## Values

| Key                     | Type   | Default                              | Description                                                                                                                                                                                                                                                                                                                                 |
| ----------------------- | ------ | ------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| image.repository        | string | `"opea/whisper"`                     |                                                                                                                                                                                                                                                                                                                                             |
| service.port            | string | `"7066"`                             |                                                                                                                                                                                                                                                                                                                                             |
| global.offline          | bool   | `false`                              | Whether to run the microservice in air gapped environment                                                                                                                                                                                                                                                                                   |
| global.modelUseHostPath | String | `""`                                 | Cached models directory on Kubernetes node, service will not download if the model is cached here. The host path "modelUseHostPath" will be mounted to the container as /data directory. Setting this to null/empty will force the pod to download the model every time during startup. May not be set if `global.modelUsePVC` is also set. |
| global.modelUsePVC      | String | `""`                                 | Name of Persistent Volume Claim to use for model cache. The Persistent Volume will be mounted to the container as /data directory. Setting this to null/empty will force the pod to download the model every time during startup. May not be set if `global.modelUseHostPath` is also set.                                                  |
| global.HF_TOKEN         | string | `insert-your-huggingface-token-here` | Hugging Face API token                                                                                                                                                                                                                                                                                                                      |
