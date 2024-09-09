# HorizontalPodAutoscaler (HPA) support

## Table of Contents

- [Introduction](#introduction)
- [Pre-conditions](#pre-conditions)
  - [Resource requests](#resource-requests)
  - [Prometheus](#prometheus)
- [Gotchas](#gotchas)
- [Enable HPA](#enable-hpa)
  - [Install](#install)
  - [Post-install](#post-install)
- [Verify](#verify)

## Introduction

`horizontalPodAutoscaler` option enables HPA scaling for the TGI and TEI inferencing deployments:
https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/

Autoscaling is based on custom application metrics provided through [Prometheus](https://prometheus.io/).

## Pre-conditions

Read [post-install](#post-install) steps before installation!

### Resource requests

HPA controlled CPU pods SHOULD have appropriate resource requests or affinity rules (enabled in their
subcharts and tested to work) so that k8s scheduler does not schedule too many of them on the same
node(s). Otherwise they never reach ready state.

If you use different models than the default ones, update TGI and TEI resource requests to match
model requirements.

Too large requests would not be a problem as long as pods still fit to available nodes. However,
unless rules have been added to pods preventing them from being scheduled on same nodes, too
small requests would be an issue:

- Multiple inferencing instances interfere / slow down each other, especially if there are no
  [NRI policies](https://github.com/opea-project/GenAIEval/tree/main/doc/platform-optimization)
  that provide further isolation
- Containers can become non-functional when their actual resource usage crosses the specified limits

### Prometheus

If cluster does not run [Prometheus operator](https://github.com/prometheus-operator/kube-prometheus)
yet, it SHOULD be be installed before enabling HPA, e.g. by using a Helm chart for it:
https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack

Prometheus-adapter is also needed, to provide k8s custom metrics based on collected TGI / TEI metrics:
https://github.com/prometheus-community/helm-charts/tree/main/charts/prometheus-adapter

To install (older versions) of them:

```console
$ ns=monitoring
$ helm install prometheus-stack prometheus-community/kube-prometheus-stack --version 55.5.2 -n $ns
$ kubectl get services -n $ns
$ helm install  prometheus-adapter prometheus-community/prometheus-adapter --version 4.10.0 -n $ns \
  --set prometheus.url=http://prometheus-stack-kube-prom-prometheus.$ns.svc \
  --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false
```

NOTE: the service name given above in `prometheus.url` must match the listed Prometheus service name,
otherwise adapter cannot access it!

(Alternative for setting the above `prometheusSpec` variable to `false` is making sure that
`prometheusRelease` value in top-level chart matches the release name given to the Prometheus
install i.e. when it differs from `prometheus-stack` used above.  That is used to annotate
created serviceMonitors with a label Prometheus requires when above option is `true`.)

## Gotchas

Why HPA is opt-in:

- Installing custom metrics for HPA requires manual post-install steps, as
  Prometheus-operator and -adapter are missing support needed to automate that
- Top level chart name needs to conform to Prometheus metric naming conventions,
  as it is also used as a metric name prefix (with dashes converted to underscores)
- By default Prometheus adds [k8s RBAC rules](https://github.com/prometheus-operator/kube-prometheus/blob/main/manifests/prometheus-roleBindingSpecificNamespaces.yaml)
  for accessing metrics from `default`, `kube-system` and `monitoring` namespaces. If Helm is
  asked to install OPEA services to some other namespace, those rules need to be updated accordingly
- Unless pod resource requests, affinity rules, scheduling topology constraints and/or cluster NRI
  policies are used to better isolate service inferencing pods from each other, instances
  scaled up on same node may never get to ready state
- Current HPA rules are just examples, for efficient scaling they need to be fine-tuned for given setup
  performance (underlying HW, used models and data types, OPEA version etc)
- Debugging missing custom metric issues is hard as logs rarely include anything helpful

## Enable HPA

### Install

ChatQnA includes pre-configured values files for scaling the services.

To enable HPA, add `-f chatqna/hpa-values.yaml` option to your `helm install` command line.

If **CPU** versions of TGI (and TEI) services are being scaled, resource requests and probe timings
suitable for CPU usage need to be used. Add `-f chatqna/cpu-values.yaml` option to your `helm install`
line.  If you need to change model specified there, update the resource requests accordingly.

### Post-install

Above step created custom metrics config for Prometheus-adapter suitable for HPA use.

Take backup of existing custom metrics config before replacing it:

```console
$ prom_ns=monitoring  # Prometheus/-adapter namespace
$ name=$(kubectl -n $prom_ns get cm --selector app.kubernetes.io/name=prometheus-adapter -o name | cut -d/ -f2)
$ kubectl -n $prom_ns get cm/$name -o yaml > adapter-config.yaml.bak
```

Save generated config with values matching current adapter config:

```console
$ chart=chatqna  # OPEA chart release name
$ kubectl get cm/$chart-custom-metrics -o yaml | sed \
  -e "s/name:.*custom-metrics$/name: $name/" \
  -e "s/namespace: default$/namespace: $prom_ns/" \
  > adapter-config.yaml
```

NOTE: if there are existing custom metric rules you need to retain, add them from saved
`adapter-config.yaml.bak` to `adapter-config.yaml` file now!

Overwrite current Prometheus-adapter configMap with generated one:

```console
$ kubectl delete -n $prom_ns cm/$name
$ kubectl apply -f adapter-config.yaml
```

And restart it, so that it will use the new config:

```console
$ kubectl -n $ns delete $(kubectl -n $ns get pod --selector app.kubernetes.io/name=prometheus-adapter -o name)
```

## Verify

To verify that horizontalPodAutoscaler options work, it's better to check that both metrics
from the inferencing services, and HPA rules using custom metrics generated from them, do work.

(Names of the object names depend on whether Prometheus was installed from manifests, or Helm,
and the release name given for its Helm install.)

Check installed Prometheus service names:

```console
$ prom_ns=monitoring  # Prometheus/-adapter namespace
$ kubectl -n $prom_ns get svc
```

Use service name matching your Prometheus installation:

```console
$ prom_svc=prometheus-stack-kube-prom-prometheus  # Metrics service
```

Verify Prometheus found metric endpoints for chart services, i.e. last number on `curl` output is non-zero:

```console
$ chart=chatqna # OPEA chart release name
$ prom_url=http://$(kubectl -n $prom_ns get -o jsonpath="{.spec.clusterIP}:{.spec.ports[0].port}" svc/$prom_svc)
$ curl --no-progress-meter $prom_url/metrics | grep scrape_pool_targets.*$chart
```

**NOTE**: TGI and TEI inferencing services provide metrics endpoint only after they've processed
their first request, and reranking service will be used only after context data has been uploaded!

Check that both Prometheus metrics required from TGI are available:

```console
$ for m in sum count; do
  curl --no-progress-meter $prom_url/api/v1/query? \
  --data-urlencode query=tgi_request_inference_duration_$m'{service="'$chart'-tgi"}' | jq;
done | grep __name__
```

PrometheusAdapter lists corresponding TGI and/or TEI custom metrics (prefixed with chart name):

```console
$ kubectl get --raw /apis/custom.metrics.k8s.io/v1beta1 | jq .resources[].name
```

And HPA rules have TARGET values for HPA controlled service deployments (instead of `<unknown>`):

```console
$ ns=default  # OPEA namespace
$ kubectl -n $ns get hpa
```
