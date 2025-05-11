---
title: How to create SSO for Cloudflare using Azure Entra ID
author: GaborZeller
date: 2025-05-11T21-52-31Z
tags:
draft: true
---

# How to create SSO for Cloudflare using Azure Entra ID

Here's a step-by-step guide on how to set up Azure Entra ID (formerly Azure AD) as an identity provider for Cloudflare:

1. **In Azure Portal:**

```
a. Register a new application:
- Go to Azure Portal → Azure Active Directory
- Select "App registrations" → "New registration"
- Give your app a name
- Set redirect URI (you'll get this from Cloudflare)
- Click "Register"

b. Get required information:
- Application (client) ID
- Directory (tenant) ID
- Create a new client secret
```

2. **In Cloudflare Dashboard:**

```
a. Go to "Settings" → "Authentication"

b. Click "Add" under "SSO"

c. Choose "Azure AD" as the identity provider

d. Enter the configuration details:
- Azure AD Domain
- Application ID (Client ID from Azure)
- Application Secret (Client Secret from Azure)
- Azure AD Directory ID (Tenant ID)

e. Copy the Callback URL/ACS URL from Cloudflare to use in Azure
```

3. **Configure Azure AD Claims (Optional):**

```
a. In Azure Portal, go to your app registration
b. Select "Token configuration"
c. Add any required claims for Cloudflare
```

4. **Configure Access Groups (Optional):**

```
a. In Cloudflare:
- Set up access groups
- Define rules for who can access what

b. In Azure AD:
- Create corresponding security groups
- Assign users to groups
```

Example Configuration in Cloudflare:

```json
{
  "type": "azure",
  "config": {
    "client_id": "your_azure_client_id",
    "client_secret": "your_azure_client_secret",
    "directory_id": "your_azure_tenant_id",
    "auth_url": "https://login.microsoftonline.com/{directory_id}/oauth2/v2.0/authorize",
    "token_url": "https://login.microsoftonline.com/{directory_id}/oauth2/v2.0/token"
  }
}
```

5. **Test the Configuration:**

```
a. Log out of Cloudflare
b. Try logging in using Azure AD credentials
c. Verify that:
   - Login works
   - Correct permissions are applied
   - User information is properly passed
```

6. **Troubleshooting Tips:**

```
- Check redirect URIs match exactly
- Verify client ID and secret
- Ensure required permissions are granted in Azure AD
- Check network connectivity
- Review Azure AD sign-in logs for errors
```

Best Practices:

```
1. Use secure client secrets
2. Implement proper access controls
3. Regular audit of access permissions
4. Monitor authentication logs
5. Set up backup admin access
6. Document the configuration
```

Error Handling:

```python
Common issues and solutions:
- Invalid redirect URI: Double-check URLs match exactly
- Authentication failed: Verify credentials and permissions
- Claims not received: Check token configuration
- Group membership issues: Verify group assignments
```

Additional Features:

```
1. Multi-factor authentication (MFA)
2. Conditional access policies
3. Role-based access control (RBAC)
4. Session management
5. Audit logging
```

Remember to:

- Keep client secrets secure
- Regularly review access permissions
- Monitor authentication logs
- Test the setup thoroughly
- Have a backup access method
- Document the configuration

This setup provides secure SSO (Single Sign-On) access to Cloudflare using Azure Entra ID as the identity provider.

> how to use azure entra identity provider to log in to cloudflare using terraform?
> I'll help you set up Azure Entra ID (formerly Azure AD) as an identity provider for Cloudflare using Terraform. Here's a step-by-step guide:

1. First, you'll need to register an application in Azure Entra ID:

```hcl
# Azure AD Application Registration
resource "azuread_application" "cloudflare_sso" {
  display_name = "Cloudflare SSO"
  web {
    homepage_url  = "https://dash.cloudflare.com"
    redirect_uris = ["https://dash.cloudflare.com/auth/oidc/callback"]
  }

  group_membership_claims = ["SecurityGroup"]
  optional_claims {
    id_token {
      name                  = "groups"
      essential            = false
      additional_properties = []
    }
  }
}

# Create Service Principal
resource "azuread_service_principal" "cloudflare_sso" {
  application_id = azuread_application.cloudflare_sso.application_id
}
```

2. Create an application secret:

```hcl
resource "azuread_application_password" "cloudflare_sso" {
  application_object_id = azuread_application.cloudflare_sso.object_id
  display_name         = "Cloudflare SSO Secret"
  end_date_relative    = "8760h" # 1 year
}
```

3. Configure the Cloudflare Access Identity Provider:

```hcl
resource "cloudflare_access_identity_provider" "azure_oauth" {
  account_id = var.cloudflare_account_id
  name       = "Azure SSO"
  type       = "azure"
  config {
    client_id     = azuread_application.cloudflare_sso.application_id
    client_secret = azuread_application_password.cloudflare_sso.value
    directory_id  = data.azuread_client_config.current.tenant_id
  }
}
```

4. (Optional) Create Azure AD groups and assign them:

```hcl
resource "azuread_group" "cloudflare_admins" {
  display_name     = "Cloudflare Administrators"
  security_enabled = true
}

resource "azuread_group_member" "cloudflare_admins" {
  group_object_id  = azuread_group.cloudflare_admins.object_id
  member_object_id = data.azuread_user.example.object_id
}
```

5. Required providers and data sources:

```hcl
terraform {
  required_providers {
    azuread = {
      source  = "hashicorp/azuread"
      version = "~> 2.0"
    }
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 4.0"
    }
  }
}

provider "azuread" {
  # Configuration options
}

provider "cloudflare" {
  api_token = var.cloudflare_api_token
}

data "azuread_client_config" "current" {}
```

6. Variables you might need:

```hcl
variable "cloudflare_api_token" {
  type        = string
  description = "Cloudflare API Token"
  sensitive   = true
}

variable "cloudflare_account_id" {
  type        = string
  description = "Cloudflare Account ID"
}
```

7. (Optional) Create an Access Application to protect:

```hcl
resource "cloudflare_access_application" "example_app" {
  account_id                = var.cloudflare_account_id
  name                      = "Example Application"
  domain                    = "app.example.com"
  type                     = "self_hosted"
  session_duration         = "24h"
  auto_redirect_to_identity = true
}

resource "cloudflare_access_policy" "example_policy" {
  account_id     = var.cloudflare_account_id
  application_id = cloudflare_access_application.example_app.id
  name           = "Example Policy"
  precedence     = "1"
  decision       = "allow"

  include {
    group = [azuread_group.cloudflare_admins.display_name]
  }
}
```

After applying this configuration:

1. Configure your DNS records to point to your application
2. Users will be able to log in to Cloudflare using their Azure Entra ID credentials
3. Access will be granted based on group membership

Important notes:

1. Make sure you have the necessary permissions in both Azure and Cloudflare to create these resources
2. Keep your client_secret secure and never commit it to version control
3. The redirect URI in Azure should match what Cloudflare expects
4. You may need to adjust the group claims and scopes based on your specific requirements

To apply this configuration:

```bash
terraform init
terraform plan
terraform apply
```

Remember to store your sensitive variables in a secure manner, such as using environment variables or a secrets management solution.
