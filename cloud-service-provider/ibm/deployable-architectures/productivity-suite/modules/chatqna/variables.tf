variable "helm_chart_path" {
  description = "Path to the ChatQNA Helm chart"
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
  description = "Main LLM model ID"
  type        = string
}

variable "embedding_model_name" {
  description = "Embedding model ID"
  type        = string
}

variable "reranker_model_name" {
  description = "Reranker model ID"
  type        = string
}

variable "llm_service_host_ip" {
  description = "LLM Service Host IP"
  type        = string
}

# Feature Flags

variable "enable_vllm" {
  description = "Enable VLLM component"
  type        = bool
  default     = false
}

variable "enable_ui" {
  description = "Enable ChatQnA UI component"
  type        = bool
  default     = false
}

variable "enable_nginx" {
  description = "Enable Nginx component"
  type        = bool
  default     = false
}