# Determine if helm_repo is a local path or remote URI
locals {
  is_remote = can(regex("^([a-zA-Z0-9.-]+)/([a-zA-Z0-9-]+)/([a-zA-Z0-9-]+)$", var.helm_repo))
  repo_parts = local.is_remote ? regex("^([a-zA-Z0-9.-]+)/([a-zA-Z0-9-]+)/([a-zA-Z0-9-]+)$", var.helm_repo) : []
}

resource "helm_release" "prompt-usvc" {
  name             = "prompt-usvc"
  chart      = local.is_remote ? local.repo_parts[3] : var.helm_repo
  repository = local.is_remote ? "oci://${local.repo_parts[1]}/${local.repo_parts[2]}" : null
  namespace        = "prompt"
  create_namespace = false
  timeout          = 600
  
  values = [
    file("${var.helm_repo}/values.yaml")
  ]

  # Prompt Backend Server
  set {
    name  = "image.repository"
    value = "opea/promptregistry-mongo-server"
  }

  set {
    name  = "image.tag"
    value = "latest"
  }

  # MongoDB dependency configuration - enabled by default per README
  set {
    name  = "mongodb.enabled"
    value = var.mongodb_enabled
  }

  # Global storage class configuration
  set {
    name  = "global.modelStorageClass"
    value = var.storage_class_name
  }

  # MongoDB persistence configuration
  set {
    name  = "mongodb.persistence.enabled"
    value = "true"
  }

  set {
    name  = "mongodb.persistence.storageClass"
    value = var.database_storage_class_name
  }

  set {
    name  = "mongodb.persistence.size"
    value = var.mongodb_storage_size
  }
}
