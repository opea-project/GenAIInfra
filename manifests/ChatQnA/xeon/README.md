<h1 align="center" id="title">Deploy ChatQnA in Kubernetes Cluster on Xeon</h1>

## Prebuilt images

You should have all the images

- redis-vector-db: redis/redis-stack:7.2.0-v9
- tei_embedding_service: ghcr.io/huggingface/text-embeddings-inference:cpu-1.2
- embedding: opea/gen-ai-comps:embedding-tei-server
- retriever: opea/gen-ai-comps:retriever-redis-server
- tei_xeon_service: ghcr.io/huggingface/text-embeddings-inference:cpu-1.2
- reranking: opea/gen-ai-comps:reranking-tei-xeon-server
- tgi_service: ghcr.io/huggingface/text-generation-inference:1.4
- llm: opea/gen-ai-comps:llm-tgi-server

> [NOTE]  
> Please refer this OPEA repo https://github.com/opea-project/GenAIExamples/blob/main/ChatQnA/microservice/xeon/README.md to build some opea images

## Deploy Services

> [NOTE]
> Be sure to modify the image name in all yaml files by your own value
> Be sure to modify the all the value in qna_configmap.yaml
> Be sure to replace "insert-your-huggingface-token-here" in qna_configmap.yaml with a real huggingface token

```
$ cd ${RepoPath}/manifests/ChatQnA/manifests/xeon
```

#### 1.1 Deploy QnA configmap

```
kubectl apply -f qna_configmap.yaml

```

#### 1.2 Deploy Services

```
# 1. deloy redis-vector-db
kubectl apply -f redis-vector-db.yaml

# 2. deploy tei_embedding_service
kubectl apply -f tei_embedding_service.yaml

# 3. deploy embedding
kubectl apply -f embedding.yaml

# 4. deploy retriever
kubectl apply -f retriever.yaml

# 5. deploy tei_xeon_service
kubectl apply -f tei_xeon_service.yaml

# 6. deploy reranking
kubectl apply -f reranking.yaml

# 7. deploy tgi_service
kubectl apply -f tgi_service.yaml

# 8. deploy llm
kubectl apply -f llm.yaml

# 9. deploy chaqna-xeon-backend-server
kubectl apply -f chaqna-xeon-backend-server.yaml

# 10. deploy chaqna-xeon-ui-server
kubectl apply -f chaqna-xeon-ui-server.yaml


```

## Verify Services

```
$ chaqna_backend_svc_ip=`kubectl get svc|grep '^chaqna-xeon-backend-server-svc'|awk '{print $3}'` && echo ${chaqna_backend_svc_ip}
$ curl http://${chaqna_backend_svc_ip}:8888/v1/chatqna -H "Content-Type: application/json" -d '{
     "model": "Intel/neural-chat-7b-v3-3",
     "messages": "What is the revenue of Nike in 2023?"
     }'
```
