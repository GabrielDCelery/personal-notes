---
title: "Terraform state file splitting"
date: 2025-10-21
tags: ["terraform"]
---

# How to split terraform state file

Create separate directories for each logical component:

```txt
infrastructure/
├── network/
│   ├── main.tf
│   ├── outputs.tf
│   └── backend.tf
├── database/
│   ├── main.tf
│   ├── outputs.tf
│   └── backend.tf
├── backend/
│   ├── main.tf
│   ├── outputs.tf
│   └── backend.tf
└── frontend/
    ├── main.tf
    ├── outputs.tf
    └── backend.tf
```

# How to specify backends

Use different `backend.tf` files

Example - Each directory has its own backend configuration (usually S3 + DynamoDB):

```tf

# database/backend.tf
terraform {
  backend "s3" {
    bucket = "my-terraform-state"
    key    = "database/terraform.tfstate"
    region = "us-east-1"
  }
}

# backend/backend.tf
terraform {
  backend "s3" {
    bucket = "my-terraform-state"
    key    = "backend/terraform.tfstate"
    region = "us-east-1"
  }
}
```

# How to share resources between different terraform state files

Use terraform_remote_state data sources to reference outputs from other state files:

```tf
# backend/main.tf
data "terraform_remote_state" "database" {
  backend = "s3"
  config = {
    bucket = "my-terraform-state"
    key    = "database/terraform.tfstate"
    region = "us-east-1"
  }
}

resource "aws_instance" "backend" {
  # Reference database output
  environment = {
    DB_HOST = data.terraform_remote_state.database.outputs.db_endpoint
  }
}

# database/outputs.tf
output "db_endpoint" {
  value = aws_db_instance.main.endpoint
}
```

# Tips and tricks

## Common Split Patterns

1. **By layer:** network → database → application → monitoring
2. **By lifecycle:** rarely-changed (VPC) vs frequently-changed (app config)
3. **By team:** platform team (infra) vs app team (services)
4. **By environment:** Use same structure but different state keys/workspaces

## Benefits

- **Smaller blast radius:** Changes to backend won't risk database
- **Faster operations:** Each terraform apply is smaller
- **Better permissions:** Different teams can own different states
- **Parallel development:** Teams can work independently

## Considerations

- **Dependencies:** Plan the order (network → database → backend)
- **Complexity:** More moving parts to coordinate
- **State drift:** Harder to see full infrastructure at once

# References

[^1] [The terraform_remote_state Data Source](https://developer.hashicorp.com/terraform/language/state/remote-state-data)
[^2] [Managing Terraform State – Best Practices & Examples](https://spacelift.io/blog/terraform-state)
