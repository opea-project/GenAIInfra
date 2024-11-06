# LVM Serving Microservice (video-llama-lvm)

Helm chart for deploying video-llama-lvm microservice.

`video-llama-lvm` is a microservice which provides inference from Video-Llama-2-7b.

## Installing the Chart

To install the chart, run the following:

```bash
cd GenAIInfra/helm-charts/common
export HFTOKEN="insert-your-huggingface-token-here"

# Set a dir to cache downloaded Models and clips
export MODELDIR=/mnt/opea-models

# When setting up for first time, model needs to be downloaded. Set LLM_DOWNLOAD flag to true to download models. Please note, when redeploying we should set this value to false, otherwise model download will restart.
export LLM_DOWNLOAD=true

# Export the proxy variables. Assign empty string if no proxy setup required.
export https_proxy="your_http_proxy"
export http_proxy="your_https_proxy"

helm install video-llama-lvm video-llama-lvm --set global.modelUseHostPath=${MODELDIR} --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set llmDownload=${LLM_DOWNLOAD} --set global.https_proxy=${https_proxy} --set global.http_proxy=${http_proxy}
```

> \***\*NOTE:\*\*** **Model download may take up to 1.5 Hours.** When installing chart for the first time we should set `llmDownload` value to be **true**. This helps download model for the first run. Afterwards, for re-installing charts set `llmDownload` value in the chart to be **false**. If not set to **false**, model download will again start and service will again long time to be ready.

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are running.

Then run the command `kubectl port-forward svc/video-llama-lvm 2080:80` to expose the video-llama-lvm service for access on the host machine.

Open another terminal and run the following command to verify the service if working:

```bash
curl -X POST \
  "http://localhost:2080/generate?video_url=silence_girl.mp4&start=0.0&duration=9&prompt=What%20is%20the%20person%20doing%3F&max_new_tokens=150" \
  -H "accept: */*" \
  -d ''
```

## Values

| Key                             | Type    | Default                              | Description                                                                                                                              |
| ------------------------------- | ------- | ------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------- |
| global.cacheUseHostPath         | string  | "/home/$USER/.cache"                 | Cache the embedding model and related data.                                                                                              |
| global.HUGGINGFACEHUB_API_TOKEN | string  | `insert-your-huggingface-token-here` | Hugging Face API token                                                                                                                   |
| global.modelUseHostPath         | string  | `"/mnt/opea-models"`                 | Cached models directory. The host path "modelUseHostPath" will be mounted to container as /home/user/model directory.                    |
| image.repository                | string  | `"opea/video-llama-lvm-server"`      |                                                                                                                                          |
| image.tag                       | string  | `"latest"`                           |                                                                                                                                          |
| llmDownload                     | boolean | `true`                               | This value when true, makes video-llama-lvm download a model. Change it to false for stopping video-llama-lvm from re-downloading model. |
