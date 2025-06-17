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

### How to Use Gateway Mode

Gateway mode is designed to be used within E2E (end-to-end) charts. The pattern is:

1. **Create an E2E chart** that includes nginx as a dependency
2. **Define environment variables** in your E2E chart's `{{ .Release.Name }}-nginx-config` ConfigMap
3. **Use custom nginx template** with environment variable substitution

### Example Implementation

```yaml
# In your E2E chart's templates/nginx-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Release.Name }}-nginx-config
data:
  FRONTEND_SERVICE_IP: "ui.ui.svc.cluster.local"
  FRONTEND_SERVICE_PORT: "5173"
  CHATQNA_SERVICE_IP: "chatqna.chatqna.svc.cluster.local"
  CHATQNA_SERVICE_PORT: "8888"
  # ... other service endpoints
```

```yaml
# In your E2E chart's values.yaml
nginx:
  nginxConfig:
    enabled: true
    template: |
      server {
        location / {
          proxy_pass http://${FRONTEND_SERVICE_IP}:${FRONTEND_SERVICE_PORT};
        }
        location /v1/chatqna {
          proxy_pass http://${CHATQNA_SERVICE_IP}:${CHATQNA_SERVICE_PORT}/v1/chatqna;
        }
        # ... other routes
      }
```

See `gateway-values.yaml` for a complete example template.

## Deployment Modes

This chart supports two deployment modes:

- **Simple Proxy Mode** (default): Acts as a reverse proxy within OPEA application charts
- **Gateway Mode**: Central router providing unified access to multiple OPEA services (used within E2E charts)

All existing deployments continue to work unchanged. Gateway mode requires creating an E2E chart that follows the pattern shown above.
