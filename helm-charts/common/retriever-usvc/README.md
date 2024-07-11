# retriever-usvc

Helm chart for deploying Retriever microservice.

retriever-usvc depends on redis and tei, you should set these endpoints before start.

## Installing the Chart

To install the chart, run the following:

```console
export REDIS_URL="redis://redis-vector-db:6379"
export TEI_EMBEDDING_ENDPOINT="http://tei"
helm install retriever retriever-usvc --set REDIS_URL=${REDIS_URL} --set TEI_EMBEDDING_ENDPOINT=${TEI_EMBEDDING_ENDPOINT}
```

## Values

| Key                    | Type   | Default                       | Description |
| ---------------------- | ------ | ----------------------------- | ----------- |
| image.repository       | string | `"opea/retriever-tgi:latest"` |             |
| service.port           | string | `"7000"`                      |             |
| REDIS_URL              | string | `""`                          |             |
| TEI_EMBEDDING_ENDPOINT | string | `""`                          |             |
