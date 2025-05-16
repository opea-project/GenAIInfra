variable "helm_chart_path" {
  description = "Path to the Docsum Helm chart"
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
  description = "Docsum model ID"
  type        = string
}

variable "llm_service_host_ip" {
  description = "LLM Service Host IP"
  type        = string
}

# Feature Flags

variable "enable_tgi" {
  description = "Enable TGI for Docsum"
  type        = bool
  default     = false
}

variable "enable_vllm" {
  description = "Enable VLLM for Docsum"
  type        = bool
  default     = false
}

variable "enable_nginx" {
  description = "Enable Nginx for Docsum"
  type        = bool
  default     = false
}

variable "enable_ui" {
  description = "Enable UI for Docsum"
  type        = bool
  default     = false
}

variable "enable_llm-uservice" {
  description = "Enable llm-uservice for Docsum"
  type        = bool
  default     = false
}

variable "enable_whisper" {
  description = "Enable Whisper for Docsum"
  type        = bool
  default     = false
}

