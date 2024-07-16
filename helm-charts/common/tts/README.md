# tts

Helm chart for deploying tts microservice.

tts depends on speecht5, you should set TTS_ENDPOINT endpoints before start.

## Installing the Chart

To install the chart, run the following:

```console
export TTS_ENDPOINT="http://speecht5:7055"
helm install tts tts --set TTS_ENDPOINT=${TTS_ENDPOINT}
```

## Verify

Use port-forward to access it from localhost.

```console
kubectl port-forward service/tts 1234:9088 &
curl http://localhost:1234/v1/audio/speech \
  -XPOST \
  -d '{"text": "Who are you?"}' \
  -H 'Content-Type: application/json'
```

## Values

| Key              | Type   | Default             | Description |
| ---------------- | ------ | ------------------- | ----------- |
| image.repository | string | `"opea/tts:latest"` |             |
| service.port     | string | `"9088"`            |             |
| TTS_ENDPOINT     | string | `""`                |             |
