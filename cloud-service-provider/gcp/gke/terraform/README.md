# OPEA applications Google Cloud GKE deployment guide

This guide shows how to deploy OPEA applications on Google Cloud Platform (GCP) Google Kubernetes Engine (GKE) using Terraform.

## Prerequisites

- Access to GCP GKE
- [Terraform](https://developer.hashicorp.com/terraform/tutorials/aws-get-started/install-cli), [GCP CLI](https://cloud.google.com/sdk/gcloud) and [Helm](https://helm.sh/docs/helm/helm_install/) installed on your local machine.

## Setup

The setup uses Terraform to create GKE cluster with the following properties:

- 1-node GKE cluster with 50 GB disk and `m7i.x8large` SPOT instance (16 vCPU and 32 GB memory)
- Cluster autoscaling up to 10 nodes
- Storage Class (SC) `efs-sc` and Persistent Volume Claim (PVC) `model-volume` for storing the model data
- `LoadBalancer` address type for the service for external consumption
- Updates the kubeconfig file for `kubectl` access

Initialize the Terraform environment.

```bash
terraform init
```

## GKE cluster

By default, 1-node cluster is created which is suitable for running the OPEA application. See `variables.tf` and `opea-<application-name>.tfvars` if you want to tune the cluster properties, e.g., number of nodes, instance types or disk size.

## Persistent Volume Claim

OPEA needs a volume where to store the model. For that we need to create Kubernetes Persistent Volume Claim (PVC). OPEA requires `ReadWriteMany` option since multiple pods needs access to the storage and they can be on different nodes. On GKE, only EFS supports `ReadWriteMany`. Thus, each OPEA application below uses the file `GKE-efs-csi-pvc.yaml` to create PVC in its namespace.

## OPEA Applications

### ChatQnA

Use the commands below to create GKE cluster.

```bash
terraform plan --var-file opea-chatqna.tfvars -out opea-chatqna.plan
terraform apply "opea-chatqna.plan"
```

Once the cluster is ready, the kubeconfig file to access the new cluster is updated automatically. By default, the file is `~/.kube/config`.

Now you should have access to the cluster via the `kubectl` command.

Deploy ChatQnA Application with Helm

```bash
helm install -n chatqna --create-namespace chatqna oci://ghcr.io/opea-project/charts/chatqna --set service.type=LoadBalancer --set global.modelUsePVC=model-volume --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN}
```

Create the PVC as mentioned [above](#-persistent-volume-claim)

```bash
kubectl apply -f GKE-efs-csi-pvc.yaml -n chatqna
```

After a while, the OPEA application should be running. You can check the status via `kubectl`.

```bash
kubectl get pod -n chatqna
```

You can now start using the OPEA application.

```bash
OPEA_SERVICE=$(kubectl get svc -n chatqna chatqna -ojsonpath='{.status.loadBalancer.ingress[0].hostname}')
curl http://${OPEA_SERVICE}:8888/v1/chatqna \
    -H "Content-Type: application/json" \
    -d '{"messages": "What is the revenue of Nike in 2023?"}'
```

Cleanup

Delete the cluster via the following command.

```bash
helm uninstall -n chatqna chatqna
terraform destroy -var-file opea-chatqna.tfvars
```
