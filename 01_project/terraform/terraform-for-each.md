---
title: "Terraform for_each"
date: 2025-10-20
tags: ["terraform"]
---

# What is it used for

For generating multiple instances of the same resource.

> [!WARNING]
> for_each can only be used with maps or sets of strings this means lists have to be converted to sets and list of objects should be converted to maps first before using them

# Basic syntax

```tf
resource "<resource type>" "<resource name>" {
  for_each = var.instances
  // Other attributes

  tags = {
    Name = each.<value/key>
  }
}
```

# How to convert list of objects to a map

```tf
variable "instance_objects" {
  type = list(object({
    name = string
    enabled = bool
    instance_type = string
    env = string
  }))
  default = [
  {
    name = "instance A"
    enabled = true
    instance_type = "t2.micro"
    env = "dev"
  },
  {
    name = "instance B"
    enabled = false
    instance_type = "t2.micro"
    env = "prod"
  },
  ]
}


resource "aws_instance" "by_object" {
  for_each = { for inst in var.instance_objects : inst.name => inst }
  ami = "ami-0b08bfc6ff7069aff"
  instance_type = each.value.instance_type

  tags = {
    Name = each.key
    Env = each.value.env
  }
}
```

# External sources

[^1] [How to Use Terraform For_Each Meta-Argument](https://spacelift.io/blog/terraform-for-each)
