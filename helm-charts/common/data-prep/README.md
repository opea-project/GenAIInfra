# data-prep

Helm chart for deploying Retriever microservice.

data-prep depends on redis, refer to redis-vector-db for more config details.

## Installing the Chart

To install the chart, run the following:

```console
$ helm install dataprep data-prep
```

## Values

| Key              | Type   | Default                       | Description |
| ---------------- | ------ | ----------------------------- | ----------- |
| image.repository | string | `"opea/retriever-tgi:latest"` |             |
| service.port     | string | `"6007"`                      |             |
