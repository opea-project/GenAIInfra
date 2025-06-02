# OPEA nginx microservice

Helm chart for deploying OPEA nginx service.

## Integration with OPEA Applications

When used as a simple proxy within OPEA application charts (ChatQnA, CodeGen, etc.), parent charts should provide a ConfigMap named `{{ .Release.Name }}-nginx-config` containing environment variables for the [OPEA nginx microservice](https://github.com/opea-project/GenAIComps/blob/main/comps/nginx/nginx.conf.template).

## Gateway Mode (Central Router)

The nginx chart can be configured to act as a central gateway/router for OPEA services. This mode provides:

- Service routing to multiple OPEA microservices
- Custom nginx configuration template support
- Health check endpoints
- Optional ingress and monitoring support

### Quick Start

```bash
# Deploy as gateway
helm install nginx-gateway . -f gateway-values.yaml

# Access OPEA services through the gateway
curl http://nginx-gateway/v1/chatqna -d '{"query": "Hello"}'
curl http://nginx-gateway/v1/health_check
```

### Configuration

The `gateway-values.yaml` file contains:
- Service endpoint definitions for routing
- UI service configuration
- Custom nginx configuration template
- Optional ingress and monitoring settings

## Deployment Modes

This chart supports two deployment modes:

- **Simple Proxy Mode** (default): Acts as a reverse proxy within OPEA application charts
- **Gateway Mode**: Central router providing unified access to multiple OPEA services

All existing deployments continue to work unchanged. Gateway mode is opt-in and only activated when using `gateway-values.yaml`.
