<h1 align="center" id="title">Deploy CodeGen in Kubernetes Cluster on Xeon</h1>

## Prebuilt images

You should have all the images

- tgi: ghcr.io/huggingface/text-generation-inference:1.4
- llm-uservice: opea/gen-ai-comps:llm-tgi-server
- codegen: opea/gen-ai-comps:codegen-megaservice-server

> [NOTE]  
> Please refer this OPEA repo https://github.com/opea-project/GenAIExamples/blob/main/CodeGen/docker-composer/xeon/README.md to build some opea images

## Deploy Services

> [NOTE]
> The following values must be set before you can deploy
> HUGGINGFACEHUB_API_TOKEN

```
cd ${RepoPath}/manifests/CodeGen/xeon
export HUGGINGFACEHUB_API_TOKEN="YourOwnToken"
sed -i "s/insert-your-huggingface-token-here/${HUGGINGFACEHUB_API_TOKEN}/g" codegen.yaml
kubectl apply -f codegen.yaml
```

## Verify Services

Make sure all the pods are running, and restart the codegen-xxxx pod if necessary.

```
curl http://codegen:6666/v1/codegen -H "Content-Type: application/json" -d '{
     "messages": "Implement a high-level API for a TODO list application. The API takes as input an operation request and updates the TODO list in place. If the request is invalid, raise an exception."
     }'
```
