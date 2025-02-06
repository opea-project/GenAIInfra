# Development

This directory contains various components for developing with GenAIInfra:

## Helm Starter Chart

Use [helm starter chart](helm-chart-starter) to create helm chart for new microservices for OPEA, e.g:

```console
helm create <mychart> -p </absolute/path/to/helm-chart-starter>
```

## NOTE

1. The `xxx-values.yaml` in the helm chart will trigger helm chart CI. Please make sure it has proper configurations.
