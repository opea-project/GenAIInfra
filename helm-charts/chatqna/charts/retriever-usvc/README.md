# retriever-usvc

Helm chart for deploying Retriever microservice.

retriever-usvc depends on redis, refer to redis-vector-db for more config details.

## Installing the Chart

To install the chart, run the following:

```console
$ helm install retriever retriever-usvc
```

## Values

| Key              | Type   | Default                       | Description |
| ---------------- | ------ | ----------------------------- | ----------- |
| image.repository | string | `"opea/retriever-tgi:latest"` |             |
| service.port     | string | `"7000"`                      |             |
