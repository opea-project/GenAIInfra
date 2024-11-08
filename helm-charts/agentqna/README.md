# AgentQnA

Helm chart for deploying AgentQnA service.

## Deploy

helm repo add opea https://opea-project.github.io/GenAIInfra

helm install agentqna opea/agentqna --set global.HUGGINGFACEHUB_API_TOKEN=${HUGGINGFACEHUB_API_TOKEN} --set tgi.enabled=True

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are running.

### Verify the workload through curl command

Run the command `kubectl port-forward svc/agentqna-supervisor 9090:9090` to expose the service for access.

Open another terminal and run the following command to verify the service if working:

```console
curl http://localhost:9090/v1/chat/completions \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{"query": "Most recent album by Michael Jackson"}'
```
