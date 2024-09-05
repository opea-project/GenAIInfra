# Usage guide for genai-microservices-connector(GMC)

genai-microservices-connector(GMC) can be used to compose and adjust GenAI pipelines dynamically. It can leverage the microservices provided
by [GenAIComps](https://github.com/opea-project/GenAIComps) and external services to compose GenAI pipelines.

Below are sample use cases:

## Use GMC to compose a chatQnA Pipeline

A sample for chatQnA can be found at config/samples/ChatQnA/chatQnA_dataprep_xeon.yaml

**Deploy chatQnA GMC custom resource**

```sh
kubectl create ns chatqa
kubectl apply -f $(pwd)/config/samples/ChatQnA/chatQnA_dataprep_xeon.yaml
# To use Gaudi device
#kubectl apply -f $(pwd)/config/samples/ChatQnA/chatQnA_dataprep_gaudi.yaml
# To use Nvidia GPU
#kubectl apply -f $(pwd)/config/samples/ChatQnA/chatQnA_nv.yaml
```

**GMC will reconcile chatQnA custom resource and get all related components/services ready**

```sh
kubectl get service -n chatqa
```

**Check GMC chatQnA custom resource to get access URL for the pipeline**

```bash
$kubectl get gmconnectors.gmc.opea.io -n chatqa
NAME     URL                                                      READY     AGE
chatqa   http://router-service.chatqa.svc.cluster.local:8080      10/0/10     3m
```

the `READY 10/0/10` means there are 10(the 2nd 10) services deployed by the GMC and 10(the 1st 10) are ready, so the 10 of 10 means the pipeline is all set. the 0 in the middle means there are no external services used, all the resources are managed by GMC inside the clusters.`

you can get the resources via `kubectl` commands

```
$ kubectl get pods -n chatqa
NAME                                            READY   STATUS    RESTARTS   AGE
data-prep-svc-deployment-68f7c5dcb9-8fbh8       1/1     Running   0          2m41s
embedding-svc-deployment-775bd5dc49-j4ltr       1/1     Running   0          2m43s
llm-svc-deployment-59f756fb56-4xckz             1/1     Running   0          2m41s
redis-vector-db-deployment-587844d666-hbchr     1/1     Running   0          2m42s
reranking-svc-deployment-846c89f79f-gv7b9       1/1     Running   0          2m42s
retriever-svc-deployment-5c44f7d46-m4qgq        1/1     Running   0          2m43s
router-service-deployment-7f6c5f4796-tzchw      1/1     Running   0          2m41s
tei-embedding-svc-deployment-54b58d57cb-9mwvk   1/1     Running   0          2m43s
tei-reranking-svc-deployment-54c5dd5795-b6wcb   1/1     Running   0          2m42s
tgi-service-m-deployment-5ff67f4db7-b7ztj       1/1     Running   0          2m41s
```

you can also get the detailed information of these resource by checking the pipeline's status, this will list all the configmap, deployment and service and their status as below:

```
$ kubectl get gmc -n chatqa chatqa -o json | jq '.status.annotations' | yq -P
ConfigMap:v1:data-prep-config:chatqa: provisioned
ConfigMap:v1:embedding-usvc-config:chatqa: provisioned
ConfigMap:v1:llm-uservice-config:chatqa: provisioned
ConfigMap:v1:reranking-usvc-config:chatqa: provisioned
ConfigMap:v1:retriever-usvc-config:chatqa: provisioned
ConfigMap:v1:tei-config:chatqa: provisioned
ConfigMap:v1:teirerank-config:chatqa: provisioned
ConfigMap:v1:tgi-config:chatqa: provisioned
Deployment:apps/v1:data-prep-svc-deployment:chatqa: |
  Replicas: 1 desired | 1 updated | 1 total | 1 available | 0 unavailable
  Conditions:
    Type: Available
    Status: True
    Reason: MinimumReplicasAvailable
    Message: Deployment has minimum availability.
    Type: Progressing
    Status: True
    Reason: NewReplicaSetAvailable
    Message: ReplicaSet "data-prep-svc-deployment-7c7c648846" has successfully progressed.
Deployment:apps/v1:embedding-svc-deployment:chatqa: |
  Replicas: 1 desired | 1 updated | 1 total | 1 available | 0 unavailable
  Conditions:
    Type: Available
    Status: True
    Reason: MinimumReplicasAvailable
    Message: Deployment has minimum availability.
    Type: Progressing
    Status: True
    Reason: NewReplicaSetAvailable
    Message: ReplicaSet "embedding-svc-deployment-775bd5dc49" has successfully progressed.
Deployment:apps/v1:llm-svc-deployment:chatqa: |
  Replicas: 1 desired | 1 updated | 1 total | 1 available | 0 unavailable
  Conditions:
    Type: Available
    Status: True
    Reason: MinimumReplicasAvailable
    Message: Deployment has minimum availability.
    Type: Progressing
    Status: True
    Reason: NewReplicaSetAvailable
    Message: ReplicaSet "llm-svc-deployment-59f756fb56" has successfully progressed.
Deployment:apps/v1:redis-vector-db-deployment:chatqa: |
  Replicas: 1 desired | 1 updated | 1 total | 1 available | 0 unavailable
  Conditions:
    Type: Available
    Status: True
    Reason: MinimumReplicasAvailable
    Message: Deployment has minimum availability.
    Type: Progressing
    Status: True
    Reason: NewReplicaSetAvailable
    Message: ReplicaSet "redis-vector-db-deployment-587844d666" has successfully progressed.
Deployment:apps/v1:reranking-svc-deployment:chatqa: |
  Replicas: 1 desired | 1 updated | 1 total | 1 available | 0 unavailable
  Conditions:
    Type: Available
    Status: True
    Reason: MinimumReplicasAvailable
    Message: Deployment has minimum availability.
    Type: Progressing
    Status: True
    Reason: NewReplicaSetAvailable
    Message: ReplicaSet "reranking-svc-deployment-846c89f79f" has successfully progressed.
Deployment:apps/v1:retriever-svc-deployment:chatqa: |
  Replicas: 1 desired | 1 updated | 2 total | 1 available | 1 unavailable
  Conditions:
    Type: Available
    Status: True
    Reason: MinimumReplicasAvailable
    Message: Deployment has minimum availability.
    Type: Progressing
    Status: True
    Reason: ReplicaSetUpdated
    Message: ReplicaSet "retriever-svc-deployment-95b967c9d" is progressing.
Deployment:apps/v1:router-service-deployment:chatqa: |
  Replicas: 1 desired | 1 updated | 1 total | 1 available | 0 unavailable
  Conditions:
    Type: Available
    Status: True
    Reason: MinimumReplicasAvailable
    Message: Deployment has minimum availability.
    Type: Progressing
    Status: True
    Reason: NewReplicaSetAvailable
    Message: ReplicaSet "router-service-deployment-79f54548f4" has successfully progressed.
Deployment:apps/v1:tei-embedding-svc-deployment:chatqa: |
  Replicas: 1 desired | 1 updated | 1 total | 1 available | 0 unavailable
  Conditions:
    Type: Available
    Status: True
    Reason: MinimumReplicasAvailable
    Message: Deployment has minimum availability.
    Type: Progressing
    Status: True
    Reason: NewReplicaSetAvailable
    Message: ReplicaSet "tei-embedding-svc-deployment-54b58d57cb" has successfully progressed.
Deployment:apps/v1:tei-reranking-svc-deployment:chatqa: |
  Replicas: 1 desired | 1 updated | 1 total | 1 available | 0 unavailable
  Conditions:
    Type: Available
    Status: True
    Reason: MinimumReplicasAvailable
    Message: Deployment has minimum availability.
    Type: Progressing
    Status: True
    Reason: NewReplicaSetAvailable
    Message: ReplicaSet "tei-reranking-svc-deployment-54c5dd5795" has successfully progressed.
Deployment:apps/v1:tgi-service-m-deployment:chatqa: |
  Replicas: 1 desired | 1 updated | 1 total | 1 available | 0 unavailable
  Conditions:
    Type: Available
    Status: True
    Reason: MinimumReplicasAvailable
    Message: Deployment has minimum availability.
    Type: Progressing
    Status: True
    Reason: NewReplicaSetAvailable
    Message: ReplicaSet "tgi-service-m-deployment-5fcff459f5" has successfully progressed.
Service:v1:data-prep-svc:chatqa: http://data-prep-svc.chatqa.svc.cluster.local:6007/v1/dataprep
Service:v1:embedding-svc:chatqa: http://embedding-svc.chatqa.svc.cluster.local:6000/v1/embeddings
Service:v1:llm-svc:chatqa: http://llm-svc.chatqa.svc.cluster.local:9000/v1/chat/completions
Service:v1:redis-vector-db:chatqa: http://redis-vector-db.chatqa.svc.cluster.local:6379
Service:v1:reranking-svc:chatqa: http://reranking-svc.chatqa.svc.cluster.local:8000/v1/reranking
Service:v1:retriever-svc:chatqa: http://retriever-svc.chatqa.svc.cluster.local:7000/v1/retrieval
Service:v1:router-service:chatqa: http://router-service.chatqa.svc.cluster.local:8080
Service:v1:tei-embedding-svc:chatqa: http://tei-embedding-svc.chatqa.svc.cluster.local:80
Service:v1:tei-reranking-svc:chatqa: http://tei-reranking-svc.chatqa.svc.cluster.local:80/rerank
Service:v1:tgi-service-m:chatqa: http://tgi-service-m.chatqa.svc.cluster.local:80/generate
```

**NOTE: if you upgrade from pre 0.9 to 0.9 or later, you might encounter below issue**

if the router-service and it's deployment are not initialized, which is mandatory for every pipeline, you also need to upgrade the gmc-router.yaml to the latest version which is mentioned in the [GMC README](https://github.com/opea-project/GenAIInfra/blob/main/microservices-connector/README.md)

**Deploy one client pod for testing the chatQnA application**

```bash
kubectl create deployment client-test -n chatqa --image=python:3.8.13 -- sleep infinity
```

**Access the pipeline using the above URL from the client pod**

```bash
export CLIENT_POD=$(kubectl get pod -n chatqa  -l app=client-test -o jsonpath={.items..metadata.name})
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

**Remove one step of the pipeline**
If you want to adjust the steps of the pipeline, for example, if you want to delete the data preparation step from chatQnA, you can simply delete this part from the yaml file config/samples/chatQnA_dataprep_xeon.yaml

```
      - name: DataPrep
        internalService:
          serviceName: data-prep-svc
          config:
            endpoint: /v1/dataprep
            REDIS_URL: redis-vector-db
            TEI_ENDPOINT: tei-embedding-svc
          isDownstreamService: true
```

and re-apply the yaml file

```
kubectl apply -f $(pwd)/config/samples/chatQnA_dataprep_xeon.yaml
```

you would see the `dataprep` is deleted

```
$ kubectl get gmc -n chatqa chatqa
NAME     URL                                                   READY   AGE
chatqa   http://router-service.chatqa.svc.cluster.local:8080   9/0/9   3m37s
```

But please be noted, **you have to make sure** the step is eligible to be deleted without affecting the pipeline function.

## Use GMC to delete the chatQnA Pipeline

you can delete all the resources by deleting the gmc custom resource

```
$ kubectl delete gmc -n chatqa chatqa
gmconnector.gmc.opea.io "chatqa" deleted

$ kubectl get gmc -n chatqa
No resources found in chatqa namespace.

$ kubectl get all -n chatqa
No resources found in chatqa namespace.
```

## Use GMC and Istio to compose an OPEA Pipeline with authentication and authorization enabled

The critical steps of authentication and authorization are vital to maintaining the integrity and safety of our GenAI workload. Please check the [readme](../authN-authZ/README.md) file for more details.
