# Helm chart for genai-microservices-connector(GMC)

Helm chart for deploying the genai-microservices-connector(GMC) service on a Kubernetes cluster.

## Installing the GMC Helm Chart

To use this GMC helm chart, you need to have a Kubernetes cluster and helm installed. If you don't have helm(version >= 3.15) installed, you can follow the instructions [here](https://helm.sh/docs/intro/install/).

This helm chart will install the following components of GMC:

- GMC CRD
- GenAI Components and GMC Router manifests
- GMC Manager

**NOTE: Because helm doesn't support updating/deleting CRD, you need to manually delete the CRD before upgrading the GMC helm chart.**

**NOTE:**
Before installting the manifests, please replace your own huggingface tokens，Google API KEY and Google CSE ID in the manifests:

```sh
export YOUR_HF_TOKEN=<your hugging facetoken>
export YOUR_GOOGLE_API_KEY=<your google api key>
export YOUR_GOOGLE_CSE_ID=<your google cse id>
find helm/manifests_common/ -name '*.yaml' -type f -exec sed -i "s#insert-your-huggingface-token-here#$YOUR_HF_TOKEN#g" {} \;
find helm/manifests_common/ -name '*.yaml' -type f -exec sed -i "s#GOOGLE_API_KEY:.*#GOOGLE_API_KEY: "$YOUR_GOOGLE_API_KEY"#g" {} \;
find helm/manifests_common/ -name '*.yaml' -type f -exec sed -i "s#GOOGLE_CSE_ID:.*#GOOGLE_CSE_ID: "$YOUR_GOOGLE_CSE_ID"#g" {} \;
```

if you have pre-defined directory to save the models on you cluster hosts, please set the path to the manifests

```sh
export MOUNT_DIR=<your model path>
find helm/manifests_common/ -name '*.yaml' -type f -exec sed -i "s#path: /mnt/opea-models#path: $MOUNT_DIR#g" {} \;
```

**NOTE:**
GMC manager, GenAI components and GMC router manifests are deployed in any namespace. Here we use `system` as an examep:

```console
helm install -n system --create-namespace gmc helm
```

## Check the installation result

Run the command `kubectl get pod -n system` to make sure all pods are running:

```
NAME                            READY   STATUS    RESTARTS   AGE
gmc-contoller-8bcb9d469-l6fsj   1/1     Running   0          55s
```

## Next step

After the GMC is installed, you can follow the [GMC user guide](../usage_guide.md) for sample use cases.

## Uninstall

**Delete the instances (CRs) from the cluster if you have ever deployed the instances from GMC user guide:**

```sh
kubectl delete -k config/samples/
```

**UnDeploy the GMC manager and GenAI components and GMC router manifest from the cluster:**

```sh
helm delete -n system gmc
```

**Delete the APIs(CRDs) from the cluster:**

```sh
kubectl delete crd gmconnectors.gmc.opea.io
```
