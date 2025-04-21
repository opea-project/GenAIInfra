# GKE cluster
data "google_container_engine_versions" "gke_version" {
  location       = var.region
  version_prefix = "1.30."
}

resource "google_container_cluster" "primary" {
  name     = "${var.cluster_name}-gke"
  location = var.region

  # We can't create a cluster with no node pool defined, but we want to only use
  # separately managed node pools. So we create the smallest possible default
  # node pool and immediately delete it.
  remove_default_node_pool = true
  initial_node_count       = var.initial_node_count
  node_locations           = var.node_locations

  network    = google_compute_network.vpc.name
  subnetwork = google_compute_subnetwork.subnet.name
}

# Separately Managed Node Pool
resource "google_container_node_pool" "primary_nodes" {
  name     = google_container_cluster.primary.name
  location = var.region
  cluster  = google_container_cluster.primary.name

  version        = data.google_container_engine_versions.gke_version.release_channel_default_version["STABLE"]
  node_count     = var.gke_num_nodes
  node_locations = var.node_locations
  autoscaling {
    max_node_count = var.max_node_count
    min_node_count = var.min_node_count
  }


  node_config {
    disk_size_gb = var.disk_size_gb
    disk_type    = "hyperdisk-balanced"
    oauth_scopes = [
      "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring",
    ]

    labels = {
      env = var.cluster_name
    }

    preemptible  = true
    image_type   = var.image_type
    machine_type = var.machine_type
    tags         = ["gke-node", "${var.cluster_name}-gke"]
    metadata = {
      disable-legacy-endpoints = "true"
    }
  }
}

