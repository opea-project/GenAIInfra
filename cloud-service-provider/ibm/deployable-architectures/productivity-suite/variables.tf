# IBM Cloud Provider Variables
variable "ibmcloud_api_key" {
  description = "IBM Cloud API key"
  type        = string
}

variable "email" {
  description = "IBM cloud email address"
  type        = string
  default     = ""
}

# Kubeconfig file variable
variable "kubeconfig_file" {
  description = "Path to the kubeconfig file for the target Kubernetes cluster."
  type        = string
  default     = "~/.kube/config"
}

# API Tokens
variable "hf_token" {
  description = "HuggingFace API token for model access"
  type        = string
  sensitive   = true
}

variable "openai_api_key" {
  description = "OpenAI API key for LLM services"
  type        = string
  sensitive   = true
}

# ChatQNA Variables
variable "chatqna_model_dir" {
  description = "Directory path for ChatQNA models"
  type        = string
}

variable "chatqna_helm_repo" {
  description = "Path to the ChatQNA Helm chart"
  type        = string
}

variable "chatqna_model_name" {
  description = "Name of the main ChatQNA model"
  type        = string
}

variable "chatqna_embedding_model_name" {
  description = "Name of the embedding model for ChatQNA"
  type        = string
}

variable "chatqna_reranker_model_name" {
  description = "Name of the reranker model for ChatQNA"
  type        = string
}

variable "chatqna_llm_service_host_ip" {
  description = "LLM Service Host IP for ChatQNA"
  type        = string
}

# Codegen Variables
variable "codegen_model_dir" {
  description = "Directory path for Codegen models"
  type        = string
}

variable "codegen_helm_repo" {
  description = "Path to the Codegen Helm chart"
  type        = string
}

variable "codegen_model_name" {
  description = "Name of the Codegen model"
  type        = string
}

variable "codegen_llm_service_host_ip" {
  description = "LLM Service Host IP for ChatQNA"
  type        = string
}

# Docsum Variables
variable "docsum_model_dir" {
  description = "Directory path for Docsum models"
  type        = string
}

variable "docsum_helm_repo" {
  description = "Path to the Docsum Helm chart"
  type        = string
}

variable "docsum_model_name" {
  description = "Name of the Docsum model"
  type        = string
}

variable "docsum_llm_service_host_ip" {
  description = "LLM Service Host IP for ChatQNA"
  type        = string
}


# Feature Flags
variable "enable_chatqna_vllm" {
  description = "Enable VLLM for ChatQNA"
  type        = bool
  default     = false
}

variable "enable_chatqna_ui" {
  description = "Enable UI for ChatQNA"
  type        = bool
  default     = false
}

variable "enable_chatqna_nginx" {
  description = "Enable Nginx for ChatQNA"
  type        = bool
  default     = false
}

variable "enable_codegen_ui" {
  description = "Enable Codegen UI"
  type        = bool
  default     = true
}

variable "enable_codegen_tgi" {
  description = "Enable TGI for Codegen"
  type        = bool
  default     = true
}

variable "enable_codegen_llm-uservice" {
  description = "Enable LLM Microservice for Codegen"
  type        = bool
  default     = true
}

variable "enable_codegen_nginx" {
  description = "Enable Nginx for Codegen"
  type        = bool
  default     = true
}

variable "enable_docsum_tgi" {
  description = "Enable TGI for Docsum"
  type        = bool
  default     = false
}

variable "enable_docsum_vllm" {
  description = "Enable VLLM for Docsum"
  type        = bool
  default     = false
}

variable "enable_docsum_nginx" {
  description = "Enable Nginx for Docsum"
  type        = bool
  default     = false
}

variable "enable_docsum_ui" {
  description = "Enable UI for Docsum"
  type        = bool
  default     = false
}

variable "enable_docsum_llm-uservice" {
  description = "Enable LLM Microservice for Docsum"
  type        = bool
  default     = false
}

variable "enable_docsum_whisper" {
  description = "Enable Whisper for Docsum"
  type        = bool
  default     = false
}

# Chathistory Variables
variable "chathistory_helm_repo" {
  description = "Path to the Chathistory Helm chart"
  type        = string
}

variable "enable_chathistory_mongodb" {
  description = "Enable MongoDB for Chathistory"
  type        = bool
  default     = true
}

# Prompt Variables
variable "prompt_helm_repo" {
  description = "Path to the Prompt Helm chart"
  type        = string
}

variable "enable_prompt_mongodb" {
  description = "Enable MongoDB for Prompt"
  type        = bool
  default     = true
}

# Nginx and UI Variables
variable "nginx_helm_repo" {
  description = "Path to the Nginx Helm chart"
  type        = string
}

variable "ui_helm_repo" {
  description = "Path to the UI Helm chart"
  type        = string
}

# Storage Variables
variable "enable_storage_csi_driver" {
  description = "Enable IBM Cloud File Storage for VPC v2.0"
  type        = bool
  default     = false
}

variable "model_storage_size" {
  description = "Size of the storage volume for models"
  type        = string
  default     = "100Gi"
}

# Storage class to use for persistent volumes
variable "storage_class_name" {
  description = "Storage class name for RWX volumes like HuggingFace cache (e.g., ibmc-vpc-file-retain-500-iops)"
  type        = string
  default     = "ibmc-vpc-file-retain-500-iops"
}

# Storage class for database volumes (RWO)
variable "database_storage_class_name" {
  description = "Storage class name for database RWO volumes (e.g., ibmc-vpc-block-retain-10iops-tier)"
  type        = string
  default     = "ibmc-vpc-block-retain-10iops-tier"
}

# Individual service storage sizes
variable "chatqna_storage_size" {
  description = "Storage size for ChatQnA service HuggingFace cache"
  type        = string
  default     = "20Gi"
}

variable "chathistory_storage_size" {
  description = "Storage size for ChatHistory MongoDB"
  type        = string
  default     = "10Gi"
}

variable "prompt_storage_size" {
  description = "Storage size for Prompt MongoDB"
  type        = string
  default     = "10Gi"
}

variable "keycloak_storage_size" {
  description = "Storage size for Keycloak PostgreSQL"
  type        = string
  default     = "5Gi"
}