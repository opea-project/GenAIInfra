# Example Terraform variables file
# Copy this file to terraform.tfvars and populate with your actual values

# IBM Cloud API key - Get from IBM Cloud console
ibmcloud_api_key = "your-ibm-cloud-api-key-here"

# Path to kubeconfig for IBM IKS cluster
kubeconfig_file = "~/.kube/ibm-iks-kubeconfig"

# API Tokens - Replace with your actual tokens
hf_token       = "hf_your-huggingface-token-here"
openai_api_key = "your-openai-api-key-or-jwt-token-here"

# ChatQNA Configuration
chatqna_model_dir = ""
chatqna_helm_repo = "../../../../helm-charts/chatqna"

chatqna_model_name           = "your-chatqna-model-name"
chatqna_embedding_model_name = "your-embedding-model-name"
chatqna_reranker_model_name  = "your-reranker-model-name"
chatqna_llm_service_host_ip  = "https://your-llm-service-endpoint.com/model-name"

# Codegen Configuration
codegen_model_dir = ""
codegen_helm_repo = "../../../../helm-charts/codegen"

codegen_model_name          = "meta-llama/Llama-3.1-405B-Instruct"
codegen_llm_service_host_ip = "https://your-llm-service-endpoint.com/model-name"

# Docsum Configuration
docsum_model_dir = ""
docsum_helm_repo = "../../../../helm-charts/docsum"

docsum_model_name          = "meta-llama/Llama-3.1-405B-Instruct"
docsum_llm_service_host_ip = "https://your-llm-service-endpoint.com/model-name"

# Nginx Configuration
nginx_helm_repo = "../../../../helm-charts/common/nginx"

# UI Configuration
ui_helm_repo = "../../../../helm-charts/common/ui"

# Chathistory Configuration
chathistory_helm_repo = "../../../../helm-charts/common/chathistory-usvc"

enable_chathistory_mongodb = true

# Prompt Configuration
prompt_helm_repo = "../../../../helm-charts/common/prompt-usvc"

enable_prompt_mongodb = true

# Feature Flags - ChatQnA
enable_chatqna_vllm  = false
enable_chatqna_ui    = false
enable_chatqna_nginx = false

# Feature Flags - Codegen
enable_codegen_ui           = false
enable_codegen_tgi          = false
enable_codegen_llm-uservice = false
enable_codegen_nginx        = false

# Feature Flags - Docsum
enable_docsum_tgi          = false
enable_docsum_vllm         = false
enable_docsum_nginx        = false
enable_docsum_ui           = false
enable_docsum_llm-uservice = false
enable_docsum_whisper      = false

# Storage Configuration
storage_class_name          = "ibmc-vpc-file-retain-500-iops"     # For RWX volumes (HuggingFace cache)
database_storage_class_name = "ibmc-vpc-block-retain-10iops-tier" # For RWO database volumes

# Individual service storage sizes
chatqna_storage_size     = "20Gi"
chathistory_storage_size = "10Gi"
prompt_storage_size      = "10Gi"
keycloak_storage_size    = "5Gi"