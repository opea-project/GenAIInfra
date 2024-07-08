# reranking-usvc

Helm chart for deploying reranking microservice.

reranking-usvc depends on TEI, set the TEI_RERANKING_ENDPOINT as teirerank endpoint.

## Installing the Chart

To install the chart, run the following:

```console
export TEI_RERANKING_ENDPOINT="http://teirerank"
helm install reranking reranking-usvc --set TEI_RERANKING_ENDPOINT=${TEI_RERANKING_ENDPOINT}
```

## Values

| Key                    | Type   | Default                       | Description |
| ---------------------- | ------ | ----------------------------- | ----------- |
| image.repository       | string | `"opea/reranking-tgi:latest"` |             |
| TEI_RERANKING_ENDPOINT | string | `""`                          |             |
| service.port           | string | `"8000"`                      |             |
