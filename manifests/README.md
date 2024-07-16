<h1 align="center" id="title">Kubernetes manifests for GenAIComps</h1>

> [NOTE]
> The manifests are generated from helm charts automatically.
> For more details, see README.md in ../helm-charts/common/${component}.

## Deploy On Xeon

```
cd GenAIInfra/manifests/common
export component=<component>
export HUGGINGFACEHUB_API_TOKEN="YourOwnToken"
sed -i "s/insert-your-huggingface-token-here/${HUGGINGFACEHUB_API_TOKEN}/g" ${component}.yaml
kubectl apply -f ${component}.yaml
```

## Deploy On Gaudi

```
cd GenAIInfra/manifests/common
export component=<component>
export HUGGINGFACEHUB_API_TOKEN="YourOwnToken"
# if there is a manifest file named after ${component}_gaudi.yaml
sed -i "s/insert-your-huggingface-token-here/${HUGGINGFACEHUB_API_TOKEN}/g" ${component}_gaudi.yaml
kubectl apply -f ${component}_gaudi.yaml
# else apply the xeon manifest file
sed -i "s/insert-your-huggingface-token-here/${HUGGINGFACEHUB_API_TOKEN}/g" ${component}.yaml
kubectl apply -f ${component}.yaml
```

## Generate the manifest file from helm chart

Refer to [update_manifests.sh](update_manifests.sh) for details.
