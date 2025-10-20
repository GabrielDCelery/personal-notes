---
title: "Terraform grouping mode"
date: 2025-10-20
tags: ["terraform"]
---

# Grouping Mode Syntax

The key difference is the ... (ellipsis) after the value expression:

## Normal mode (no grouping)

```tf
{ for k, v in collection : v.key => v }
```

## Grouping mode (with ...)

```tf
{ for k, v in collection : v.key => v... }
```

## How Grouping Mode Works

Without grouping mode, if multiple items map to the same key, you get an error. With grouping mode `(...)`, Terraform creates a list of all values for each key.

## Example: Group users by role

```tf
variable "users" {
  type = map(object({
    role = string
  }))
  default = {
    "alice"   = { role = "admin" }
    "bob"     = { role = "developer" }
    "charlie" = { role = "admin" }
    "diana"   = { role = "developer" }
  }

}


locals {
  # Without grouping mode - would error if duplicate keys
  users_by_role_error = {
    for name, user in var.users : user.role => name
    # ERROR: duplicate key "admin"
  }

  # With grouping mode - creates lists
  users_by_role = {
    for name, user in var.users : user.role => name...
  }
}

# Result:
# users_by_role = {
#   "admin"     = ["alice", "charlie"]
#   "developer" = ["bob", "diana"]
# }
```

## Real world example: Security groups

```tf
variable "instances" {
  type = map(object({
    security_group = string
    instance_type  = string
  }))
  default = {
    "web-1"  = { security_group = "sg-web", instance_type = "t2.micro" }
    "web-2"  = { security_group = "sg-web", instance_type = "t2.micro" }
    "db-1"   = { security_group = "sg-db", instance_type = "t2.small" }
  }
}

locals {
  # Group instances by security group
  instances_by_sg = {
    for name, config in var.instances : config.security_group => name...
  }

  # Result:
  # {
  #   "sg-web" = ["web-1", "web-2"]
  #   "sg-db"  = ["db-1"]
  # }
}

# Use grouped data
output "web_instances" {
  value = local.instances_by_sg["sg-web"]
  # Output: ["web-1", "web-2"]
}
```

## Key Points

1. ... enables grouping mode - allows duplicate keys by collecting values into lists
2. Without ... - duplicate keys cause an error
3. Inverted objects - transforming the data structure by swapping what serves as keys vs values
4. Common use case - grouping resources by tags, roles, regions, or other attributes

# References

[^1] [for Expressions](https://developer.hashicorp.com/terraform/language/expressions/for)
