data "google_client_config" "default" {}
data "google_project" "current" { project_id = var.project_id }

provider "kubernetes" {
  host                   = "https://${module.gke.endpoint}"
  token                  = data.google_client_config.default.access_token
  cluster_ca_certificate = base64decode(module.gke.ca_certificate)
}

provider "helm" {
  kubernetes {
    host                   = "https://${module.gke.endpoint}"
    token                  = data.google_client_config.default.access_token
    cluster_ca_certificate = base64decode(module.gke.ca_certificate)    
  }
}
resource "google_compute_firewall" "default" {
  #count    = var.firewall ? 1 : 0
  name    = "${var.cluster_name}-firewall"
  network = google_compute_network.default.name

  deny {
    protocol = "tcp"
    ports    = ["20-22", "3389"]
  }
  #target_tags = [ "${var.cluster_name}-firewall" ]
  source_ranges = ["0.0.0.0/0"]
}

resource "google_compute_network" "default" {
  name                    = "standalone"
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "default" {
  name          = "example-subnetwork"
  region        = var.region
  ip_cidr_range = "10.0.0.0/16"
  stack_type    = "IPV4_ONLY"

  network = google_compute_network.default.id
  secondary_ip_range {
    range_name    = "services-range"
    ip_cidr_range = "192.168.0.0/24"
  }

  secondary_ip_range {
    range_name    = "pod-ranges"
    ip_cidr_range = "172.16.0.0/12"
  }
}

module "gke" {
  source                   = "terraform-google-modules/kubernetes-engine/google"
  version                  = "34.0.0"
  project_id               = var.project_id
  name                     = var.cluster_name
  region                   = var.region
  kubernetes_version       = var.cluster_version
  network                  = google_compute_network.default.name
  subnetwork               = google_compute_subnetwork.default.name
  ip_range_pods            = google_compute_subnetwork.default.secondary_ip_range[1].range_name
  ip_range_services        = google_compute_subnetwork.default.secondary_ip_range[0].range_name
  gcs_fuse_csi_driver      = true
  deletion_protection      = false
  remove_default_node_pool = true
  node_pools               = var.cpu_pool

  node_pools_oauth_scopes = {
    all = [
      "https://www.googleapis.com/auth/cloud-platform",
      "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring",
      "https://www.googleapis.com/auth/service.management.readonly",
      "https://www.googleapis.com/auth/servicecontrol",
    ]
  }
}

resource "null_resource" "kubectl" {
  provisioner "local-exec" {
    command = "gcloud container clusters get-credentials ${var.cluster_name} --region ${var.region}"
  }
  depends_on = [ module.gke ]
}

resource "kubernetes_namespace" "opea_app" {
  metadata {
    name = var.namespace
  }
}

resource "kubernetes_service_account" "opea_gcs_sa" {
  metadata {
    name = "opea-gcs-sa"
    namespace = var.namespace
  }
  depends_on = [kubernetes_namespace.opea_app]
}

resource "google_storage_bucket" "model" {
  name          = "${var.gcs_bucket_name}"
  location      = var.gcs_bucket_location
  force_destroy = true
  
  uniform_bucket_level_access = true
}

resource "google_storage_bucket_iam_binding" "opea_gcs_sa_binding" {
  bucket = google_storage_bucket.model.name
  role = "roles/storage.objectUser"
  members = [
  # FIXME: we can't use the SA we created due to #532
  #  "principal://iam.googleapis.com/projects/${data.google_project.current.number}/locations/global/workloadIdentityPools/${data.google_project.current.project_id}.svc.id.goog/subject/ns/${kubernetes_service_account.opea_gcs_sa.metadata[0].namespace}/sa/${kubernetes_service_account.opea_gcs_sa.metadata[0].name}",
    "principal://iam.googleapis.com/projects/${data.google_project.current.number}/locations/global/workloadIdentityPools/${data.google_project.current.project_id}.svc.id.goog/subject/ns/${kubernetes_service_account.opea_gcs_sa.metadata[0].namespace}/sa/default",
  ]
  depends_on = [kubernetes_service_account.opea_gcs_sa]
}

resource "kubernetes_persistent_volume_claim" "model" {
  metadata {
    name = "model-volume"
    namespace = var.namespace
  }
  spec {
    storage_class_name = "dummy"
    access_modes = ["ReadWriteMany"]
    resources {
      requests = {
        storage = "50Gi"
      }
    }
    volume_name = "${kubernetes_persistent_volume.model.metadata.0.name}"
  }
  depends_on = [ null_resource.kubectl ]
}

resource "kubernetes_persistent_volume" "model" {
  metadata {
    name = "opea-model-pv"
  }
  spec {
    capacity = {
      storage = "50Gi"
    }
    storage_class_name = "dummy"
    access_modes = ["ReadWriteMany"]
    persistent_volume_source {
      csi {
        driver = "gcsfuse.csi.storage.gke.io"
        volume_handle = google_storage_bucket.model.name
      }
    }
    mount_options = [ "implicit-dirs", "uid=1000", "gid=1000" ]
  }
  depends_on = [ null_resource.kubectl ]
}
