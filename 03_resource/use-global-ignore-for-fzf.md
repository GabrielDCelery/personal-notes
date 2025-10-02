---
title: "Use global ignore in fzf"
date: 2025-10-02
tags: ["fzf"]
---

Victor told me that

You can use a global ~/.ignore file that is obey by ripgrep and fd without having to mess with their defaults or the .gitignore.

His settins are:

```sh
# always include in search
!**/.local/**
!**/mise.*.toml
!**/.env

#always excluse from search
**/package-lock.json
**/yarn.lock
```
