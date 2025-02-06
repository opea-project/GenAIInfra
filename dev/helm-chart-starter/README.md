# OPEA <CHARTNAME> microservice

Helm chart for deploying OPEA example service.

## Installing the Chart

To install the chart, run the following:

```console
cd GenAIInfra/helm-charts/common
export HFTOKEN="insert-your-huggingface-token-here"
# To deploy <CHARTNAME> microserice on CPU
helm install <CHARTNAME> <CHARTNAME> --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN}

```

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are running and in ready state.

Then run the command `kubectl port-forward svc/<CHARTNAME> 8080:8080` to expose the corresponding K8s service for access.

Open another terminal and run the following command to verify the service if working:

```
curl http://localhost:8080/v1/<CHARTNAME> \
     -X POST \
     -H 'Content-Type: application/json' \
     -d '<service specific content>'

```

## Values

| Key                             | Type   | Default                              | Description           |
| ------------------------------- | ------ | ------------------------------------ | --------------------- |
| global.HUGGINGFACEHUB_API_TOKEN | string | `insert-your-huggingface-token-here` | HuggingFace API token |
