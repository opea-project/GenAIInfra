variable "gke_num_nodes" {
  default     = 0
  description = "number of gke nodes"
}

variable "cluster_name" {
  description = "A suffix to append to the default cluster name"
  default     = null
}

variable "node_locations" {
  type    = list(string)
  default = null
}

variable "initial_node_count" {
  type    = number
  default = null
}

variable "disk_size_gb" {
  type    = number
  default = 20
}

variable "image_type" {
  type    = string
  default = null
}

variable "machine_type" {
  type    = string
  default = null
}

variable "max_node_count" {
  type    = number
  default = null
}

variable "min_node_count" {
  type    = number
  default = null
}
