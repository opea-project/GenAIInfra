# data-prep

Helm chart for deploying data-prep microservice.

data-prep will use redis and tei service, please specify the endpoints.

## Installing the Chart

To install the chart, run the following:

```console
export REDIS_URL="redis://redis-vector-db:6379"
export TEI_EMBEDDING_ENDPOINT="http://tei"
helm install dataprep data-prep --set REDIS_URL=${REDIS_URL} --set TEI_EMBEDDING_ENDPOINT=${TEI_EMBEDDING_ENDPOINT}
```

## Values

| Key                    | Type   | Default                        | Description |
| ---------------------- | ------ | ------------------------------ | ----------- |
| image.repository       | string | `"opea/dataprep-redis:latest"` |             |
| service.port           | string | `"6007"`                       |             |
| REDIS_URL              | string | `""`                           |             |
| TEI_EMBEDDING_ENDPOINT | string | `""`                           |             |
