# Embedding Microservice (embedding-usvc)

**Helm chart for deploying Embedding Microservice.**

Embedding microservice is consumed by several reference applications present in [GenAIExample](https://github.com/opea-project/GenAIExamples/tree/main).

There are 2 versions of embedding microservice. First version is unimodal based on TEI creating embeddings for textual data. An alternative version based on `vdms-values.yaml` file, uses multimodal embedding models for creating embedding for visual data. Follow along to select and install the version which suites your use case.

embedding-usvc depends on TEI, and TEI_EMBEDDING_ENDPOINT should be set properly. Multimodal version has no dependencies.

## Install Embedding microservice chart based on TEI

### (Option1): Installing the chart separately

First, you need to install the tei chart, please refer to the [tei](../tei) chart for more information.

After you've deployted the tei chart successfully, please run `kubectl get svc` to get the tei service endpoint, i.e. `http://tei`.

To install the embedding-usvc chart, run the following:

```console
cd GenAIInfra/helm-charts/common/embedding-usvc
export TEI_EMBEDDING_ENDPOINT="http://tei"
helm dependency update
helm install embedding-usvc . --set TEI_EMBEDDING_ENDPOINT=${TEI_EMBEDDING_ENDPOINT}
```

### (Option2): Installing the chart with dependencies automatically

```console
cd GenAIInfra/helm-charts/common/embedding-usvc
helm dependency update
helm install embedding-usvc . --set tei.enabled=true
```

## Install Multimodal Embedding Microservice chart

To install the multimodal embedding-usvc chart, run the following:

```bash
cd GenAIInfra/helm-charts/common/embedding-usvc

# Set a directory to cache emdedding models
export CACHEDIR="/home/$USER/.cache"

# Export the proxy variables. Assign empty string if no proxy setup required.
export https_proxy="your_http_proxy"
export http_proxy="your_https_proxy"

helm dependency update
helm install embedding-usvc . -f ./vdms-values.yaml --set global.cacheUseHostPath=${CACHEDIR} --set global.https_proxy=${https_proxy} --set global.http_proxy=${http_proxy}
```

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are running.

Then run the command `kubectl port-forward svc/embedding-usvc 6000:6000` to expose the embedding-usvc service for access.

Open another terminal and run the following command to verify the service if working:

```bash
curl http://localhost:6000/v1/embeddings \
    -X POST \
    -d '{"text":"hello"}' \
    -H 'Content-Type: application/json'
```

## Values

| Key                          | Type   | Default                            | Description |
| ---------------------------- | ------ | ---------------------------------- | ----------- |
| image.repository             | string | `"opea/embedding-tei"`             |             |
| service.port                 | string | `"6000"`                           |             |
| TEI_EMBEDDING_ENDPOINT       | string | `""`                               |             |
| global.monitoring            | bop;   | `false`                            |             |
| vdms-values:image.repository | string | `"opea/embedding-multimodal-clip"` |             |
