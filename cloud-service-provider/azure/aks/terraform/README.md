# OPEA applications Azure AKS deployment guide

This guide shows how to deploy OPEA applications on Azure Kubernetes Service (AKS) using Terraform.

## Prerequisites

- Access to Azure AKS
- [Terraform](https://developer.hashicorp.com/terraform/tutorials/azure-get-started/install-cli), [Azure CLI](https://learn.microsoft.com/en-us/cli/azure/) and [Helm](https://helm.sh/docs/helm/helm_install/) installed on your local machine.
- Keep the Azure subscription handy and enter the subscription id when prompted during the terraform execution.

## Setup

The setup uses Terraform to create AKS cluster with the following properties:

- 1-node AKS cluster with 50 GB disk and `Standard_D32d_v5` SPOT (or standard based on the application variables) instance (16 vCPU and 32 GB memory)
- Cluster autoscaling up to 10 nodes
- Storage Class (SC) `azurefile-csi` and Persistent Volume Claim (PVC) `model-volume` for storing the model data

Initialize the Terraform environment.

```bash
terraform init
```

## AKS cluster

By default, 1-node cluster is created which is suitable for running the OPEA application. See `variables.tf` and `opea-<application-name>.tfvars` if you want to tune the cluster properties, e.g., number of nodes, instance types or disk size.

## Cosmos DB

By default Cosmos DB will not be provisioned. If you want Cosmos DB as part of your resource provisioning, update `is_cosmosdb_required` property present in `opea-<application-name>.tfvars` to `true`.

## Persistent Volume Claim

OPEA needs a volume where to store the model. For that we need to create Kubernetes Persistent Volume Claim (PVC). OPEA requires `ReadWriteMany` option since multiple pods needs access to the storage and they can be on different nodes. On AKS, only Azure File Service supports `ReadWriteMany`. Thus, each OPEA application below uses the file `aks-azfs-csi-pvc.yaml` to create PVC in its namespace.

## OPEA Applications

### ChatQnA

Use the commands below to create AKS cluster.
User has to input their Azure subscription id while running the following commands when prompted.

```bash
terraform plan --var-file opea-chatqna.tfvars -out opea-chatqna.plan
terraform apply "opea-chatqna.plan"
```

Once the cluster is ready, the kubeconfig file to access the new cluster is updated automatically. By default, the file is `~/.kube/config`.

Now you should have access to the cluster via the `kubectl` command.

Deploy ChatQnA Application with Helm

```bash
helm install -n chatqna --create-namespace chatqna oci://ghcr.io/opea-project/charts/chatqna --set service.type=LoadBalancer --set global.modelUsePVC=model-volume --set global.HF_TOKEN=${HFTOKEN}
```

Create the PVC as mentioned [above](#-persistent-volume-claim)

```bash
kubectl apply -f aks-azfs-csi-pvc.yaml -n chatqna
```

After a while, the OPEA application should be running. You can check the status via `kubectl`.

```bash
kubectl get pod -n chatqna
```

Ensure that all pods are running.
You can now start using the OPEA application.

```bash
OPEA_SERVICE=$(kubectl get svc -n chatqna chatqna -ojsonpath='{.status.loadBalancer.ingress[0].hostname}')
curl http://${OPEA_SERVICE}:8888/v1/chatqna \
    -H "Content-Type: application/json" \
    -d '{"messages": "What is the revenue of Nike in 2023?"}'
```

Cleanup

Delete the cluster via the following command. User has to input their Azure subscription id while running the following commands when prompted.

```bash
helm uninstall -n chatqna chatqna
terraform destroy -var-file opea-chatqna.tfvars
```
