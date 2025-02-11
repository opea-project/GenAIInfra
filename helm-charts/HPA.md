# HorizontalPodAutoscaler (HPA) support

## Table of Contents

- [Introduction](#introduction)
- [Pre-conditions](#pre-conditions)
  - [Resource requests](#resource-requests)
  - [Prometheus metrics](#prometheus-metrics)
  - [Prometheus-adapter](#prometheus-adapter)
- [Gotchas](#gotchas)
- [Enable HPA](#enable-hpa)
  - [Install](#install)
  - [Post-install](#post-install)
- [Verify](#verify)

## Introduction

`autoscaling` option enables HPA scaling for relevant service components:
https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/

Autoscaling is based on custom application metrics provided through [Prometheus](https://prometheus.io/).

## Pre-conditions

Read [post-install](#post-install) steps before installation!

### Resource requests

HPA controlled _CPU_ pods SHOULD have appropriate resource requests or affinity rules (enabled in their
subcharts and tested to work) so that k8s scheduler does not schedule too many of them on the same
node(s). Otherwise they never reach ready state.

If you use different models than the default ones, update inferencing services (vLLM, TGI, TEI) resource
requests to match model requirements.

Too large requests would not be a problem as long as pods still fit to available nodes. However,
unless rules have been added to pods preventing them from being scheduled on same nodes, too
small requests would be an issue:

- Multiple inferencing instances interfere / slow down each other, especially if there are no
  [NRI policies](https://github.com/opea-project/GenAIEval/tree/main/doc/platform-optimization/README.md)
  that provide further isolation
- Containers can become non-functional when their actual resource usage crosses the specified limits

### Prometheus metrics

Autoscaling requires k8s Prometheus installation and monitoring to be enabled in the top level chart.
See [monitoring instructions](monitoring.md) for details.

### Prometheus-adapter

Prometheus-adapter is also needed, to provide k8s custom metrics based on collected service metrics:
https://github.com/prometheus-community/helm-charts/tree/main/charts/prometheus-adapter

Install adapter after installing Prometheus:

```console
$ prom_ns=monitoring  # namespace for Prometheus/-adapter
$ kubectl get svc -n $prom_ns
$ helm install  prometheus-adapter prometheus-community/prometheus-adapter --version 4.10.0 -n $prom_ns \
  --set prometheus.url=http://prometheus-stack-kube-prom-prometheus.$prom_ns.svc \
  --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false
```

NOTE: the service name given above in `prometheus.url` must match the listed Prometheus service name,
otherwise adapter cannot access it!

(Alternative for setting the above `prometheusSpec` variable to `false` is making sure that
`prometheusRelease` value in top-level chart matches the release name given to the Prometheus
install i.e. when it differs from `prometheus-stack` used above. That is used to annotate
created serviceMonitors with a label Prometheus requires when above option is `true`.)

## Gotchas

Why HPA is opt-in:

- Installing custom metrics for HPA requires manual post-install steps, as
  Prometheus-operator and -adapter are missing support needed to automate that
- Top level chart name needs to conform to Prometheus metric naming conventions,
  as it is also used as a metric name prefix (with dashes converted to underscores)
- Unless pod resource requests, affinity rules, scheduling topology constraints and/or cluster NRI
  policies are used to better isolate _CPU_ inferencing pods from each other, service instances
  scaled up on same node may never get to ready state
- Current HPA rules are just examples, for efficient scaling they need to be fine-tuned for given setup
  performance (underlying HW, used models and data types, OPEA version etc)
- Debugging missing custom metric issues is hard as logs rarely include anything helpful

## Enable HPA

### Install

ChatQnA includes pre-configured values files for scaling the services.

To enable HPA, add `-f chatqna/hpa-values.yaml` option to your `helm install` command line.

If **CPU** versions of TGI (and TEI) services are being scaled, resource requests and probe timings
suitable for CPU usage need to be used. `chatqna/cpu-values.yaml` provides example of such constraints
which can be added (with `-f` option) to your Helm install. As those values depend on the underlying HW,
used model, data type and image versions, the specified resource values may need to be updated.

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
$ selector=app.kubernetes.io/name=prometheus-adapter
$ kubectl -n $prom_ns delete $(kubectl -n $prom_ns get pod --selector $selector -o name)
```

## Verify

After [verifying that service metrics work](monitoring.md#verify),
one can verify that HPA rules can access custom metrics based on them.

Verify that there are custom metrics from inferencing service(s), prefixed with the chart name:

```console
$ kubectl get --raw /apis/custom.metrics.k8s.io/v1beta1 | jq .resources[].name
```

And HPA rules have TARGET values for HPA controlled service deployments (instead of `<unknown>`):

```console
$ ns=default  # OPEA namespace
$ kubectl -n $ns get hpa
```

**NOTE**: inferencing services provide metrics only after they've processed their first request.
And reranking service is used only after query context data has been uploaded!
