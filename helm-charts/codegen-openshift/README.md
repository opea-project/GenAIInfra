# CodeGen

Helm chart for deploying CodeGen service on Red Hat OpenShift.

# Prerequisites

1. Red Hat OpenShift Cluster with dynamic StorageClass to provision PersistentVolumes e.g. OpenShift Data Foundation)
2. Image registry to push there docker images (https://docs.openshift.com/container-platform/4.16/registry/securing-exposing-registry.html).
3. Account on https://huggingface.co/, access to model _ise-uiuc/Magicoder-S-DS-6.7B_ (for Xeon) or _meta-llama/CodeLlama-7b-hf_ (for Gaudi) and token with Read permissions.

## Installing the Chart

To install the chart, login to OpenShift CLI, go to your project and run the following:

```console
cd GenAIInfra/helm-charts/
./update_dependency.sh
helm dependency update codegen-openshift

export NAMESPACE="insert-your-namespace-here"
export CLUSTERDOMAIN="$(oc get Ingress.config.openshift.io/cluster -o jsonpath='{.spec.domain}' | sed 's/^apps.//')"
export HFTOKEN="insert-your-huggingface-token-here"

# To run on Xeon
helm install codegen codegen-openshift --set image.repository=image-registry.openshift-image-registry.svc:5000/${NAMESPACE}/codegen --set llm-uservice.image.repository=image-registry.openshift-image-registry.svc:5000/${NAMESPACE}/llm-tgi --set react-ui.image.repository=image-registry.openshift-image-registry.svc:5000/${NAMESPACE}/react-ui --set global.clusterDomain=${CLUSTERDOMAIN} --set global.huggingfacehubApiToken=${HFTOKEN}

# To run on Gaudi
helm install codegen codegen-openshift --set image.repository=image-registry.openshift-image-registry.svc:5000/${NAMESPACE}/codegen --set llm-uservice.image.repository=image-registry.openshift-image-registry.svc:5000/${NAMESPACE}/llm-tgi --set react-ui.image.repository=image-registry.openshift-image-registry.svc:5000/${NAMESPACE}/react-ui --set global.clusterDomain=${CLUSTERDOMAIN} --set global.huggingfacehubApiToken=${HFTOKEN} --values codegen-openshift/gaudi-values.yaml
```

## Verify

To verify the installation, run the command `oc get pods` to make sure all pods are running. Wait about 5 minutes for building images. When 3 pods achieve _Completed_ status, the rest with services should go to _Running_.

## Launch the UI

To access the frontend, find the route for _react-ui_ with command `oc get routes` and open it in the browser.
