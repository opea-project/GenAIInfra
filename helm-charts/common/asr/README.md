# asr

Helm chart for deploying asr microservice.

Web retriever depends on whisper, you should set ASR_ENDPOINT endpoints before start.

## Installing the Chart

To install the chart, run the following:

```console
export ASR_ENDPOINT="http://whisper:7066"
helm install asr asr --set ASR_ENDPOINT=${ASR_ENDPOINT}
```

## Verify

Use port-forward to access it from localhost.

```console
kubectl port-forward service/asr 1234:9099 &
curl http://localhost:1234/v1/audio/transcriptions \
  -XPOST \
  -d '{"byte_str": "UklGRigAAABXQVZFZm10IBIAAAABAAEARKwAAIhYAQACABAAAABkYXRhAgAAAAEA"}' \
  -H 'Content-Type: application/json'
```

## Values

| Key              | Type   | Default             | Description |
| ---------------- | ------ | ------------------- | ----------- |
| image.repository | string | `"opea/asr:latest"` |             |
| service.port     | string | `"9099"`            |             |
| ASR_ENDPOINT     | string | `""`                |             |
