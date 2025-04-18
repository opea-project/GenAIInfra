# OPEA <CHARTNAME> microservice

Helm chart for deploying OPEA example service.

## Installing the Chart

To install the chart, run the following:

```console
cd GenAIInfra/helm-charts/common
export HFTOKEN="insert-your-huggingface-token-here"
# To deploy nginx-central-gateway microserice on CPU
helm install nginx-central-gateway nginx-central-gateway

```

## Configuration

Update the `values.yaml` file to configure:

1. The UI service host and port
2. The backend services (chatqna, codegen, docsum, dataprep) hosts and ports

## Accessing Services

After deployment, you can access all services through the central gateway:

- UI: `http://<gateway-ip>/`
- ChatQnA: `http://<gateway-ip>/v1/chatqna`
- CodeGen: `http://<gateway-ip>/v1/codegen`
- DocSum: `http://<gateway-ip>/v1/docsum`
- DataPrep: `http://<gateway-ip>/v1/dataprep`

## Values

| Key                    | Type   | Default                                     | Description           |
| ---------------------- | ------ | ------------------------------------------- | --------------------- |
| service.type           | string | `LoadBalancer`                              | Service type          |
| service.port           | int    | `80`                                        | Service port          |
| ui.host                | string | `ui-service.ui-namespace.svc.cluster.local` | UI service host       |
| ui.port                | string | `5174`                                      | UI service port       |
| services.chatqna.host  | string | `chatqna.chatqna.svc.cluster.local`         | ChatQnA service host  |
| services.chatqna.port  | string | `8888`                                      | ChatQnA service port  |
| services.codegen.host  | string | `codegen.codegen.svc.cluster.local`         | CodeGen service host  |
| services.codegen.port  | string | `7778`                                      | CodeGen service port  |
| services.docsum.host   | string | `docsum.docsum.svc.cluster.local`           | DocSum service host   |
| services.docsum.port   | string | `6996`                                      | DocSum service port   |
| services.dataprep.host | string | `data-prep.chatqna.svc.cluster.local`       | DataPrep service host |
| services.dataprep.port | string | `6007`                                      | DataPrep service port |
