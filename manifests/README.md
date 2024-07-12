<h1 align="center" id="title">Kubernetes manifests for GenAIComps</h1>

> [NOTE]
> The manifests are generated from helm charts automatically.
> For more details, see README.md in ../helm-charts/common/${component}.

## Deploy On Xeon

```
cd GenAIInfra/manifests/${component}/xeon
export HUGGINGFACEHUB_API_TOKEN="YourOwnToken"
sed -i "s/insert-your-huggingface-token-here/${HUGGINGFACEHUB_API_TOKEN}/g" *.yaml
kubectl apply -f *.yaml
```

## Deploy On Gaudi

```
cd GenAIInfra/manifests/${component}/gaudi
export HUGGINGFACEHUB_API_TOKEN="YourOwnToken"
sed -i "s/insert-your-huggingface-token-here/${HUGGINGFACEHUB_API_TOKEN}/g" *.yaml
kubectl apply -f *.yaml
```

## Generate the manifest file from helm chart

Refer to update_manifests.sh for details.

Here is the one example:

```
cd GenAIInfra/manifests/CodeTrans
export HF_TOKEN="insert-your-huggingface-token-here"
export MODELDIR="/mnt"
helm template codetrans ../../helm-charts/common/llm-uservice --set global.HUGGINGFACEHUB_API_TOKEN=${HF_TOKEN} --set image.repository="opea/llm-tgi:latest" --set tgi.volume=${MODELDIR} --set tgi.LLM_MODEL_ID="HuggingFaceH4/mistral-7b-grok" --values ../../helm-charts/common/llm-uservice/values.yaml > xeon/llm.yaml
helm template codetrans ../../helm-charts/common/llm-uservice --set global.HUGGINGFACEHUB_API_TOKEN=${HF_TOKEN} --set image.repository="opea/llm-tgi:latest" --set tgi.volume=${MODELDIR} --set tgi.LLM_MODEL_ID="HuggingFaceH4/mistral-7b-grok" --values ../../helm-charts/common/llm-uservice/gaudi-values.yaml > gaudi/llm.yaml

```
