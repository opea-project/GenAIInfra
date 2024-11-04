# LVM Microservice

**Helm chart for deploying lvm-uservice microservice.**

There are two versions of `lvm-uservice`. First version runs with `tgi` service and another one runs with `lvm-serving` service. We will try to learn both setups in following sections.

## 1. Installing lvm-uservice to be used with tgi microservice

In this setup, lvm-uservice depends on TGI, you should set LVM_ENDPOINT as tgi endpoint.

### (Option1): Installing the chart separately

First, you need to install the tgi chart, please refer to the [tgi](../tgi) chart for more information.

After you've deployted the tgi chart successfully, please run `kubectl get svc` to get the tgi service endpoint, i.e. `http://tgi`.

To install the chart, run the following:

```bash
cd GenAIInfra/helm-charts/common/lvm-uservice
export HFTOKEN="insert-your-huggingface-token-here"
export LVM_ENDPOINT="http://tgi"
helm dependency update
helm install lvm-uservice . --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set LVM_ENDPOINT=${LVM_ENDPOINT} --wait
```

### (Option2): Installing the chart with dependencies automatically (with auto-installing tgi)

```bash
cd GenAIInfra/helm-charts/common/lvm-uservice
export HFTOKEN="insert-your-huggingface-token-here"
helm dependency update
helm install lvm-uservice . --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set tgi.enabled=true --wait
```

## 2. Installing lvm-uservice to be used with lvm-serving microservice (serving VideoLlama-7B)

This setup of `lvm-uservice` is utilized in some of the examples like [VideoQnA](https://github.com/opea-project/GenAIExamples/tree/main/VideoQnA). Here, `lvm-uservice` helps communicate to `lvm-serving` microservice. It facilitates sending queries and receiving response from `lvm-serving` microservice. Hence, it depends on lvm-serving microservice and you should make sure that `lvmEndpoint` value is set properly.

### (Option1): Installing the chart separately

First, you need to install the `lvm-serving` chart. Please refer to the [lvm-serving](../lvm-serving) chart for more information.

After you've deployed the `lvm-serving` chart successfully, please run `kubectl get svc` to get `lvm-serving` service host and port. The endpoint url for `lvm-serving` will be formed using the host and port. For example, default value would be `http://lvm-serving:80`.

To install the chart, run the following:

```bash
cd GenAIInfra/helm-charts/common/lvm-uservice
export LVM_ENDPOINT="http://lvm-serving:80"

# Export the proxy variables. Assign empty string if no proxy setup required.
export https_proxy="your_http_proxy"
export http_proxy="your_https_proxy"

helm dependency update
helm install lvm-uservice . -f ./variant_videoqna-values.yaml --set lvmEndpoint=${LVM_ENDPOINT} --set global.https_proxy=${https_proxy} --set global.http_proxy=${http_proxy} --wait
```

### (Option2): Installing the chart with dependencies automatically (with auto-installing lvm-serving dependency)

```bash
cd GenAIInfra/helm-charts/common/lvm-uservice

export HFTOKEN="insert-your-huggingface-token-here"
# Set a dir to cache downloaded Video-Llama Model
export MODELDIR=/mnt/opea-models
# Set a directory to cache emdedding models and other related data
export CACHEDIR="/home/$USER/.cache"
# When setting up for first time, model needs to be downloaded. Set LLM_DOWNLOAD flag to true to download models. Please note, when redeploying we should set this value to false, otherwise model download will restart.
export LLM_DOWNLOAD=true

# Set the proxy variables. Assign empty string if no proxy setup required.
export https_proxy="your_http_proxy"
export http_proxy="your_https_proxy"

helm dependency update
helm install lvm-uservice . -f ./variant_videoqna-values.yaml --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set lvm-serving.enabled=true --set lvm-serving.llmDownload=${LLM_DOWNLOAD} --set global.modelUseHostPath=${MODELDIR} --set global.cacheUseHostPath=${CACHEDIR} --set global.https_proxy=${https_proxy} --set global.http_proxy=${http_proxy} --wait
```

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are running.

### For TGI based lvm-uservice
Run the command `kubectl port-forward svc/lvm-uservice 9399:9399` to expose the lvm-uservice service for access.

### For lvm-serving based lvm-uservice
Run the command `kubectl port-forward svc/lvm-uservice 9000:9000` to expose the lvm-uservice service for access.


Open another terminal and run the following command to verify the service if working:

### Verify lvm-uservice running with lvm-serving (Video-Llama 7B) service

```bash
curl http://localhost:9000/v1/lvm \
  -X POST \
  -d '{"video_url":"https://github.com/DAMO-NLP-SG/Video-LLaMA/raw/main/examples/silence_girl.mp4","chunk_start": 0,"chunk_duration": 7,"prompt":"What is the person doing?","max_new_tokens": 50}' \
  -H 'Content-Type: application/json'
```

### Verify lvm-uservice running with TGI service

```bash
curl http://localhost:9399/v1/chat/completions \
    -X POST \
    -d '{"query":"What is Deep Learning?","max_tokens":17,"top_k":10,"top_p":0.95,"typical_p":0.95,"temperature":0.01,"repetition_penalty":1.03,"streaming":true}' \
    -H 'Content-Type: application/json'
```

## Values

| Key                             | Type   | Default          | Description                     |
| ------------------------------- | ------ | ---------------- | ------------------------------- |
| global.HUGGINGFACEHUB_API_TOKEN | string | `""`             | Your own Hugging Face API token |
| image.repository                | string | `"opea/lvm-tgi"` |                                 |
| videoqna: image.repository      | string | `"opea/lvm-video-llama"` |                         |
| service.port                    | string | `"9000"`         |                                 |
| LVM_ENDPOINT                    | string | `""`             | LVM endpoint                    |
| global.monitoring               | bop;   | false            | Service usage metrics           |
