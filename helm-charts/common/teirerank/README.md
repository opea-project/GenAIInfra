# teirerank

Helm chart for deploying Hugging Face Text Generation Inference service.

## Installing the Chart

To install the chart, run the following:

```console
cd ${GenAIInfro_repo}/helm-charts/common
export MODELDIR=/mnt/opea-models
export MODELNAME="BAAI/bge-reranker-base"
helm install teirerank teirerank --set global.modelUseHostPath=${MODELDIR} --set RERANK_MODEL_ID=${MODELNAME}
```

By default, the teirerank service will downloading the "BAAI/bge-reranker-base" which is about 1.1GB.

If you already cached the model locally, you can pass it to container like this example:

MODELDIR=/mnt/opea-models

MODELNAME="/data/BAAI/bge-reranker-base"

## HorizontalPodAutoscaler (HPA) support

`horizontalPodAutoscaler` option enables HPA scaling for the deployment:
https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/

Autoscaling is based on custom application metrics provided through [Prometheus](https://prometheus.io/).

### Pre-conditions

If cluster does not run [Prometheus operator](https://github.com/prometheus-operator/kube-prometheus)
yet, it SHOULD be be installed before enabling HPA, e.g. by using:
https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack

`horizontalPodAutoscaler` enabled in top level Helm chart depending on this component (e.g. `chatqna`),
so that relevant custom metric queries are configured for PrometheusAdapter.

### Gotchas

Why HPA is opt-in:

- Enabling chart `horizontalPodAutoscaler` option will _overwrite_ cluster's current
  `PrometheusAdapter` configuration with its own custom metrics configuration.
  Take copy of the existing one before install, if that matters:
  `kubectl -n monitoring get cm/adapter-config -o yaml > adapter-config.yaml`
- `PrometheusAdapter` needs to be restarted after install, for it to read the new configuration:
  `ns=monitoring; kubectl -n $ns delete $(kubectl -n $ns get pod --selector app.kubernetes.io/name=prometheus-adapter -o name)`
- By default Prometheus adds [k8s RBAC rules](https://github.com/prometheus-operator/kube-prometheus/blob/main/manifests/prometheus-roleBindingSpecificNamespaces.yaml)
  for accessing metrics from `default`, `kube-system` and `monitoring` namespaces. If Helm is
  asked to install OPEA services to some other namespace, those rules need to be updated accordingly
- Provided HPA rules are examples for Xeon, for efficient scaling they need to be fine-tuned for given setup
  (underlying HW, used models, OPEA version etc)

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are runinng.

Then run the command `kubectl port-forward svc/teirerank 2082:80` to expose the tei service for access.

Open another terminal and run the following command to verify the service if working:

```console
curl http://localhost:2082/rerank \
    -X POST \
    -d '{"query":"What is Deep Learning?", "texts": ["Deep Learning is not...", "Deep learning is..."]}' \
    -H 'Content-Type: application/json'
```

### Verify HPA metrics

To verify that metrics required by horizontalPodAutoscaler option work, check that:

Prometheus has found the metric endpoints, i.e. last number on the line is non-zero:

```console
prom_url=http://$(kubectl -n monitoring get -o jsonpath="{.spec.clusterIP}:{.spec.ports[0].port}" svc/prometheus-k8s)
curl --no-progress-meter $prom_url/metrics | grep scrape_pool_targets.*rerank
```

Prometheus adapter provides custom metrics for their data:

```console
kubectl get --raw /apis/custom.metrics.k8s.io/v1beta1 | jq .resources[].name
```

And those custom metrics have valid values for HPA rules:

```console
ns=default;  # OPEA namespace
url=/apis/custom.metrics.k8s.io/v1beta1;
for m in $(kubectl get --raw $url | jq .resources[].name | cut -d/ -f2 | sort -u | tr -d '"'); do
  kubectl get --raw $url/namespaces/$ns/metrics/$m | jq;
done | grep -e metricName -e value
```

NOTE: HuggingFace TGI and TEI services provide metrics endpoint only after they've processed their first request!

## Values

| Key                             | Type   | Default                                           | Description                                                                                                                                                                                                                 |
| ------------------------------- | ------ | ------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| RERANK_MODEL_ID                 | string | `"BAAI/bge-reranker-base"`                        | Models id from https://huggingface.co/, or predownloaded model directory                                                                                                                                                    |
| global.modelUseHostPath         | string | `"/mnt/opea-models"`                              | Cached models directory, teirerank will not download if the model is cached here. The host path "modelUseHostPath" will be mounted to container as /data directory. Set this to null/empty will force it to download model. |
| image.repository                | string | `"ghcr.io/huggingface/text-embeddings-inference"` |                                                                                                                                                                                                                             |
| image.tag                       | string | `"cpu-1.5"`                                       |                                                                                                                                                                                                                             |
| horizontalPodAutoscaler.enabled | bool   | false                                             | Enable HPA autoscaling for the service deployments based on metrics it provides. See #pre-conditions and #gotchas before enabling!                                                                                          |
