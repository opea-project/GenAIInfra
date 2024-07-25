# Deploy CodeTrans llm-uservice in Kubernetes Cluster

> [NOTE]
> The following values must be set before you can deploy:
> HUGGINGFACEHUB_API_TOKEN

## Deploy On Xeon

```
cd GenAIInfra/manifests/CodeTrans/xeon
export HUGGINGFACEHUB_API_TOKEN="YourOwnToken"
sed -i "s/insert-your-huggingface-token-here/${HUGGINGFACEHUB_API_TOKEN}/g" llm.yaml
kubectl apply -f llm.yaml
```

## Deploy On Gaudi

```
cd GenAIInfra/manifests/CodeTrans/gaudi
export HUGGINGFACEHUB_API_TOKEN="YourOwnToken"
sed -i "s/insert-your-huggingface-token-here/${HUGGINGFACEHUB_API_TOKEN}/g" llm.yaml
kubectl apply -f llm.yaml
```

## Verify llm Services

Make sure all the pods are running, and restart the llm-xxxx pod if necessary.

```
kubectl get pods
curl http://codetrans-llm-uservice:9000/v1/chat/completions\
  -X POST \
  -d '{"query":"    ### System: Please translate the following Golang codes into  Python codes.    ### Original codes:    '\'''\'''\''Golang    \npackage main\n\nimport \"fmt\"\nfunc main() {\n    fmt.Println(\"Hello, World!\");\n    '\'''\'''\''    ### Translated codes:"}' \
  -H 'Content-Type: application/json'
```

## Generate the llm file from helm chart

The llm file is generated from llm-uservice helm chart automatically.

Here is the exact command lines:

```
cd GenAIInfra/manifests/CodeTrans
export HF_TOKEN="insert-your-huggingface-token-here"
export MODELDIR="/mnt"
helm template codetrans ../../helm-charts/common/llm-uservice --set global.HUGGINGFACEHUB_API_TOKEN=${HF_TOKEN} --set image.repository="opea/llm-tgi:latest" --set tgi.volume=${MODELDIR} --set tgi.LLM_MODEL_ID="HuggingFaceH4/mistral-7b-grok" --values ../../helm-charts/common/llm-uservice/values.yaml > xeon/llm.yaml
helm template codetrans ../../helm-charts/common/llm-uservice --set global.HUGGINGFACEHUB_API_TOKEN=${HF_TOKEN} --set image.repository="opea/llm-tgi:latest" --set tgi.volume=${MODELDIR} --set tgi.LLM_MODEL_ID="HuggingFaceH4/mistral-7b-grok" --values ../../helm-charts/common/llm-uservice/gaudi-values.yaml > gaudi/llm.yaml

```
