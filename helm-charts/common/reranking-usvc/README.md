# Re-ranking Microservice (reranking-usvc)

**Helm chart for deploying reranking microservice.**

There are two versions of Reranking microservice. First one is based on TEI model and does reranking for textual data. An alternative multimodal version based on `vdms-values.yaml` file, performs reranking for visual data. Follow along to select and install the version which suites your use case.

reranking-usvc depends on teirerank. Set the TEI_RERANKING_ENDPOINT as teirerank endpoint. Multimodal version depends on data-prep and vdms-vector-db service. Endpoints for these dependencies should be set properly before installing the chart.

## Install Re-ranking Microservice based on TEI

### (Option1): Installing the chart separately

First, you need to install the teirerank chart, please refer to the [teirerank](../teirerank) chart for more information.

After you've deployted the teirerank chart successfully, please run `kubectl get svc` to get the tei service endpoint, i.e. `http://teirerank`.

To install the reranking-usvc chart, run the following:

```console
cd GenAIInfra/helm-charts/common/reranking-usvc
export TEI_RERANKING_ENDPOINT="http://teirerank"
helm dependency update
helm install reranking-usvc . --set TEI_RERANKING_ENDPOINT=${TEI_RERANKING_ENDPOINT}
```

### (Option2): Installing the chart with dependencies automatically

```console
cd GenAIInfra/helm-charts/common/reranking-usvc
helm dependency update
helm install reranking-usvc . --set teirerank.enabled=true
```

## Install Re-ranking Microservice for visual data

### (Option1): Installing the chart separately

First, you need to install the multimodal data-prep chart with vdms based values file and `vdms-vector-db` chart. `vdms-vector-db` is used by multimodal version of data-prep service. Please refer to the [data-prep](../data-prep) and [vdms-vector-db](../vdms-vector-db) charts guide for more information.

After you've deployed the data-prep chart successfully, please run `kubectl get svc` to get the data-prep service endpoint and port, i.e. `http://data-prep:6007`.

To install the reranking-usvc chart, run the following:

```bash
cd GenAIInfra/helm-charts/common/reranking-usvc

# Use the host and port returned in first step for setting the following environment variables.
export FILE_SERVER_ENDPOINT="http://data-prep:6007/v1/dataprep/get_file"
export GET_VIDEO_LIST_ENDPOINT="http://data-prep:6007/v1/dataprep/get_videos"

# Export the proxy variables if you are behind a proxy. Assign empty string if no proxy setup required.
export https_proxy="your_http_proxy"
export http_proxy="your_https_proxy"

helm dependency update
helm install reranking-usvc . -f ./vdms-values.yaml --set fileServerEndpoint=${FILE_SERVER_ENDPOINT} --set getVideoListEndpoint=${GET_VIDEO_LIST_ENDPOINT} --set global.https_proxy=${https_proxy} --set global.http_proxy=${http_proxy}
```

### (Option2): Installing the chart with dependencies automatically

Multimodal Reranking microservice depends on Multimodal Data-Prep Microservice. Multimodal Data-Prep microservice in turn depends on vdms-vector-db microservice.

```bash
cd GenAIInfra/helm-charts/common/reranking-usvc

# Export the proxy variables. Assign empty string if no proxy setup required.
export https_proxy="your_http_proxy"
export http_proxy="your_https_proxy"

helm dependency update
helm install reranking-usvc . -f ./vdms-values.yaml --set data-prep.enabled=true --set data-prep.vdms-vector-db.enabled=true --set global.https_proxy=${https_proxy} --set global.http_proxy=${http_proxy}
```

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are running.

Then run the command `kubectl port-forward svc/reranking-usvc 8000:8000` to expose the reranking-usvc service for access.

Open another terminal and run the following command to verify the service if working:

### 1. For Reranking Service based on TEI

````bash

curl http://localhost:8000/v1/reranking \
    -X POST \
    -d '{"initial_query":"What is Deep Learning?", "retrieved_docs": [{"text":"Deep Learning is not..."}, {"text":"Deep learning is..."}]}' \
    -```H 'Content-Type: application/json'
````

### 2. For visual data Reranking service

```bash
curl http://localhost:8000/v1/reranking \
  -X 'POST' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
    "retrieved_docs": [{"doc": [{"text": "this is the retrieved text"}]}],
    "initial_query": "this is the query",
    "top_n": 1,
    "metadata": [
        {"other_key": "value", "video":"top_video_name", "timestamp":"20"}
    ]
  }'
```

## Values

| Key                              | Type   | Default                     | Description |
| -------------------------------- | ------ | --------------------------- | ----------- |
| image.repository                 | string | `"opea/reranking-tgi"`      |             |
| TEI_RERANKING_ENDPOINT           | string | `""`                        |             |
| service.port                     | string | `"8000"`                    |
| vdms-values:image.repository     | string | `"opea/reranking-videoqna"` |             |
| vdms-values:fileServerEndpoint   | string | `""`                        |             |
| vdms-values:getVideoListEndpoint | string | `""`                        |             |
