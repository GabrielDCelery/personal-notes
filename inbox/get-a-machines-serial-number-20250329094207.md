---
title: Get a machine's serial number
author: GaborZeller
date: 2025-03-29T09-42-07Z
tags:
draft: true
---

# Get a machine's serial number

## Windows

```sh
wmic bios get serialnumber
```

## Linux

```sh
sudo dmidecode -t system | grep Serial
```
