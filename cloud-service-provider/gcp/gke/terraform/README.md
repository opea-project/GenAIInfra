# OPEA applications GCP GKE deployment guide

This guide shows how to deploy OPEA applications on Google Cloud Platform (GCP) Google Kubernetes Engine (GKE) using Terraform.

## Prerequisites

- Access to GCP GKE
- [Terraform](https://developer.hashicorp.com/terraform/tutorials/gcp-get-started/install-cli), [GCP CLI](https://cloud.google.com/sdk/docs/install-sdk) and [Helm](https://helm.sh/docs/helm/helm_install/),[kubectl](https://kubernetes.io/docs/tasks/tools/) installed on your local machine.

## Setup

The setup uses Terraform to create GKE cluster with the following properties:

- 1-node GKE cluster with 100 GB disk and `n4-standard-8` preemptible SPOT instance (8 vCPU and 32 GB memory)
- Cluster autoscaling up to 5 nodes

Pre GKE Cluster setup

- After you've installed the gcloud SDK, initialize it by running the following command.

```bash
gcloud init
```

- This will authorize the SDK to access GCP using your user account credentials and add the SDK to your PATH. This steps requires you to login and select the project you want to work in. Finally, add your account to the Application Default Credentials (ADC). This will allow Terraform to access these credentials to provision resources on GCloud.

```bash
gcloud auth application-default login
```

In here, you will find four files used to provision a VPC, subnets and a GKE cluster.

- vpc.tf provisions a VPC and subnet. A new VPC is created for this tutorial so it doesn't impact your existing cloud environment and resources. This file outputs region.

- main.tf provisions a GKE cluster and a separately managed node pool (recommended). Separately managed node pools allows you to customize your Kubernetes cluster profile — this is useful if some Pods require more resources than others. You can learn more here. The number of nodes in the node pool is defined also defined here.

- opea-chatqna.tfvars is a template for the project_id and region variables.

- versions.tf sets the Terraform version to at least 0.14.

## Update your opea-chatqna.tfvars file

Replace the values in your opea-chatqna.tfvars file with your project_id and region. Terraform will use these values to target your project when provisioning your resources. Your opea-chatqna.tfvars file should look like the following.

```bash
 # opea-chatqna.tfvars
  project_id = "REPLACE_ME"
  region     = "us-central1"
```

You can find the project your gcloud is configured to with this command.

```bash
 gcloud config get-value project
```

The region has been defaulted to us-central1; you can find a full list of gcloud regions - https://cloud.google.com/compute/docs/regions-zones

Initialize the Terraform environment.

```bash
terraform init
```

## GKE cluster

By default, 1-node cluster is created which is suitable for running the OPEA application. See `main.tf` upto max_node_count = 5, if you want to tune the cluster properties, e.g., number of nodes, instance types or disk size.

## Persistent Volume Claim

OPEA needs a volume where to store the model. For that we need to create Kubernetes Persistent Volume Claim (PVC). OPEA requires `ReadWriteOnce` option since multiple pods needs access to the storage and they can be on different nodes. On GKE, We are installing Storage Class that support n4-standard-8 which is hyper-balanced . Thus, each OPEA application below uses the file `eks-fs-pvc.yaml` to create Storage Class and PVC in its namespace.

## OPEA Applications

### ChatQnA

Use the commands below to create GKE cluster.

```bash
terraform plan --var-file opea-chatqna.tfvars -out opea-chatqna.plan
terraform apply "opea-chatqna.plan"
```

Once the cluster is ready, update kubectl config

```bash
gcloud container clusters get-credentials "project_id"-gke --region us-central1 --project "project_id"
```

Now you should have access to the cluster via the `kubectl` command.

Deploy ChatQnA Application with Helm

```bash
helm install -n chatqna --create-namespace chatqna oci://ghcr.io/opea-project/charts/chatqna --set service.type=LoadBalancer --set global.modelUsePVC=model-volume --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN}
```

Create the Storage Class and PVC as mentioned [above](#-persistent-volume-claim)

```bash
kubectl apply -f gke-fs-pvc.yaml -n chatqna
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
