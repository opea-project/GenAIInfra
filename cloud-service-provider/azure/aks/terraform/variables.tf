variable "location" {
  description = "Azure region"
  type        = string
  default     = "eastus"
}

variable "cluster_name" {
  description = "ASK cluster name"
  type        = string
  default     = "opea ask cluster"
}

variable "kubernetes_version" {
  description = "ASK cluster version"
  type        = string
  default     = "1.30" 
}

variable "use_custom_node_config" {
  description = "Enable custom node configuration"
  type        = bool
  default     = true
}

variable "subscription_id" {
  description  = "This is the Azure subscription id of the user"
  type         = string
}

variable "os_disk_size_gb" {
  description = "OS disk size in GB for nodes"
  type        = number
  default     = 50
}

variable "node_pool_type" {
  description = "Azure VM Scale Set priority"
  type        = string
  default     = "Regular"  # Regular for on-demand, Spot for spot instances
}

variable "min_count" {
  description = "Minimum number of nodes"
  type        = number
  default     = 1
}

variable "max_count" {
  description = "Maximum number of nodes"
  type        = number
  default     = 10
}

variable "node_count" {
  description = "Desired number of nodes"
  type        = number
  default     = 1
}

variable "resource_group_name" {
  description = "Name of the resource group"
  type        = string
  default     = null
}

variable "vnet_subnet_id" {
  description = "ID of the subnet where the cluster will be deployed"
  type        = string
  default     = null
}


variable "cluster_version" {
  description = "Kubernetes version for the cluster"
  type = string
  default = "1.30"
}

variable "instance_types" {
  description = "Azure VM size for nodes"
  type    = list(string)
  default = ["Standard_D32d_v5"]
}
