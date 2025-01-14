# lvm-uservice

Helm chart for deploying LVM microservice.

## Installing the chart

lvm-uservice depends on one of the following backend services:

- TGI: please refer to [tgi](../tgi) chart for more information

- one of the large vision model inference engine: please refer to [lvm-serve](../lvm-serve) chart for more information

First, you need to get the dependent service deployed, i.e. deploy the tgi helm chart, or lvm helm chart.

After you've deployed the dependent service successfully, please run `kubectl get svc` to get the backend service URL, e.g. `http://tgi`, `http://lvm-serve`.

To install the `lvm-uservice` chart, run the following:

```console
cd GenAIInfra/helm-charts/common/lvm-uservice
helm dependency update
export HFTOKEN="insert-your-huggingface-token-here"

# Use TGI as the backend
export LVM_BACKEND="TGI"
export LVM_ENDPOINT="http://tgi"
helm install lvm-uservice . --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set LVM_BACKEND=${LVM_BACKEND} --set LVM_ENDPOINT=${LVM_ENDPOINT} --wait

# Use other lvm-serve engine variant as the backend, see file `values.yaml` more details
# export LVM_ENDPOINT="http://lvm-serve"
# export LVM_BACKEND="LLaVA"
# helm install lvm-uservice . --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set LVM_BACKEND=${LVM_BACKEND} --set LVM_ENDPOINT=${LVM_ENDPOINT} --wait
```

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are running.

Then run the command `kubectl port-forward svc/lvm-uservice 9399:9399` to expose the lvm-uservice service for access.

Open another terminal and run the following command to verify the service if working:

```console
curl http://localhost:9000/v1/lvm \
    -X POST \
    -d '{"image": "iVBORw0KGgoAAAANSUhEUgAAAAoAAAAKCAYAAACNMs+9AAAAFUlEQVR42mP8/5+hnoEIwDiqkL4KAcT9GO0U4BxoAAAAAElFTkSuQmCC", "prompt":"What is this?"}' \
    -H 'Content-Type: application/json'
```

## Values

| Key                             | Type   | Default | Description                                                                                       |
| ------------------------------- | ------ | ------- | ------------------------------------------------------------------------------------------------- |
| global.HUGGINGFACEHUB_API_TOKEN | string | `""`    | Your own Hugging Face API token                                                                   |
| LVM_BACKEND                     | string | `"TGI"` | lvm backend engine, possible value "TGI", "LLaVA", "VideoLlama", "LlamaVision", "PredictionGuard" |
| LVM_ENDPOINT                    | string | `""`    | LVM endpoint                                                                                      |
| global.monitoring               | bool   | `false` | Service usage metrics                                                                             |
