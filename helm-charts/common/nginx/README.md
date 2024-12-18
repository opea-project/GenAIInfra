# OPEA nginx microservice

Helm chart for deploying OPEA nginx service.

## Using nginx chart

This chart is not meant to be used standalone. It should be used with OPEA E2E helm charts as the reverse proxy for UI.

1. In the OPEA E2E chart, define a ConfigMap containing the environment variables as defined in [OPEA nginx microservice](https://github.com/opea-project/GenAIComps/blob/main/comps/nginx/nginx.conf.template).

2. Pass the name of the ConfigMap to helm configuration `cmName`.
