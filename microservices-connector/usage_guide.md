# Usage guide for genai-microservices-connector(GMC)

genai-microservices-connector(GMC) can be used to compose and adjust GenAI pipelines dynamically. It can leverage the microservices provided
by [GenAIComps](https://github.com/opea-project/GenAIComps) and external services to compose GenAI pipelines.

Below are sample use cases:

## Use GMC to compose a chatQnA Pipeline

A sample for chatQnA can be found at config/samples/chatQnA_xeon.yaml

**Deploy chatQnA GMC custom resource**

```sh
kubectl create ns chatqa
kubectl apply -f $(pwd)/config/samples/chatQnA_xeon.yaml
```

**GMC will reconcile chatQnA custom resource and get all related components/services ready**

```sh
kubectl get service -n chatqa
```

**Check GMC chatQnA custom resource to get access URL for the pipeline**

```bash
$kubectl get gmconnectors.gmc.opea.io -n chatqa
NAME     URL                                                      READY     AGE
chatqa   http://router-service.chatqa.svc.cluster.local:8080      8/0/8     3m
```

**Deploy one client pod for testing the chatQnA application**

```bash
kubectl create deployment client-test -n chatqa --image=python:3.8.13 -- sleep infinity
```

**Access the pipeline using the above URL from the client pod**

```bash
export CLIENT_POD=$(kubectl get pod  -l app=client-test -o jsonpath={.items..metadata.name})
export accessUrl=$(kubectl get gmc -n chatqa -o jsonpath="{.items[?(@.metadata.name=='chatqa')].status.accessUrl}")
kubectl exec "$CLIENT_POD" -n chatqa -- curl $accessUrl  -X POST  -d '{"text":"What is the revenue of Nike in 2023?","parameters":{"max_new_tokens":17, "do_sample": true}}' -H 'Content-Type: application/json'
```

## Use GMC to adjust the chatQnA Pipeline

**Modify chatQnA custom resource to change to another LLM model**

```yaml
- name: Tgi
  internalService:
    serviceName: tgi-svc
    config:
      LLM_MODEL_ID: Llama-2-7b-chat-hf
```

**Check the tgi-svc-deployment has been changed to use the new LLM Model**

```sh
kubectl get deployment tgi-svc-deployment -n chatqa -o jsonpath="{.spec.template.spec.containers[*].env[?(@.name=='LLM_MODEL_ID')].value}"
```

**Access the updated pipeline using the above URL from the client pod**

```bash
kubectl exec "$CLIENT_POD" -n chatqa -- curl $accessUrl  -X POST  -d '{"text":"What is the revenue of Nike in 2023?","parameters":{"max_new_tokens":17, "do_sample": true}}' -H 'Content-Type: application/json'
```
