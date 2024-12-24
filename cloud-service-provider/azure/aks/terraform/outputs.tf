output "cluster_endpoint" {
  description = "Endpoint for AKS control plane"
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

output "cosmosdb_connection_strings" {
  description  = "Azure Cosmos DB Connection Strings"
  value = "AccountEndpoint=${azurerm_cosmosdb_account.main.endpoint};AccountKey=${azurerm_cosmosdb_account.main.primary_key};"
  sensitive   = true
}

output "instrumentation_key" {
  description = "App Insights Instrumentation Key"
  value = azurerm_application_insights.t_appinsights.instrumentation_key
  sensitive = true
}

output "app_id" {
  description = "App Insights App Id"
  value = azurerm_application_insights.t_appinsights.app_id
}

output "app_insights_connection_string" {
  description = "App Insights Connection String"
  value = azurerm_application_insights.t_appinsights.connection_string
  sensitive = true
}
