# providers.tf

terraform {
  required_version = ">= 1.0.0"

  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.0"
    }
  }
}

# Configure the Kubernetes provider to use the local Minikube cluster
provider "kubernetes" {
  config_path = "~/.kube/config" # Default path for kubectl config
}

# Configure the Helm provider
provider "helm" {
  kubernetes {
    config_path = "~/.kube/config"
  }
}

