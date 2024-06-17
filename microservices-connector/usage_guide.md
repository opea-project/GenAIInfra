# Usage guide for genai-microservices-connector(GMC)

genai-microservices-connector(GMC) can be used to compose and adjust GenAI pipelines dynamically. It can leverage the microservices provided
by [GenAIComps](https://github.com/opea-project/GenAIComps) and external services to compose GenAI pipelines.

Below are sample users cases:

## Use GMC to compose a chatQnA Pipeline

A sample for chatQnA can be found at config/samples/chatQnA_xeon.yaml

**Deploy chatQnA GMC custom resource**

```sh
kubectl create ns chatqa
kubectl apply -f $(pwd)/config/samples/chatQnA_xeon.yaml
```

**GMC will reconcile chatQnA custom resource and get all related components/service ready**

```sh
kubectl get service -n chatqa
```

**Check GMC chatQnA custom resource to get access URL for the pipeline**

```bash
$kubectl get gmconnectors.gmc.opea.io -n chatqa
NAME     URL                                                      READY     AGE
chatqa   http://router-service.chatqa.svc.cluster.local:8080   Success   3m
```

**Access the pipeline using the URL**

## Use GMC to adjust the chatQnA Pipeline

**Modify chatQnA custom resource to change to another LLM model**

```yaml
internalService:
  serviceName: llm-service
  config:
    LLM_MODEL_ID: Llama-2-7b-chat-hf
