variable "helm_chart_path" {
  description = "Path to the Codegen Helm chart"
  type        = string
}

variable "hf_token" {
  description = "HuggingFace API token"
  type        = string
  sensitive   = true
}

variable "openai_api_key" {
  description = "OpenAI API Key"
  type        = string
  sensitive   = true
}

variable "model_dir" {
  description = "Directory path for model storage"
  type        = string
}

variable "model_name" {
  description = "Codegen model ID"
  type        = string
}

variable "llm_service_host_ip" {
  description = "LLM Service Host IP"
  type        = string
}

variable "enable_ui" {
  description = "Enable Codegen UI component"
  type        = bool
  default     = true
}

variable "enable_tgi" {
  description = "Enable TGI for Codegen"
  type        = bool
  default     = false
}

variable "enable_llm-uservice" {
  description = "Enable Codegen LLM Uservice"
  type        = bool
  default     = false
}

variable "enable_nginx" {
  description = "Enable Codegen NGINX"
  type        = bool
  default     = false
}


