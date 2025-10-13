---
title: "Order fields for json or yaml"
date: 2025-10-13
tags: ["yaml", "json"]
---

# The problem

I wanted to be able to sort the fields on json and yaml objects because sometimes infrastructure tools tends to reorder fields and makes comparing raw files difficult

## How to order

```sh
yq -o json '.' somefile.yaml | jq --sort-keys | yq -P > somefile.sorted.yaml
```

- `-o json` specify output
- `'.'` path expression that mean `select the entire document`
- `-P` pretty print
