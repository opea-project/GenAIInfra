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

# ASK Cluster
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

# Azure Files Storage Account
resource "azurerm_storage_account" "main" {
  name                     = replace(lower("${var.cluster_name}st"), "-", "")
  resource_group_name      = azurerm_resource_group.main.name
  location                 = azurerm_resource_group.main.location
  account_tier             = "Premium"
  account_replication_type = "LRS"
  account_kind            = "FileStorage"
}

# Azure Files Share
resource "azurerm_storage_share" "main" {
  name                 = "aksshare"
  storage_account_id   = azurerm_storage_account.main.id
  quota               = 100
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
    command = "az ask get-credentials --resource-group ${azurerm_resource_group.main.name} --name ${azurerm_kubernetes_cluster.main.name} --overwrite-existing"
  }
  depends_on = [azurerm_kubernetes_cluster.main]
}

# Data source for Azure subscription information
data "azurerm_client_config" "current" {}
