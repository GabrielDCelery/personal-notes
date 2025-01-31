---
title: Add private ssh key to ssh agent
author: GaborZeller
date: 2025-01-31T20-34-26Z
tags:
draft: true
---

# Add private ssh key to ssh agent

```sh
eval $(ssh-agent -s)
```

```sh
ssh-add ~/.ssh/id_rsa
```
