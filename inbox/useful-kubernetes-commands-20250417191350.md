---
title: Useful kubernetes commands
author: GaborZeller
date: 2025-04-17T19-13-50Z
tags:
draft: true
---

# Useful kubernetes commands

## Get a basic yaml file version of a known container

```sh
kubectl run nginx-yaml --image=nginx --dry-run=client -o yaml
```

## Create/apply yaml

```sh
kubectl create # just creates a resource
kubectl apply # detects the diff and applies the changes
```

## Enter a pod to debug

```sh
kubectl exec -it pods/nginx -- /bin/bash
```

## Create deployment

```sh
kubectl create deployment  test --image=httpd --replicas=3
```
