# OPEA applications AWS EKS deployment guide

This guide shows how to deploy OPEA applications on Amazon Web Service (AWS) Elastic Kubernetes Service (EKS) using Terraform.

## Prerequisites

- Access to AWS EKS
- [Terraform](https://developer.hashicorp.com/terraform/tutorials/aws-get-started/install-cli), [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html) and [Helm](https://helm.sh/docs/helm/helm_install/) installed on your local machine.

## Setup

The setup uses Terraform to create EKS cluster with the following properties:

- 1-node EKS cluster with 50 GB disk and `m7i.x8large` SPOT instance (16 vCPU and 32 GB memory)
- Cluster autoscaling up to 10 nodes
- Storage Class (SC) `efs-sc` and Persistent Volume Claim (PVC) `model-volume` for storing the model data
- `LoadBalancer` address type for the service for external consumption
- Updates the kubeconfig file for `kubectl` access

Initialize the Terraform environment.

```bash
terraform init
```

Add OPEA Helm chart repository.

```bash
helm repo add opea https://opea-project.github.io/GenAIInfra
helm repo update
```

## Create EKS cluster

By default, 1-node EKS cluster is created which suitable for running an OPEA application. See `variables.tf` and `opea-<application-name>.tfvars` if you want to tune the EKS cluster properties, e.g., number of nodes, instance types or disk size.

## Create Persistent Volume Claim

OPEA needs a volume where to store the model. For that we need to create Kubernetes Persistent Volume Claim (PVC). OPEA requires `ReadWriteMany` option since multiple pods needs access to the storage and they can be on different nodes. On EKS, only EFS supports `ReadWriteMany`. Ensure `<namespace>` below matches the namespace you deploy the application to. It is better for create the PVC after you have deployed the OPEA application.

```bash
kubectl apply -f eks-efs-csi-pvc.yaml -n <namescape>
```

### ChatQnA

Use the commands below to create EKS cluster.

```bash
terraform plan --var-file opea-chatqna.tfvars -out opea-chatqna.plan
terraform apply "opea-chatqna.plan"
```

Once the cluster is ready, the kubeconfig file to access the new cluster is updated automatically. By default, the file is `~/.kube/config`.

Now your should have access to the cluster via `kubectl` command.

Deploy ChatQnA Application with Helm

```bash
helm install -n chatqna --create-namespace chatqna opea/chatqna --set service.type=LoadBalancer --set global.modelUsePVC=model-volume --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN}
```

After a while, the OPEA application should be running. You can check the status via kubectl.

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
