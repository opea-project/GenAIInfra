# Determine if helm_repo is a local path or remote URI
locals {
  is_remote = can(regex("^([a-zA-Z0-9.-]+)/([a-zA-Z0-9-]+)/([a-zA-Z0-9-]+)$", var.helm_repo))
  repo_parts = local.is_remote ? regex("^([a-zA-Z0-9.-]+)/([a-zA-Z0-9-]+)/([a-zA-Z0-9-]+)$", var.helm_repo) : []
}

resource "helm_release" "chatqna" {
  name             = "chatqna"
  chart      = local.is_remote ? local.repo_parts[3] : var.helm_repo
  repository = local.is_remote ? "oci://${local.repo_parts[1]}/${local.repo_parts[2]}" : null
  namespace        = "chatqna"
  create_namespace = false
  timeout          = 600
  
  values = [
    file("${var.helm_repo}/values.yaml"),
    file("${path.root}/helm_values/chatqna_values.yaml")
  ]

# ChatQnA Backend Server
  set {
    name  = "image.repository"
    value = "us.icr.io/ibm-opea-terraform/chatqna"
  }

  set {
    name  = "image.tag"
    value = "accelerate"
  }

  set {
    name  = "image.pullPolicy"
    value = "Always"
  }

  set {
    name  = "imagePullSecrets[0].name"
    value = "regcred"
  }

  # VLLM Configuration
  set {
    name  = "vllm.enabled"
    value = var.enable_vllm
  }
  
  set {
    name  = "vllm.image.repository"
    value = "opea/vllm"
  }
  
  set {
    name  = "vllm.image.tag"
    value = "latest"
  }

  # Global Configuration
  set_sensitive {
    name  = "global.HUGGINGFACEHUB_API_TOKEN"
    value = var.hf_token
  }

  set {
    name  = "global.modelUsePVC"
    value = "chatqna-storage"
  }

  set {
    name  = "global.modelStorageClass"
    value = var.storage_class_name
  }

  # External LLM Configuration - updated to use externalLLM prefix
  set {
    name  = "externalLLM.LLM_SERVER_HOST_IP"
    value = var.llm_service_host_ip
  }

  set {
    name  = "externalLLM.LLM_MODEL"
    value = var.model_name
  }
  
  set_sensitive {
    name  = "externalLLM.OPENAI_API_KEY"
    value = var.openai_api_key
  }

  # Keep other model configurations with their original prefixes
  set {
    name  = "tei.EMBEDDING_MODEL_ID"
    value = var.embedding_model_name
  }

  set {
    name  = "teirerank.RERANK_MODEL_ID"
    value = var.reranker_model_name
  }

  # Component Enablement - override values from YAML when needed
  # These settings will take precedence over the values in the YAML file

  set {
    name  = "vllm.enabled"
    value = var.enable_vllm
  }

  set {
    name  = "chatqna-ui.enabled"
    value = var.enable_ui
  }

  set {
    name  = "nginx.enabled"
    value = var.enable_nginx
  }
}