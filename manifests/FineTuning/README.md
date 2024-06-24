## Deploy On Xeon

```bash
cd GenAIInfra/manifests/FineTuning/xeon
kubectl apply -f finetuning-ray-cluster-autoscaler.yaml
kubectl apply -f finetuning.yaml
```

## Deploy On Gaudi

TBD

## Verify llm Services

Make sure all the pods are running.

```bash
kubectl get pods
```
