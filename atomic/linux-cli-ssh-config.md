---
title: Linux CLI SSH config
description: Tips and tricks on how to use the ~/.ssh/config file
author: GaborZeller
date: 2025-03-26T20-34-26Z
tags:
  - linux-cli
  - ssh
draft: false
---

# Basic use of the ssh config file

The `~/.ssh/config` file is really useful for configuring connections that you use commonly. Especially with `TUI tools` like `sshs`. A common configuration looks something like this:

```sh
# How to set the ~/.ssh/config file
Host myserver
  HostName server.example.com
  User username
  IdentityFile ~/.ssh/id_rsa
  Port 2222

# How to use it
ssh myserver
```

# Advanced host matching patterns

The ssh config file gets evaluated from top to bottom, until a match is found. The host section can have multiple patterns listed in comma-separated manner and you can also use special characters lik `*` for any character or `?` for exactly one character. `!` used for negation.

```sh
Host * # matches all hosts
Host *.com # any host with the .com domain
Host 192.168.1.? # matches all hosts in 192.168.1.[0-9]
Host !192.168.1.1, 192.168.1.* # all hosts except for gateway
```
