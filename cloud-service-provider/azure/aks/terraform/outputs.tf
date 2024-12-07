output "cluster_endpoint" {
  description = "Endpoint for ASK control plane"
  #sensitive = false
  sensitive = true
  value       = azurerm_kubernetes_cluster.main.kube_config.0.host
}

output "oidc_issuer_url" {
  description = "The URL for the OpenID Connect issuer"
  value       = azurerm_kubernetes_cluster.main.oidc_issuer_url
}

output "location" {
  description = "Azure region"
  value       = var.location
}

output "cluster_name" {
  description = "Kubernetes Cluster Name"
  value       = azurerm_kubernetes_cluster.main.name
}
