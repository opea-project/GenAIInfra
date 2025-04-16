resource "helm_release" "prompt-usvc" {
  name             = "prompt-usvc"
  chart            = var.helm_chart_path
  namespace        = "prompt"
  create_namespace = true
  timeout          = 600
  
  values = [
    file("${var.helm_chart_path}/values.yaml")
  ]

  # Prompt Backend Server
  set {
    name  = "image.repository"
    value = "opea/promptregistry-mongo-server"
  }

  set {
    name  = "image.tag"
    value = "1.2"
  }

  set {
    name  = "image.pullPolicy"
    value = "IfNotPresent"
  }

  # MongoDB dependency configuration - enabled by default per README
  set {
    name  = "mongodb.enabled"
    value = var.mongodb_enabled
  }

  # Global storage class configuration
  set {
    name  = "global.modelStorageClass"
    value = "standard"
  }
}
