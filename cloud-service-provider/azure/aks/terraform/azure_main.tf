provider "kubernetes" {
  config_path = "~/.kube/config"
}

# Resource Group
resource "azurerm_resource_group" "main" {
  name     = "${var.cluster_name}-rg"
  location = var.location
}

# Virtual Network
module "vnet" {
  source              = "Azure/vnet/azurerm"
  resource_group_name = azurerm_resource_group.main.name
  vnet_name           = "${var.cluster_name}-vnet"
  vnet_location       = azurerm_resource_group.main.location
  
  tags = {
    environment = "dev"
  }
  depends_on = [azurerm_resource_group.main]
}

# Cosmos DB
resource "azurerm_cosmosdb_account" "main" {
  count                     = var.is_cosmosdb_required ? 1 : 0
  name                      = "${var.cluster_name}-cosmosdb"
  location                  = var.cosmosdb_account_location
  resource_group_name       = azurerm_resource_group.main.name
  offer_type                = "Standard"
  kind                      = "GlobalDocumentDB"
  geo_location {
    location          = var.cosmosdb_account_location
    failover_priority = 0
  }
  consistency_policy {
    consistency_level       = "BoundedStaleness"
    max_interval_in_seconds = 300
    max_staleness_prefix    = 100000
  }
  capabilities {
    name = "EnableNoSQLFullTextSearch"
  }
  capabilities {
    name = "EnableNoSQLVectorSearch"
  }
  depends_on = [
    azurerm_resource_group.main
  ]
}

resource "azurerm_cosmosdb_sql_database" "main" {
  count               = var.is_cosmosdb_required ? 1 : 0
  name                = "${var.cluster_name}-sqldb"
  resource_group_name = azurerm_resource_group.main.name
  account_name        = azurerm_cosmosdb_account.main[count.index].name
  throughput          = var.throughput
  depends_on = [
    azurerm_cosmosdb_account.main
  ]
}

resource "azurerm_cosmosdb_sql_container" "main" {
  count                 = var.is_cosmosdb_required ? 1 : 0
  name                  = "${var.cluster_name}-sql-container"
  resource_group_name   = azurerm_resource_group.main.name
  account_name          = azurerm_cosmosdb_account.main[count.index].name
  database_name         = azurerm_cosmosdb_sql_database.main[count.index].name
  partition_key_paths   = ["/definition/id"]
  partition_key_version = 1
  throughput            = var.throughput

  indexing_policy {
    indexing_mode = "consistent"

    included_path {
      path = "/*"
    }

    included_path {
      path = "/included/?"
    }

    excluded_path {
      path = "/excluded/?"
    }
  }

  unique_key {
    paths = ["/definition/idlong", "/definition/idshort"]
  }
  depends_on = [
    azurerm_cosmosdb_sql_database.main
  ]
}

# AKS Cluster
resource "azurerm_kubernetes_cluster" "main" {
  name                = var.cluster_name
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  dns_prefix          = var.cluster_name
  kubernetes_version  = var.cluster_version
  private_cluster_public_fqdn_enabled = true

  default_node_pool {
    name                = "default"
    auto_scaling_enabled = true
    node_count          = var.node_count
    vm_size            = var.instance_types[0]
    min_count          = var.min_count
    max_count          = var.max_count
    vnet_subnet_id     = module.vnet.vnet_subnets[0]
    os_disk_size_gb    = var.os_disk_size_gb
  }

  identity {
    type = "SystemAssigned"
  }

  network_profile {
    network_plugin     = "azure"
    load_balancer_sku = "standard"
    service_cidr = "10.0.4.0/24"
    dns_service_ip = "10.0.4.10"
  }

}

# Key Vault
resource "azurerm_key_vault" "main" {
  name                       = "${var.cluster_name}-kv"
  location                   = azurerm_resource_group.main.location
  resource_group_name        = azurerm_resource_group.main.name
  tenant_id                  = data.azurerm_client_config.current.tenant_id
  sku_name                  = "standard"
  soft_delete_retention_days = 7
  purge_protection_enabled   = false

  access_policy {
    tenant_id = data.azurerm_client_config.current.tenant_id
    object_id = data.azurerm_client_config.current.object_id

    key_permissions = [
      "Create",
      "Delete",
      "Get",
      "List",
    ]

    secret_permissions = [
      "Set",
      "Get",
      "Delete",
      "List",
    ]
  }
}

# Update kubeconfig
resource "null_resource" "kubectl" {
  provisioner "local-exec" {
    command = "az aks get-credentials --resource-group ${azurerm_resource_group.main.name} --name ${azurerm_kubernetes_cluster.main.name} --overwrite-existing"
  }
  depends_on = [azurerm_kubernetes_cluster.main]
}

# Application Insights
resource "azurerm_log_analytics_workspace" "main" {
  name                = "workspace-${var.cluster_name}"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  retention_in_days   = 30
}

resource "azurerm_application_insights" "t_appinsights" {
  name                = "${var.cluster_name}-appinsights"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  workspace_id        = azurerm_log_analytics_workspace.main.id
  application_type    = "web"
}

# Data source for Azure subscription information
data "azurerm_client_config" "current" {}

# Cosmos db Primary connection string into key vault
resource "azurerm_key_vault_secret" "cosdb_primary" {
  count        = var.is_cosmosdb_required ? 1 : 0
  name         = "AzCosmosDBConnectionStringPrimary"
  value        = tostring("AccountEndpoint=${azurerm_cosmosdb_account.main[count.index].endpoint};AccountKey=${azurerm_cosmosdb_account.main[count.index].primary_key};")
  key_vault_id = azurerm_key_vault.main.id
}

# Cosmos db Secondary connection string into key vault
resource "azurerm_key_vault_secret" "cosdb_secondary" {
  count        = var.is_cosmosdb_required ? 1 : 0
  name         = "AzCosmosDBConnectionStringSecondary"
  value        = tostring("AccountEndpoint=${azurerm_cosmosdb_account.main[count.index].endpoint};AccountKey=${azurerm_cosmosdb_account.main[count.index].secondary_key};")
  key_vault_id = azurerm_key_vault.main.id
}

# Kubernetes cluster end point into key vault
resource "azurerm_key_vault_secret" "kube_cluster_endpoint" {
  name         = "KubeClusterEndPoint"
  value        = tostring("${azurerm_kubernetes_cluster.main.kube_config.0.host}")
  key_vault_id = azurerm_key_vault.main.id
}

# App Insights Instrumentation Key
resource "azurerm_key_vault_secret" "app_insights_instrumentation_key" {
  name         = "AppInsightsInstrumentationKey"
  value        = tostring("${azurerm_application_insights.t_appinsights.instrumentation_key}")
  key_vault_id = azurerm_key_vault.main.id
}

# App Insights app id
resource "azurerm_key_vault_secret" "app_insights_app_id" {
  name         = "AppInsightsAppId"
  value        = tostring("${azurerm_application_insights.t_appinsights.app_id}")
  key_vault_id = azurerm_key_vault.main.id
}

# App Insights Connection String
resource "azurerm_key_vault_secret" "app_insights_connection_string" {
  name         = "AppInsightsConnectionString"
  value        = tostring("${azurerm_application_insights.t_appinsights.connection_string}")
  key_vault_id = azurerm_key_vault.main.id
}
