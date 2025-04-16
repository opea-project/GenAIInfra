resource "helm_release" "chathistory-usvc" {
  name             = "chathistory-usvc"
  chart            = var.helm_chart_path
  namespace        = "chathistory"
  create_namespace = true
  timeout          = 600
  
  values = [
    file("${var.helm_chart_path}/values.yaml")
  ]

  # Chathistory Backend Server
  set {
    name  = "image.repository"
    value = "opea/chathistory-mongo-server"
  }

  set {
    name  = "image.tag"
    value = "1.2"
  }

  set {
    name  = "image.pullPolicy"
    value = "IfNotPresent"
  }

  # MongoDB dependency configuration
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
