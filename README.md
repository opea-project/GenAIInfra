# OPEA Helm Charts

## Usage

[Helm](https://helm.sh) must be installed to use the charts.
Please refer to Helm's [documentation](https://helm.sh/docs/) to get started.

Once Helm is set up properly, add the repo as follows:

```console
helm repo add opea https://opea-project.github.io/GenAIInfra
```

You can then run `helm search repo opea` to see the charts.


If you had already added this repo earlier, run `helm repo update` to retrieve
the latest versions of the packages.  You can then run `helm search repo
opea` to see the charts.

To install the <chart-name> chart:

    helm install my-<chart-name> opea/<chart-name>

To uninstall the chart:

    helm delete my-<chart-name>

For more information and options of the helm charts, see [GenAIInfra helm charts](https://github.com/opea-project/GenAIInfra/blob/main/helm-charts/README.md)
