# embedding-usvc

Helm chart for deploying embedding microservice.

embedding-usvc depends on TEI, set TEI_EMBEDDING_ENDPOINT.

## Installing the Chart

To install the chart, run the following:

```console
$ export TEI_EMBEDDING_ENDPOINT="http://tei"
$ helm install embedding embedding-usvc --set TEI_EMBEDDING_ENDPOINT=${TEI_EMBEDDING_ENDPOINT}
```

## Values

| Key                    | Type   | Default                       | Description |
| ---------------------- | ------ | ----------------------------- | ----------- |
| image.repository       | string | `"opea/embedding-tei:latest"` |             |
| service.port           | string | `"6000"`                      |             |
| TEI_EMBEDDING_ENDPOINT | string | `""`                          |             |
