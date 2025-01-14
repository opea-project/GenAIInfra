# OPEA lvm-serve microservice

Helm chart for deploying OPEA large vision model service.

## Installing the Chart

To install the chart, run the following:

```console
cd GenAIInfra/helm-charts/common
export MODELDIR=/mnt/opea-models
export HFTOKEN="insert-your-huggingface-token-here"
export LVM_MODEL_ID="llava-hf/llava-1.5-7b-hf"
# To deploy lvm-llava microserice on CPU
helm install lvm-serve lvm-serve --set global.modelUseHostPath=${MODELDIR} --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set LVM_MODEL_ID=${LVM_MODEL_ID}
# To deploy lvm-llava  microserice on Gaudi
# helm install lvm-serve lvm-serve --set global.modelUseHostPath=${MODELDIR} --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set LVM_MODEL_ID=${LVM_MODEL_ID} --values lvm-serve/gaudi-values.yaml
# To deploy lvm-video-llama microserice on CPU
helm install lvm-serve lvm-serve --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --values lvm-serve/variant_video-llama-values.yaml
```

By default, the lvm-serve-llava service will downloading the model "llava-hf/llava-1.5-7b-hf" which is about 14GB.

If you already cached the model locally, you can pass it to container like this example:

MODELDIR=/mnt/opea-models

MODELNAME="/data/models--llava-hf--llava-1.5-7b-hf"

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are runinng and in ready state.

Then run the command `kubectl port-forward svc/lvm-serve 9399:9399` to expose the lvm-serve service for access.

Open another terminal and run the following command to verify the service if working:

```console
# Verify with lvm-llava
pip install Pillow requests
image_b64_str=$(python -c 'import base64;from io import BytesIO;import PIL.Image;import requests;image_path = "https://avatars.githubusercontent.com/u/39623753?s=40&v=4";image = PIL.Image.open(requests.get(image_path, stream=True, timeout=3000).raw);buffered = BytesIO();image.save(buffered, format="PNG");img_b64_str = base64.b64encode(buffered.getvalue()).decode();print(img_b64_str)')
body="{\"img_b64_str\": \"${image_b64_str}\", \"prompt\": \"What is this?\", \"max_new_tokens\": 32}"
url="http://localhost:9399/generate"
curl $url -XPOST -d "$body" -H 'Content-Type: application/json'

# Verify with lvm-video-llama
body='{"image": "iVBORw0KGgoAAAANSUhEUgAAAAoAAAAKCAYAAACNMs+9AAAAFUlEQVR42mP8/5+hnoEIwDiqkL4KAcT9GO0U4BxoAAAAAElFTkSuQmCC", "prompt": "Describe the image.", "max_new_tokens": 32}'
url="http://localhost:9399/v1/lvm-serve"
curl $url -XPOST -d "$body" -H 'Content-Type: application/json'
```

## Values

| Key                             | Type   | Default                              | Description                                                                                                                                                                                                               |
| ------------------------------- | ------ | ------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| global.HUGGINGFACEHUB_API_TOKEN | string | `insert-your-huggingface-token-here` | Hugging Face API token                                                                                                                                                                                                    |
| global.modelUseHostPath         | string | `""`                                 | Cached models directory, service will not download if the model is cached here. The host path "modelUseHostPath" will be mounted to container as /data directory. Set this to null/empty will force it to download model. |
| LVM_MODEL_ID                    | string | `"llava-hf/llava-1.5-7b-hf"`         |                                                                                                                                                                                                                           |
| autoscaling.enabled             | bool   | `false`                              | Enable HPA autoscaling for the service deployment based on metrics it provides. See [HPA instructions](../../HPA.md) before enabling!                                                                                     |
| global.monitoring               | bool   | `false`                              | Enable usage metrics for the service. Required for HPA. See [monitoring instructions](../../monitoring.md) before enabling!                                                                                               |
