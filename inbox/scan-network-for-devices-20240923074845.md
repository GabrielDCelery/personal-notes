---
title: Scan network for devices
author: GaborZeller
date: 2024-09-23T07-48-45Z
tags:
draft: true
---

# Scan network for devices

```sh
sudo nmap -sn -n 192.168.1.0/24
```

`nmap` - network exploration tool
`sn` - no port scan (type of the scan)
`n` - disable DNS host resolution
`192.168.1.0/24` - performs a scan on the entire 192.168.1.x subnet
