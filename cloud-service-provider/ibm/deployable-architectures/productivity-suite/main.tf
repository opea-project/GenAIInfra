# ChatQnA Module
module "chatqna" {
  source = "./modules/chatqna"

  hf_token        = local.hf_token
  openai_api_key = local.openai_api_key
  model_dir       = var.chatqna_model_dir
  helm_chart_path = var.chatqna_helm_chart_path
  
  # ChatQNA specific model names
  model_name          = var.chatqna_model_name
  embedding_model_name = var.chatqna_embedding_model_name
  reranker_model_name = var.chatqna_reranker_model_name

  # Model Endpoint for ChatQnA
  llm_service_host_ip = var.chatqna_llm_service_host_ip
  
  # Feature flags
  enable_vllm            = var.enable_chatqna_vllm
  enable_ui       = var.enable_chatqna_ui
  enable_nginx = var.enable_chatqna_nginx
}

# Codegen Module
module "codegen" {
  source = "./modules/codegen"

  hf_token        = local.hf_token
  openai_api_key = local.openai_api_key
  model_dir       = var.codegen_model_dir
  helm_chart_path = var.codegen_helm_chart_path
  model_name      = var.codegen_model_name

  # Model Endpoint for CodeGen
  llm_service_host_ip = var.codegen_llm_service_host_ip

  # Feature flags
  enable_ui    = var.enable_codegen_ui
  enable_tgi   = var.enable_codegen_tgi
  enable_llm-uservice = var.enable_codegen_llm-uservice
  enable_nginx = var.enable_codegen_nginx
}

# Docsum Module
module "docsum" {
  source = "./modules/docsum"

  hf_token        = local.hf_token
  openai_api_key = local.openai_api_key
  model_dir       = var.docsum_model_dir
  helm_chart_path = var.docsum_helm_chart_path
  model_name      = var.docsum_model_name

  # Model Endpoint for DocSum
  llm_service_host_ip = var.docsum_llm_service_host_ip

  # Feature flags
  enable_tgi   = var.enable_docsum_tgi
  enable_vllm  = var.enable_docsum_vllm
  enable_nginx = var.enable_docsum_nginx
  enable_ui    = var.enable_docsum_ui
  enable_llm-uservice = var.enable_docsum_llm-uservice
  enable_whisper = var.enable_docsum_whisper
}

# Chathistory Microservice Module
module "chathistory-usvc" {
  source = "./modules/chathistory-usvc"

  helm_chart_path = var.chathistory_helm_chart_path
  mongodb_enabled = var.enable_chathistory_mongodb
}

# Prompt Microservice Module
module "prompt-usvc" {
  source = "./modules/prompt-usvc"

  helm_chart_path = var.prompt_helm_chart_path
  mongodb_enabled = var.enable_prompt_mongodb
}

# UI Helm Release
resource "helm_release" "ui" {
  name             = "ui"
  chart            = var.ui_helm_chart_path
  namespace        = "ui"
  create_namespace = true
  timeout          = 600
  
  # Ensure the UI is deployed before updating nginx
  # (if you're using both in the same Terraform configuration)
  # Use the ui_values.yaml file from helm-values
  values = [
    file("${path.root}/helm-values/ui_values.yaml")
  ]
  
  # Explicitly set values that can be overridden
  set {
    name  = "image.repository"
    value = "immersive-ui"
  }
  
  set {
    name  = "image.tag"
    value = "lite"
  }
  
  set {
    name  = "containerPort"
    value = "5173"
  }
  
  set {
    name  = "APP_KEYCLOAK_SERVICE_ENDPOINT"
    value = "/auth"
  }
}

# Nginx Helm Release
resource "helm_release" "nginx" {
  name             = "nginx"
  chart            = var.nginx_helm_chart_path
  namespace        = "nginx"
  create_namespace = true
  timeout          = 600
}

# Keycloak Helm Release
resource "helm_release" "keycloak" {
  name       = "keycloak"
  repository = "https://charts.bitnami.com/bitnami"
  chart      = "keycloak"
  namespace  = "keycloak"
  create_namespace = true
  version          = "15.0.0"

  set {
    name  = "auth.adminUser"
    value = "admin"
  }
  
  set {
    name  = "auth.adminPassword"
    value = "admin"
  }

  set {
    name  = "postgresql.enabled"
    value = "true"
  }

  set {
    name  = "postgresql.primary.persistence.enabled"
    value = "false"
  }

  # Use Nodeport
  set {
    name  = "service.type"
    value = "NodePort"
  }

  set {
    name  = "service.nodePorts.http"
    value = "31893"
  }
}