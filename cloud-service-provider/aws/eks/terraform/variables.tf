variable "region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "cluster_name" {
  description = "EKS cluster name"
  type        = string
  default     = null
}

variable "cluster_version" {
  description = "EKS cluster version"
  type        = string
  default     = "1.31"
}

variable "instance_types" {
  description = "EC2 instance types"
  type        = list
  default     = ["t3.medium"]
}

variable "use_custom_launch_template" {
  description = "Disk size in GiB for nodes."
  type        = bool
  default     = true
}

variable "disk_size" {
  description = "Disk size in GiB for nodes."
  type        = number
  default     = 20
}

variable "capacity_type" {
  description = "EC2 spot or on-demand instance types"
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
