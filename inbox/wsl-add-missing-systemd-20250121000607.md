---
title: WSL add missing systemd
author: GaborZeller
date: 2025-01-21T00-06-07Z
tags:
draft: true
---

# WSL add missing systemd

Make sure your wsl is up-to-date

```sh
wsl --version
```

Create or edit a file at `/etc/wsl.conf` with the following content

```sh
[boot]
systemd=true
```

Restart WSL from powershell

```sh
wsl --shutdown
wsl
```

Check if systemctl is up

```sh
sudo systemctl status
```
