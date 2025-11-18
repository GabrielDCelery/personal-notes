---
title: Linux CLI SSH jump server
description: How to use a bastion host/jump server to connect to a target server
author: GaborZeller
date: 2025-03-26T20-34-26Z
tags:
  - linux-cli
  - ssh
draft: false
---

# Hot to connect to a target server using a jump/bastion host

All it takes is using the `ssh -J` command:

```sh
# Connecting to target machine using a single jump server
ssh -J jumpuser@jumphost targetuser@targethost

# Using a chain of jump hosts
ssh -J user1@jump1.com,user2@jump2.com user3@final-destination.com
```

# Using jump hosts that use a different key

Sometimes the bastion host uses a different key from the one that the target host is using. In this case just use the `-i` flag before every host that you are trying to connect to.

```sh
ssh -i ~/.ssh/jump_key -J jump-user@jump.example.com -i ~/.ssh/target_key target-user@target.example.net
```

# Configuring the SSH config file

If you are connecting to a target server often using a jump server it might be worth just configuring the `~/.ssh/config` file

```sh
# Add to ~/.ssh/config
Host jump-server
    HostName jump.example.com
    IdentityFile ~/.ssh/jump_key
    User jump-user

Host target-server
    HostName target.example.net
    ProxyJump jump-server
    IdentityFile ~/.ssh/target_key
    User target-user

# Connect with:
ssh target-server
```
