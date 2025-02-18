# Dashboards

Helm chart for installing Grafana dashboard(s) for OPEA application(s).

## Installing the Chart

After cluster [monitoring support](../../monitoring.md) with Prometheus + Grafana,
and some OPEA application(s) with monitoring enabled (`--set global.monitoring=true`)
have been [installed](../../README.md) to the cluster, Grafana dashboard(s) for
those OPEA application(s) can be installed with:

```console
cd GenAIInfra/helm-charts/common
PROM_RELEASE=prometheus-stack   # release name for Prometheus/Grafana Helm install
PROM_NAMESPACE=monitoring       # namespace where they are installed
helm install opea-dashboard dashboard/ --set global.prometheusNamespace=PROM_NAMESPACE --set global.prometheusRelease=$PROM_RELEASE
```

## Values

| Key                       | Type   | Default            | Description                                                                                             |
| ------------------------- | ------ | ------------------ | ------------------------------------------------------------------------------------------------------- |
| prefix                    | string | `OPEA application` | Title prefix for the installed dashboards                                                               |
| metrics                   | bool   | `true`             | Whether to install metrics dashboard for the application services                                       |
| scaling                   | bool   | `false`            | Whether to install scaling dashboard for the application services scaling, use with [HPA](../../HPA.md) |
| global.promeheusNamespace | string | `monitoring`       | Namespace where Prometheus/Grafana are installed, needed to for Grafana to pick up installed dashboard  |
| global.promeheusRelease   | string | `prometheus-stack` | Release name for Prometheus/Grafana Helm install, needed to for Grafana to pick up installed dashboard  |
