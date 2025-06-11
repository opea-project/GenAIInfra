variable "helm_repo" {
  description = "Path to the Chathistory Helm chart"
  type        = string
}

variable "mongodb_enabled" {
  description = "Enable MongoDB dependency"
  type        = bool
  default     = true
}

variable "storage_class_name" {
  description = "Storage class name for RWX volumes"
  type        = string
  default     = "ibmc-vpc-file-retain-500-iops"
}

variable "database_storage_class_name" {
  description = "Storage class name for database RWO volumes"
  type        = string
  default     = "ibmc-vpc-block-retain-10iops-tier"
}

variable "mongodb_storage_size" {
  description = "Storage size for MongoDB"
  type        = string
  default     = "8Gi"
}
