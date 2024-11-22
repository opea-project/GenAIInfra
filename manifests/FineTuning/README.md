# Deploy the KubeRay operator with the Helm chart

```bash
helm repo add kuberay https://ray-project.github.io/kuberay-helm/
helm repo update

# Install both CRDs and KubeRay operator v1.1.1.
helm install kuberay-operator kuberay/kuberay-operator --version 1.1.1

# Confirm that the operator is running in the namespace `default`.
kubectl get pods
# NAME                                READY   STATUS    RESTARTS   AGE
# kuberay-operator-7fbdbf8c89-pt8bk   1/1     Running   0          27s
```

# Deploy On Xeon

```bash
cd GenAIInfra/manifests/FineTuning/xeon
kubectl apply -f finetuning-ray-cluster-autoscaler.yaml
kubectl apply -f finetuning.yaml
```

# Deploy On Gaudi

TBD

# Verify LLM Fine-tuning Service

Make sure all the pods are running.

```bash
kubectl get pods
```
