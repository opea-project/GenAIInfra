terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "4.74.0"
    }
kubernetes = {
      source = "hashicorp/kubernetes"
      version = "2.33.0"
    }
  }

  required_version = ">= 0.14"
}

