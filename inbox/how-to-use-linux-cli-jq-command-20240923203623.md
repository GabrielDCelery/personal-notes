---
title: How to use Linux CLI Jq command
author: GaborZeller
date: 2024-09-23T20-36-23Z
tags: linux
draft: true
---

# How to use Linux CLI Jq command

## Filter array of objects based on field matching value

Use the `jq '.[] | select(<condition>)'` pattern.

```sh
echo '[{"name": "John", "age": 30}, {"name": "Jane", "age": 25}]' | jq '.[] | select(.name == "John")'
```

```sh
# Output
{
  "name": "John",
  "age": 30
}

```
