resource "helm_release" "chatqna" {
  name             = "chatqna"
  chart            = var.helm_chart_path
  namespace        = "chatqna"
  create_namespace = true
  timeout          = 600
  
  values = [
    file("${var.helm_chart_path}/values.yaml"),
    file("${path.root}/helm-values/chatqna_values.yaml")
  ]

  # ChatQnA Backend Server
  set {
    name  = "image.repository"
    value = "opea/chatqna"
  }

  set {
    name  = "image.tag"
    value = "rag_template"
  }

  set {
    name  = "image.pullPolicy"
    value = "Never"
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
    value = "1.2"
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