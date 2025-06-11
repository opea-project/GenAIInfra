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
      version = ">= 2.17.0"
    }
  }
}

# Configure the Kubernetes provider to use the IBM Kubernetes Service cluster
provider "kubernetes" {
  config_path = "~/.kube/ibm-iks-kubeconfig" # Path for IBM Kubernetes Service config
}

# Configure the Helm provider
provider "helm" {
  kubernetes {
    config_path = "~/.kube/ibm-iks-kubeconfig"
  }
}

variable "namespaces" {
  type    = set(string)
  default = ["chatqna", "codegen", "docsum", "ui", "chathistory", "keycloak", "nginx", "prompt"]
}

# Create namespaces first
resource "kubernetes_namespace" "namespaces" {
  for_each = var.namespaces

  metadata {
    name = each.value
  }
}

# Then create secrets with dependency
resource "kubernetes_secret" "regcred" {
  for_each = var.namespaces

  depends_on = [kubernetes_namespace.namespaces]

  metadata {
    name      = "regcred"
    namespace = each.value
  }

  type = "kubernetes.io/dockerconfigjson"

  data = {
    ".dockerconfigjson" = jsonencode({
      "auths" : {
        "us.icr.io" : {
          "username" : "iamapikey",
          "password" : var.ibmcloud_api_key,
          "email" : var.email,
          "auth" : base64encode("iamapikey:${var.ibmcloud_api_key}")
        }
      }
    })
  }
}

