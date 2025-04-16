resource "helm_release" "codegen" {
  name             = "codegen"
  chart            = var.helm_chart_path
  namespace        = "codegen"
  create_namespace = true
  timeout          = 600
  
  values = [
    file("${var.helm_chart_path}/values.yaml"),
    file("${path.root}/helm-values/codegen_values.yaml")
  ]

  # CodeGen Backend Server
  set {
    name  = "image.repository"
    value = "opea/codegen"
  }

  set {
    name  = "image.tag"
    value = "fix_first_token"
  }

  set {
    name  = "image.pullPolicy"
    value = "Never"
  }

  # Global Configuration
  set_sensitive {
    name  = "global.HUGGINGFACEHUB_API_TOKEN"
    value = var.hf_token
  }
  
  set {
    name  = "global.modelUseHostPath"
    value = var.model_dir
  }

  set {
    name  = "global.modelStorageClass"
    value = "standard"
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


