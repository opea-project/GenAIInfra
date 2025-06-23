# HorizontalPodAutoscaler (HPA) support

## Table of Contents

- [Introduction](#introduction)
- [Pre-conditions](#pre-conditions)
  - [Resource requests](#resource-requests)
  - [Prometheus metrics](#prometheus-metrics)
  - [Prometheus adapter](#prometheus-adapter)
- [Gotchas](#gotchas)
- [Enable HPA](#enable-hpa)
  - [Install](#install)
  - [Post-install](#post-install)
- [Verify](#verify)
- [Scaling metric considerations](#scaling-metric-considerations)
  - [Autoscaling principles](#autoscaling-principles)
  - [Current scaling metrics](#current-scaling-metrics)
  - [Other potential metrics](#other-potential-metrics)

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

### Prometheus adapter

Prometheus adapter is also needed, to provide k8s custom metrics based on collected service metrics:
https://github.com/prometheus-community/helm-charts/tree/main/charts/prometheus-adapter

Install adapter after installing Prometheus:

```console
$ prom_ns=monitoring  # namespace for Prometheus+adapter
$ kubectl get svc -n $prom_ns
$ helm install  prometheus-adapter prometheus-community/prometheus-adapter --version 4.10.0 -n $prom_ns \
  --set prometheus.url=http://prometheus-stack-kube-prom-prometheus.$prom_ns.svc \
  --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false
```

> **NOTE**: the service name given above in `prometheus.url` must match the listed Prometheus
> service name, otherwise adapter cannot access it!

(Alternative for setting the above `prometheusSpec` variable to `false` is making sure that
`prometheusRelease` value in top-level chart matches the release name given to the Prometheus
install i.e. when it differs from `prometheus-stack` used above. That is used to annotate
created serviceMonitors with a label Prometheus requires when above option is `true`.)

## Gotchas

Why HPA is opt-in:

- Installing custom metrics for HPA requires manual post-install steps, as
  both Prometheus operator and adapter are missing support needed to automate that
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

The above step created a custom metrics configuration for the Prometheus adapter, suitable for HPA use.

However, Helm does not allow OPEA chart to overwrite adapter configMap, as it belongs to another chart.
Therefore, a manual step is needed to overwrite its current custom metric rules.

The following will overwrite the current adapter custom metric rules with the ones generated by OPEA
Helm and will restart the adapter to apply the new rules.

```console
scripts/install-custom-metrics.sh monitoring chatqna
```

(It assumes adapter to be in the `monitoring` namespace, and the new rules are expected to be
generated by a ChatQnA chart release named `chatqna`.)

YAML backups of the new and previous rules are saved to the current directory.

## Verify

After [verifying that service metrics work](monitoring.md#verify),
one can verify that HPA rules can access custom metrics based on them.

Verify that custom metric values are available for scaling the services:

```console
watch -n 5 scale-monitor-helm.sh default chatqna
```

(Assumes that HPA scaled chart is installed to `default` namespace with `chatqna` release name.)

> **NOTE**: inferencing services provide metrics only after they've processed their first request.
> The reranking service is used only after the query context data has been uploaded. Until then,
> no metrics will be available for them.

## Scaling metric considerations

### Autoscaling principles

The used model, underlying HW and engine parameters are supposed to be selected so that engine
instance can satisfy service SLA (Service Level Agreement) requirements for its own requests,
also when it's becoming saturated. Autoscaling is then intended to scale up the service so that
requests can be directed to unsaturated instances.

Problem is finding a good metric, and its threshold, for indicatating this saturation point.
Preferably it should be something that can anticipate this point, so that startup delay for
the new engine instances does not cause SLA breakage (or in the worst case requests being
rejected, if the engine queue fills up).

> **NOTE**: Another problem is Kubernetes service routing sending requests (also) to already saturated
> instances, instead of idle ones. Using [KubeAI](../kubeai/#readme) (instead of HPA) to manage
> both engine scaling + query routing can solve that.

### Current scaling metrics

The following inference engine metrics are used to autoscale their replica counts:

- vLLM: Active requests i.e. count of waiting (queued) + (already) running requests
  - Good overall scaling metric, used also by [KubeAI](../kubeai/#readme) for scaling vLLM
  - Threshold depends on how many requests underlying HW / engine config can process for given model in parallel
- TGI / TEI: Queue size, i.e. how many requests are waiting to be processed
  - Used because TGI and TEI do not offer metric for (already) running requests, just waiting ones
  - Independent of the used model, so works well as an example, but not that good for production because
    scaling happens late and fluctuates a lot (due to metric being zero when engine is not saturated)

### Other potential metrics

All the metrics provided by the inference engines are listed in their documentation:

- [vLLM metrics](https://docs.vllm.ai/en/v0.8.5/serving/metrics.html)
  - [Metric design](https://docs.vllm.ai/en/v0.8.5/design/v1/metrics.html)
- [TGI metrics](https://huggingface.co/docs/text-generation-inference/en/reference/metrics)
  - TEI (embed and reranking) services provide a subset of these TGI metrics

OPEA application [dashboard](monitoring.md#dashboards) provides (Prometheus query) examples
for deriving service performance metrics out of engine Histogram metrics.

Their suitability for autoscaling:

- Request latency, request per second (RPS) - not suitable
  - Depends completely on input and output token counts and is an indicator for past performance, not incoming load
- First token latency (TTFT) - potential
  - Relevancy depends on use-case; number of used tokens and what's important
- Next token latency (TPOT, ITL), tokens per second (TPS) - potential
  - Relevancy depends on use-case; number of used tokens and what's important

Performance metrics will be capped by the performance of the underlying engine setup.
Beyond a certain point, they no longer reflect the actual incoming load or indicate how
much scaling is needed.

Therefore such metrics could be used in production _when_ their thresholds are carefully
fine-tuned and rechecked every time underlying setup (model, HW, engine config) changes.
In OPEA Helm charts that setup is user selectable, so such metrics are unsuitable for
autoscaling examples.

(General [explanation](https://docs.nvidia.com/nim/benchmarking/llm/latest/metrics.html) on how these metrics are measured.)
