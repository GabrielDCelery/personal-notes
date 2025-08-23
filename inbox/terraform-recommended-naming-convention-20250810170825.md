---
title: Terraform recommended naming convention
author: GaborZeller
date: 2025-08-10T17-08-25Z
tags:
draft: true
---

# Terraform recommended naming convention

## Current Issues

1. **Inconsistent Terraform resource naming**: Using "homelab" in secrets but "tf_remote_state" in bootstrap
2. **Generic resource group names**: "secrets" and "bootstrap" are too generic
3. **Inconsistent Azure resource naming**: "gazehomelabsecrets" vs "gazehomelabtfstate"
4. **Key Vault naming**: Doesn't follow Azure best practices (no separators, hard to read)

## Recommended Naming Convention

### Terraform Resource Names

Use descriptive, consistent names that reflect the resource purpose:

```hcl
# Current
resource "azurerm_resource_group" "homelab" {
resource "azurerm_key_vault" "homelab" {

# Suggested
resource "azurerm_resource_group" "secrets" {
resource "azurerm_key_vault" "secrets" {
```

### Azure Resource Names

Follow a consistent pattern: `{prefix}-{environment}-{purpose}-{resource-type}`

```hcl
# Current
resource "azurerm_resource_group" "secrets" {
  name = "secrets"
}

resource "azurerm_key_vault" "secrets" {
  name = "gazehomelabsecrets"
}

# Suggested
resource "azurerm_resource_group" "secrets" {
  name = "rg-homelab-secrets"
}

resource "azurerm_key_vault" "secrets" {
  name = "kv-homelab-secrets"
}
```

## Complete Improved Configuration

```hcl
resource "azurerm_resource_group" "secrets" {
  name       = "rg-homelab-secrets"
  location   = var.azure_region
  managed_by = var.azure_homelab_subscription_id
}

resource "azurerm_key_vault" "secrets" {
  name                          = "kv-homelab-secrets"
  resource_group_name           = azurerm_resource_group.secrets.name
  location                      = azurerm_resource_group.secrets.location
  tenant_id                     = var.azure_homelab_tenant_id
  sku_name                      = "standard"
  soft_delete_retention_days    = 90
  purge_protection_enabled      = false
  public_network_access_enabled = true
}
```

## Benefits of This Approach

1. **Consistency**: All resources follow the same naming pattern
2. **Clarity**: Resource type prefixes (rg-, kv-) make it immediately clear what each resource is
3. **Scalability**: Easy to add environment suffixes later (e.g., kv-homelab-secrets-dev)
4. **Azure Best Practices**: Follows Microsoft's recommended naming conventions
5. **Readability**: Hyphens make names easier to read and parse

## Additional Recommendations

- Consider adding environment variables if you plan to have dev/staging/prod environments
- Use consistent abbreviations (rg = resource group, kv = key vault, st = storage account)
- Keep names under Azure's character limits (Key Vault: 24 characters, Resource Group: 90 characters)

This naming convention will make your infrastructure more maintainable and easier to understand at scale.

---

- Datetime: 2025-08-10 18:00:56
- Model: claude/claude-sonnet-4-20250514
- Selected files:
  - /home/gaze/projects/github/personal/personal-homelab/terraform/deployments/secrets/main.tf

- Selected code:

```tf /home/gaze/projects/github/personal/personal-homelab/terraform/deployments/secrets/main.tf
resource "azurerm_resource_group" "homelab" {
  name       = "secrets"
  location   = var.azure_region
  managed_by = var.azure_homelab_subscription_id
}

resource "azurerm_key_vault" "homelab" {
```

> in the previoud block what do rg- and kv- prefixes mean?

✓ Task Completed

The `rg-` and `kv-` prefixes are **resource type abbreviations** that indicate what kind of Azure resource it is:

- **`rg-`** = **Resource Group**
- **`kv-`** = **Key Vault**

## Why Use Resource Type Prefixes?

This is a Microsoft-recommended naming convention that provides several benefits:

1. **Instant Recognition**: You can immediately tell what type of resource it is just from the name
2. **Organization**: When viewing resources in Azure Portal, they group together alphabetically by type
3. **Consistency**: Follows Azure's official naming guidelines
4. **Scalability**: Makes it easy to manage hundreds of resources across different types

## Common Azure Resource Type Prefixes

Here are standard abbreviations for common Azure resources:

| Resource Type          | Prefix  | Example              |
| ---------------------- | ------- | -------------------- |
| Resource Group         | `rg-`   | `rg-homelab-secrets` |
| Key Vault              | `kv-`   | `kv-homelab-secrets` |
| Storage Account        | `st`    | `sthomelabtfstate`   |
| Virtual Machine        | `vm-`   | `vm-homelab-web01`   |
| Virtual Network        | `vnet-` | `vnet-homelab-main`  |
| Subnet                 | `snet-` | `snet-homelab-web`   |
| Network Security Group | `nsg-`  | `nsg-homelab-web`    |
| Application Gateway    | `agw-`  | `agw-homelab-main`   |
| Load Balancer          | `lb-`   | `lb-homelab-web`     |

## Your Current vs. Suggested Names

```hcl
# Current (unclear what type of resource)
name = "secrets"           # Could be anything
name = "gazehomelabsecrets" # Hard to parse, unclear type

# Suggested (clear resource type)
name = "rg-homelab-secrets"  # Clearly a Resource Group
name = "kv-homelab-secrets"  # Clearly a Key Vault
```

This naming convention is widely adopted in enterprise Azure environments because it makes infrastructure much easier to navigate and manage.
