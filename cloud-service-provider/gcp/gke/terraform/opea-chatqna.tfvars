#This value required as part of Google Project id.
project_id = "arun-poc"
# Kubernetes cluster Name
cluster_name       = "opea-cluster"
region             = "us-central1"
initial_node_count = 1
max_node_count     = 5
min_node_count     = 1
node_locations     = ["us-central1-a"]
disk_size_gb       = 100
image_type         = "COS_CONTAINERD"
machine_type       = "n4-standard-8"
