# IBM Cloud Storage Configuration
# This file is configured for IBM Cloud File Storage for VPC v2.0

# Create a namespace for storage management
resource "kubernetes_namespace" "storage" {
  count = var.enable_storage_csi_driver ? 1 : 0
  metadata {
    name = "ibm-storage"
  }

  lifecycle {
    ignore_changes = [
      metadata[0].labels,
      metadata[0].annotations,
    ]
  }
}

# Using predefined IBM Cloud File Storage classes
# Available classes: ibmc-vpc-file-min-iops, ibmc-vpc-file-500-iops, ibmc-vpc-file-1000-iops, ibmc-vpc-file-3000-iops
# For retain policy: ibmc-vpc-file-retain-500-iops, ibmc-vpc-file-retain-1000-iops, ibmc-vpc-file-retain-3000-iops

# PVC for ChatQnA service - used as HuggingFace hub cache
resource "kubernetes_persistent_volume_claim" "chatqna_pvc" {
  metadata {
    name      = "chatqna-storage"
    namespace = "chatqna"
  }
  spec {
    access_modes = ["ReadWriteMany"]
    resources {
      requests = {
        storage = var.chatqna_storage_size
      }
    }
    storage_class_name = var.storage_class_name
  }

  depends_on = [kubernetes_namespace.namespaces]
}

# MongoDB and PostgreSQL persistence is configured directly in the Helm charts
# No separate PVCs needed for chathistory, prompt, and keycloak services
