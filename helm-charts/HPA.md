# HorizontalPodAutoscaler (HPA) support

## Table of Contents

- [Introduction](#introduction)
- [Pre-conditions](#pre-conditions)
- [Gotchas](#gotchas)
- [Enable](#enable)
- [Verify](#verify)

## Introduction

`horizontalPodAutoscaler` option enables HPA scaling for the TGI and TEI inferencing deployments:
https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/

Autoscaling is based on custom application metrics provided through [Prometheus](https://prometheus.io/).

### Pre-conditions

HPA controlled pods SHOULD have appropriate resource requests or affinity rules (enabled in their
subcharts and tested to work) so that k8s scheduler does not schedule too many of them on the same
node(s). Otherwise they never reach ready state.

Too large requests would not be a problem as long as pods still fit to available nodes, but too
small requests would be an issue:

- Multiple inferencing instances interfere / slow down each other, especially if there are no
  [NRI policies](https://github.com/opea-project/GenAIEval/tree/main/doc/platform-optimization)
  that provide further isolation
- Containers can become non-functional when their actual resource usage crosses the specified limits

If cluster does not run [Prometheus operator](https://github.com/prometheus-operator/kube-prometheus)
yet, it SHOULD be be installed before enabling HPA, e.g. by using:
https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack

Enabling HPA in top-level Helm chart (e.g. `chatqna`), overwrites cluster's current _PrometheusAdapter_
configuration with relevant custom metric queries. If that has existing queries that should be retained,
relevant queries need to be added to existing _PrometheusAdapter_ configuration _manually_ from the
custom metrics Helm template (in top-level Helm chart).

Names of the _Prometheus-operator_ related objects depend on where it is installed from.
Default ones are:

- "kube-prometheus" upstream manifests:
  - Namespace: `monitoring`
  - Metrics service: `prometheus-k8s`
  - Adapter configMap: `adapter-config`
- Helm chart for "kube-prometheus" (linked above):
  - Namespace: `monitoring`
  - Metrics service: `prom-kube-prometheus-stack-prometheus`
  - Adapter configMap: `prom-adapter-prometheus-adapter`

Make sure correct "configMap" name is used in top-level (e.g. `chatqna`) Helm chart
`hpa-values.yaml` file, and in commands below!

### Gotchas

Why HPA is opt-in:

- Enabling (top level) chart `horizontalPodAutoscaler` option will _overwrite_ cluster's current
  `PrometheusAdapter` configuration with its own custom metrics configuration.
  Take copy of the existing `configMap` before install, if that matters:
  ```console
  ns=monitoring;
  cm=prom-adapter-prometheus-adapter;
  kubectl -n $ns get configmap/$cm -o yaml > adapter-config.yaml
  ```
- `PrometheusAdapter` needs to be restarted after install, for it to read the new configuration:
  ```console
  ns=monitoring;
  kubectl -n $ns delete $(kubectl -n $ns get pod --selector app.kubernetes.io/name=prometheus-adapter -o name)
  ```
- By default Prometheus adds [k8s RBAC rules](https://github.com/prometheus-operator/kube-prometheus/blob/main/manifests/prometheus-roleBindingSpecificNamespaces.yaml)
  for accessing metrics from `default`, `kube-system` and `monitoring` namespaces. If Helm is
  asked to install OPEA services to some other namespace, those rules need to be updated accordingly
- Unless pod resource requests, affinity rules and/or cluster NRI policies are used to better isolated
  service inferencing pods from each other, scaled up instances may never get to ready state
- Current HPA rules are examples for Xeon, for efficient scaling they need to be fine-tuned for given setup
  performance (underlying HW, used models and data types, OPEA version etc)

## Enable

ChatQnA includes pre-configured values files for scaling the services.

To enable HPA, add `-f chatqna/hpa-values.yaml` option to your `helm install` command line.

If **CPU** versions of TGI (and TEI) services are being scaled, resource requests and probe timings
suitable for CPU usage need to be used. Add `-f chatqna/cpu-values.yaml` option to your `helm install`
line.

### Verify

To verify that horizontalPodAutoscaler options work, it's better to check that both inferencing
services metrics, and HPA rules using custom metrics generated from them work.

Use k8s object names matching your Prometheus installation:

```console
prom_svc=prom-kube-prometheus-stack-prometheus # Metrics service
prom_ns=monitoring;                            # Prometheus namespace
```

Verify Prometheus found OPEA services metric endpoints, i.e. last number on `curl` output is non-zero:

```console
chart=chatqna; # OPEA services prefix
prom_url=http://$(kubectl -n $prom_ns get -o jsonpath="{.spec.clusterIP}:{.spec.ports[0].port}" svc/$prom_svc);
curl --no-progress-meter $prom_url/metrics | grep scrape_pool_targets.*$chart
```

**NOTE**: TGI and TEI inferencing services provide metrics endpoint only after they've processed their first request!

PrometheusAdapter lists TGI and/or TGI custom metrics (`te_*` / `tgi_*`):

```console
kubectl get --raw /apis/custom.metrics.k8s.io/v1beta1 | jq .resources[].name
```

HPA rules list valid (not `<unknown>`) TARGET values for service deployments:

```console
ns=default;  # OPEA namespace
kubectl -n $ns get hpa
```
