---
title: "Swap space"
date: 2025-12-29
tags: ["linux"]
---

# Swap space

## To check if there is swap space

```sh
free -h

# outputs
               total        used        free      shared  buff/cache   available
Mem:           458Mi       295Mi        18Mi       3.1Mi       168Mi       163Mi
Swap:          1.0Gi       132Mi       891Mi

```

### Memory (RAM) Row:

total: 458Mi - Total physical RAM installed in the system
used: 295Mi - RAM actively used by processes (programs, services, Docker containers)
free: 18Mi - Completely unused RAM sitting idle (this is VERY low!)
shared: 3.1Mi - RAM shared between processes (like shared libraries)
buff/cache: 168Mi - RAM used for file buffers and disk cache (Linux uses "spare" RAM to speed up disk access). This can be freed if processes need it.
available: 163Mi - RAM that can be made available for new processes without swapping (includes free + reclaimable cache).

> [!INFO]
> available is the most important number - it's what's actually available for use.

### Swap Row:

total: 1.0Gi - Total swap space
used: 132Mi - Data that's been moved from RAM to swap (disk). This means the system ran out of RAM and moved some data to disk.
free: 891Mi - Unused swap space available

## How to create swap space

```sh
# Create a 2GB swap file
# Traditional rule: 2x RAM = ~1GB swap
sudo fallocate -l 2G /swapfile

# Set correct permissions
sudo chmod 600 /swapfile

# Format as swap
sudo mkswap /swapfile

# Enable swap
sudo swapon /swapfile

# Verify it's working
free -h

# To make it permanent (survive reboots):
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab

# Watch how the swap space is being used
watch -n 2 free -h
```
