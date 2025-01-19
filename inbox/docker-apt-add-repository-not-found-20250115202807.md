---
title: Docker apt-add-repository not found
author: GaborZeller
date: 2025-01-15T20-28-07Z
tags: docker
draft: true
---

# Docker apt-add-repository not found

apt-add-repository is not in the base Ubuntu image. It has to be installed first

```sh
RUN apt install software-properties-common -y
```
