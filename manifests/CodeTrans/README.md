<h1 align="center" id="title">Deploy CodeTrans llm-uservice in Kubernetes Cluster</h1>

> [NOTE]
> The following values must be set before you can deploy:
> HUGGINGFACEHUB_API_TOKEN
> You can also customize the "MODEL_ID" and "model-volume"
> The manifest llm.yaml is generated from helm chart.

## Deploy On Xeon

```
cd GenAIExamples/CodeTrans/kubernetes/manifests/xeon
export HUGGINGFACEHUB_API_TOKEN="YourOwnToken"
sed -i "s/insert-your-huggingface-token-here/${HUGGINGFACEHUB_API_TOKEN}/g" llm.yaml
kubectl apply -f llm.yaml
```

## Deploy On Gaudi

```
cd GenAIExamples/CodeTrans/kubernetes/manifests/gaudi
export HUGGINGFACEHUB_API_TOKEN="YourOwnToken"
sed -i "s/insert-your-huggingface-token-here/${HUGGINGFACEHUB_API_TOKEN}/g" llm.yaml
kubectl apply -f llm.yaml
```

## Verify llm Services

Make sure all the pods are running, and restart the llm-xxxx pod if necessary.

```
kubectl get pods
curl http://llm-llm-uservice:9000/v1/chat/completions\
  -X POST \
  -d '{"query":"    ### System: Please translate the following Golang codes into  Python codes.    ### Original codes:    '\'''\'''\''Golang    \npackage main\n\nimport \"fmt\"\nfunc main() {\n    fmt.Println(\"Hello, World!\");\n    '\'''\'''\''    ### Translated codes:"}' \
  -H 'Content-Type: application/json'
```
