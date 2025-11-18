---
title: "Terraform map vs object"
date: 2025-10-20
tags: ["terraform"]
---

# Differences between maps and objects

## Map

- Homogeneous collection - all values must be the same type
- Dynamic keys - keys are determined at runtime, not predefined
- Accessed dynamically - map_var["any_key"]

```tf
variable "instance_sizes" {
  type = map(string)
}
```

Example:

```tf
instance_sizes = {
  "web" = "t2.micro"
  "db" = "t2.small"
  "api" = "t2.medium"
}
```

## Object

- Heterogeneous collection - values can be different types
- Fixed schema - attributes are predefined in the type definition
- Accessed by attribute - obj_var.attribute_name

```tf
variable "server_config" {
  type = object({
    hostname = string
    port = number
    is_enabled = bool
  })
}
```

Example:

```tf
server_config = {
  hostname = "app.example.com"
  port = 8080
  is_enabled = true
}


```

## Key Differences Summary

| Feature     | Map                         | Object                         |
| ----------- | --------------------------- | ------------------------------ |
| Keys        | Dynamic, any name           | Fixed, predefined              |
| Value types | All same type               | Can be different types         |
| Schema      | Flexible size               | Fixed structure                |
| Use case    | Unknown keys at design time | Known structure at design time |

## Tips and tricks when using maps

1. **Use for expressions with clear key-value mappings:** When converting a list to a map, ensure that each element in the list contains enough structure to derive a unique key.
2. **Ensure key uniqueness:** Keys in a map must be unique. When using a property from list elements as the map key (e.g., user.id), validate or sanitize input if necessary.
3. **Handle complex structures with object unpacking:** If the list contains objects, use property unpacking to map meaningfully
4. **Use conditionals cautiously in for expressions:** Include filters only when relevant, to avoid introducing sparse maps or unexpected null values.
5. **Prefer clarity over compactness:** Avoid deeply nested or overly concise expressions that hinder maintainability.

# References

[^1] [Terraform Map Variable – What It is & How to Use](https://spacelift.io/blog/terraform-map-variable)
