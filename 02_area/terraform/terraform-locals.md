---
title: "Terraform locals"
date: 2025-10-20
tags: ["terraform"]
---

# Locals vs variables

Locals are scoped within a module and variables are accesible globally. Think of them as local temp variables scoped to a function.

# Common use case

Local values are `immutable` once they are declared so a common use case is to calculate some dynamic value then use it as reference across the code.

Locals are primarily used to `manipulate information` from an other terraform component.

Example:

```tf
locals {
  even_numbers = [for i in [1, 2, 3, 4, 5, 6] : i if i % 2 == 0]
}

```

> [!WARNING]
> Locals have to be declared in the blocks called locals (plural) but referenced as local.even_numbers (singular)

# Tips and tricks

## Using locals as object

```tf
locals {
  resource_tags = {
    project_name = "mytest",
    category     = "devresource"
  }
}

resource "aws_iam_role" "myrole" {
  name = "my_role"
  ...

  tags = local.resource_tags
}
```

## Using locals to set defaults

```tf
variable "res_tags" {
  type = map(string)
  default = {
    dept = "finance",
    type = "app"
  }
}

locals {
  all_tags = {
    env       = "dev",
    terraform = true
  }
  applied_tags = merge(local.all_tags, var.res_tags)
}


resource "aws_s3_bucket" "tagsbucket" {
  bucket = "tags-bucket"
  acl    = "private"

  tags = local.applied_tags
}


output "out_tags" {
  value = local.applied_tags
}
```
