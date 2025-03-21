# AudioQnA

Helm chart for deploying AudioQnA service.

AudioQnA depends on the following micro services:

- [asr](../common/asr/README.md)
- [whisper](../common/whisper/README.md)
- [llm-uservice](../common/llm-uservice/README.md)
- [tgi](../common/tgi/README.md) or [vllm](../common/vllm/README.md)
- [tts](../common/tts/README.md) or [gpt-sovits](../common/gpt-sovits/README.md)
- [speecht5](../common/speecht5/README.md)

## Installing the Chart

To install the chart, run the following:

```console
cd GenAIInfra/helm-charts/
./update_dependency.sh
helm dependency update audioqna
export HFTOKEN="insert-your-huggingface-token-here"
# To use CPU with vLLM
helm install audioqna audioqna --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} -f audioqna/cpu-values.yaml
# To use CPU with TGI
# helm install audioqna audioqna --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} -f audioqna/cpu-tgi-values.yaml
# To use CPU with vLLM with multilang tts
# helm install audioqna audioqna --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} -f cpu-multilang-values.yaml
# To use Gaudi device with vLLM
# helm install audioqna audioqna --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} -f audioqna/gaudi-values.yaml
# To use Gaudi device with TGI
# helm install audioqna audioqna --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} -f audioqna/gaudi-tgi-values.yaml
```

### IMPORTANT NOTE

1. If you want to cache the downloaded model for later reuse, please set the bash environment variable `MODELDIR` to an existing directory on the node, then append `--set global.modelUseHostPath=${MODELDIR}` to the `helm install` commands.

2. Make sure your `${MODELDIR}` and `${MODELDIR}/.locks` is writable to all the users if you want to use the cached downloaded models, i.e. `$ sudo chmod 0777 ${MODELDIR} && sudo chmod 0777 ${MODELDIR}/.locks `.

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are running.

Test the AudioQnA megaservice by recording a .wav file, encoding the file into the base64 format, and then sending the base64 string to the megaservice endpoint. The megaservice will return a spoken response as a base64 string. To listen to the response, decode the base64 string and save it as a .wav file.

### Verify the workload through curl command

Then run the command `kubectl port-forward svc/audioqna 3008:3008` to expose the service for access.

Open another terminal and run the following command to verify the service if working:

```console
curl http://localhost:3008/v1/audioqna \
  -X POST \
  -d '{"audio": "UklGRigAAABXQVZFZm10IBIAAAABAAEARKwAAIhYAQACABAAAABkYXRhAgAAAAEA", "max_tokens":64}' \
  -H 'Content-Type: application/json' | sed 's/^"//;s/"$//' | base64 -d > output.wav
```

## Values

| Key               | Type   | Default                     | Description                                                                            |
| ----------------- | ------ | --------------------------- | -------------------------------------------------------------------------------------- |
| image.repository  | string | `"opea/audioqna"`           |                                                                                        |
| service.port      | string | `"3008"`                    |                                                                                        |
| tgi.LLM_MODEL_ID  | string | `Intel/neural-chat-7b-v3-3` | Models id from https://huggingface.co/, or predownloaded model directory               |
| global.monitoring | bool   | `false`                     | Enable usage metrics for the service components. See ../monitoring.md before enabling! |
