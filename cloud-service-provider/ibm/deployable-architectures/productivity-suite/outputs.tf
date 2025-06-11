# Keycloak outputs
output "keycloak_endpoint" {
  description = "LoadBalancer endpoint for Keycloak"
  value = try("http://${coalesce(
    data.kubernetes_service.keycloak.status[0].load_balancer[0].ingress[0].ip,
    data.kubernetes_service.keycloak.status[0].load_balancer[0].ingress[0].hostname
  )}:8080", "pending")
}

output "keycloak_cluster_ip" {
  description = "Cluster IP for Keycloak"
  value       = "${helm_release.keycloak.name}.${helm_release.keycloak.namespace}.svc.cluster.local"
}

output "keycloak_status" {
  description = "Status of the Keycloak Helm release"
  value       = helm_release.keycloak.status
}

# Add data source to get service details
data "kubernetes_service" "nginx" {
  depends_on = [helm_release.nginx]
  metadata {
    name      = "nginx-nginx-central-gateway" # Update this name
    namespace = "nginx"
  }
}

# UI outputs
output "ui_endpoint" {
  description = "UI Service Endpoint"
  value = try(
    format(
      "http://%s",
      coalesce(
        data.kubernetes_service.nginx.status[0].load_balancer[0].ingress[0].ip,
        data.kubernetes_service.nginx.status[0].load_balancer[0].ingress[0].hostname
      )
    ),
    "Pending"
  )

  depends_on = [
    helm_release.nginx,
    data.kubernetes_service.nginx
  ]
}

output "ui_cluster_ip" {
  description = "Cluster IP for UI"
  value       = "${helm_release.ui.name}.${helm_release.ui.namespace}.svc.cluster.local"
}

output "ui_status" {
  description = "Status of the UI Helm release"
  value       = helm_release.ui.status
}

# ChatQnA outputs
output "chatqna_status" {
  description = "Status of the ChatQnA Helm release"
  value       = module.chatqna.release_status
}

output "chatqna_namespace" {
  description = "Namespace where ChatQnA is deployed"
  value       = module.chatqna.namespace
}

# CodeGen outputs
output "codegen_status" {
  description = "Status of the CodeGen Helm release"
  value       = module.codegen.release_status
}

output "codegen_namespace" {
  description = "Namespace where CodeGen is deployed"
  value       = module.codegen.namespace
}

# DocSum outputs
output "docsum_status" {
  description = "Status of the DocSum Helm release"
  value       = module.docsum.release_status
}

output "docsum_namespace" {
  description = "Namespace where DocSum is deployed"
  value       = module.docsum.namespace
}

# Chathistory outputs
output "chathistory_release_status" {
  description = "Status of the Chathistory Helm release"
  value       = module.chathistory-usvc.release_status
}

# Prompt outputs
output "prompt_release_status" {
  description = "Status of the Prompt Helm release"
  value       = module.prompt-usvc.release_status
}

# Nginx endpoint (CLI helper)
output "nginx_endpoint" {
  description = "The endpoint for the nginx central gateway"
  value       = "Once deployed, run: kubectl get svc -n nginx nginx -o jsonpath='{.status.loadBalancer.ingress[0].ip}'"
}

# Admin credentials (marked as sensitive)
output "keycloak_admin_credentials" {
  description = "Keycloak admin credentials"
  value = {
    username = "admin"
    password = "admin" # In production, use a secret manager or variable
  }
  sensitive = true
}

# Overall deployment status
output "deployment_status" {
  description = "Status of all deployments"
  value = {
    keycloak    = helm_release.keycloak.status
    ui          = helm_release.ui.status
    chatqna     = module.chatqna.release_status
    codegen     = module.codegen.release_status
    docsum      = module.docsum.release_status
    chathistory = module.chathistory-usvc.release_status
    prompt      = module.prompt-usvc.release_status
  }
}