# Deploy DocSum llm-uservice in Kubernetes Cluster

> [NOTE]
> The following values must be set before you can deploy:
> HUGGINGFACEHUB_API_TOKEN

## Deploy On Xeon

```
cd GenAIInfra/manifests/DocSum/xeon
export HUGGINGFACEHUB_API_TOKEN="YourOwnToken"
sed -i "s/insert-your-huggingface-token-here/${HUGGINGFACEHUB_API_TOKEN}/g" llm.yaml
kubectl apply -f llm.yaml
```

## Deploy On Gaudi

```
cd GenAIInfra/manifests/DocSum/gaudi
export HUGGINGFACEHUB_API_TOKEN="YourOwnToken"
sed -i "s/insert-your-huggingface-token-here/${HUGGINGFACEHUB_API_TOKEN}/g" llm.yaml
kubectl apply -f llm.yaml
```

## Verify llm Services

Make sure all the pods are running, and restart the llm-xxxx pod if necessary.

```
kubectl get pods
curl http://docsum-llm-uservice:9000/v1/chat/docsum \
  -X POST \
  -d '{"query":"Text Embeddings Inference (TEI) is a toolkit for deploying and serving open source text embeddings and sequence classification models. TEI enables high-performance extraction for the most popular models, including FlagEmbedding, Ember, GTE and E5."}' \
  -H 'Content-Type: application/json'
```

## Generate the llm file from helm chart

The llm file is generated from llm-uservice helm chart automatically.

Here is the exact command lines:

```
cd GenAIInfra/manifests/DocSum
export HF_TOKEN="insert-your-huggingface-token-here"
export MODELDIR="/mnt"
helm template docsum ../../helm-charts/common/llm-uservice --set global.HUGGINGFACEHUB_API_TOKEN=${HF_TOKEN} --set image.repository="opea/llm-docsum-tgi:latest" --set tgi.volume=${MODELDIR} --set tgi.LLM_MODEL_ID="Intel/neural-chat-7b-v3-3" --values ../../helm-charts/common/llm-uservice/values.yaml > xeon/llm.yaml
helm template docsum ../../helm-charts/common/llm-uservice --set global.HUGGINGFACEHUB_API_TOKEN=${HF_TOKEN} --set image.repository="opea/llm-docsum-tgi:latest" --set tgi.volume=${MODELDIR} --set tgi.LLM_MODEL_ID="Intel/neural-chat-7b-v3-3" --values ../../helm-charts/common/llm-uservice/gaudi-values.yaml > gaudi/llm.yaml

```
