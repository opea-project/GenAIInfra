#hf_token = ""
project_id = "service-mesh-296815"
region = "europe-west4"
cluster_name = "opea"
app_name = "chatqna"
namespace = "chatqna"
cpu_pool = [ {
  name: "cpu-pool"
  machine_type: "c4-standard-32"
  autoscaling: false
  min_count: 1
  max_count: 5
  disk_size_gb: 100
  disk_type: "hyperdisk-balanced"
} ]