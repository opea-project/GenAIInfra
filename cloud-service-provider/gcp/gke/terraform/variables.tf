variable "hf_token" {
  description = "Hugginface API token"
  type        = string
}

variable "project_id" {
  description = "Google Cloud PROJECT_ID"
  type        = string
}

variable "region" {
  description = "Google Cloud region"
  type        = string
  default     = "europe-west1"
}

variable "zone" {
  description = "Google zone"
  type        = string
  default     = "a"
}

variable "cluster_name" {
  description = "GKE cluster name"
  type        = string
  default     = null
}

variable "cluster_version" {
  description = "GKE cluster version"
  type        = string
  default     = "1.31"
}

variable "namespace" {
  description = "OPEA application namespace"
  type        = string
  default     = "default"
}

variable "cpu_pool" {
  type = list(map(any))
}

variable "disk_size" {
  description = "Disk size in GiB for nodes."
  type        = number
  default     = 20
}

variable "capacity_type" {
  description = "EC2 spot or on-demad instance types"
  type        = string
  default     = "ON_DEMAND"
}

variable "min_size" {
  description = "min size"
  type        = number
  default     = 1
}

variable "max_size" {
  description = "max size"
  type        = number
  default     = 10
}

variable "desired_size" {
  description = "desired size"
  type        = number
  default     = 1
}

variable "compute_engine_service_account" {
  description = "SA for managing the nodes"
  type = string
  default = null
}

variable "gcs_bucket_name" {
  description = "Bucket name for storing model data"
  type = string
  default = "opea-models"
}

variable "gcs_bucket_location" {
  description = "Bucket location"
  type = string
  default = "EU"
}

