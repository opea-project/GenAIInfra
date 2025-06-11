# Determine if helm_repo is a local path or remote URI
locals {
  is_remote = can(regex("^([a-zA-Z0-9.-]+)/([a-zA-Z0-9-]+)/([a-zA-Z0-9-]+)$", var.helm_repo))
  repo_parts = local.is_remote ? regex("^([a-zA-Z0-9.-]+)/([a-zA-Z0-9-]+)/([a-zA-Z0-9-]+)$", var.helm_repo) : []
}

resource "helm_release" "codegen" {
  name             = "codegen"
  chart      = local.is_remote ? local.repo_parts[3] : var.helm_repo
  repository = local.is_remote ? "oci://${local.repo_parts[1]}/${local.repo_parts[2]}" : null
  namespace        = "codegen"
  create_namespace = false
  timeout          = 600
  
  values = [
    file("${var.helm_repo}/values.yaml"),
    file("${path.root}/helm_values/codegen_values.yaml")
  ]

# CodeGen Backend Server
  set {
    name  = "image.repository"
    value = "us.icr.io/ibm-opea-terraform/codegen"
  }

  set {
    name  = "image.tag"
    value = "ibm"
  }

  set {
    name  = "image.pullPolicy"
    value = "Always"
  }

  set {
    name  = "imagePullSecrets[0].name"
    value = "regcred"
  }

  # Global Configuration
  set_sensitive {
    name  = "global.HUGGINGFACEHUB_API_TOKEN"
    value = var.hf_token
  }
  
  # No additional settings needed - already configured above

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

  # Component Enablement - override values from YAML when needed
  # These settings will take precedence over the values in the YAML file
  set {
    name  = "codegen-ui.enabled"
    value = var.enable_ui
  }

  set {
    name  = "tgi.enabled"
    value = var.enable_tgi
  }

  set {
    name  = "llm-uservice.enabled"
    value = var.enable_llm-uservice
  }

  set {
    name  = "nginx.enabled"
    value = var.enable_nginx
  }
}


