# embedding-usvc

Helm chart for deploying OPEA embedding microservice.

## Installing the chart

The OPEA embedding microservice depends on one of the following backend services:

- TEI: please refer to [tei](../tei) chart for more information

- multimodal embedding BridgeTower: please refer to [mm-embedding](../mm-embedding) chart for more information.

- prediction guard: please refert to external [Prediction Guard](https://predictionguard.com) for more information.

First, you need to get the dependent service deployed, i.e. deploy the `tei` helm chart, `mm-embedding` helm chart, or contact prediction guard to get access info.

After you've deployed the dependent service successfully, please run `kubectl get svc` to get the backend service URL, e.g. `http://tei`, `http://mm-embedding`.

To install the embedding-usvc chart, run the following:

```console
cd GenAIInfra/helm-charts/common/embedding-usvc
helm dependency update

# Use TEI as the backend(default)
export EMBEDDING_BACKEND="TEI"
export EMBEDDING_ENDPOINT="http://tei"
helm install embedding-usvc . --set EMBEDDING_BACKEND=${EMBEDDING_BACKEND} --set EMBEDDING_ENDPOINT=${EMBEDDING_ENDPOINT}

# Use multimodal embedding BridgeTower as the backend
# export EMBEDDING_BACKEND="BridgeTower"
# export EMBEDDING_ENDPOINT="http://mm-embedding"
# helm install embedding-usvc . --set EMBEDDING_BACKEND=${EMBEDDING_BACKEND} --set EMBEDDING_ENDPOINT=${EMBEDDING_ENDPOINT}

# Use predcition guard as the backend
# export EMBEDDING_BACKEND="PredictionGuard"
# export API_KEY=<your PedictionGuard api key>
# helm install embedding-usvc . --set EMBEDDING_BACKEND=${EMBEDDING_BACKEND} --set PREDICTIONGUARD_API_KEY=${API_KEY}

```

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are running.

Then run the command `kubectl port-forward svc/embedding-usvc 6000:6000` to expose the embedding-usvc service for access.

Open another terminal and run the following command to verify the service if working:

```console
# Verify with TEI or prediction guard backend:
curl http://localhost:6000/v1/embeddings \
    -X POST \
    -H 'Content-Type: application/json' \
    -d '{"input":"What is Deep Learning?"}'

# Verify with multimodal embedding BridgeTower backend:
curl http://localhost:6000/v1/embeddings \
    -X POST \
    -H 'Content-Type: application/json' \
    -d '{"text": {"text" : "This is some sample text."}, "image" : {"url": "https://github.com/docarray/docarray/blob/main/tests/toydata/image-data/apple.png?raw=true"}}'
```

## Values

| Key                | Type   | Default  | Description                                                           |
| ------------------ | ------ | -------- | --------------------------------------------------------------------- |
| service.port       | string | `"6000"` |                                                                       |
| EMBEDDING_BACKEND  | string | `"TEI"`  | backend engine to use, one of "TEI", "BridgeTower", "PredictionGuard" |
| EMBEDDING_ENDPOINT | string | `""`     |                                                                       |
| global.monitoring  | bool   | `false`  |                                                                       |
