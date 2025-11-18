---
title: Unix vs TCP/IP socket
author: GaborZeller
date: 2024-09-22T22-46-15Z
tags:
draft: true
---

# Unix vs TCP/IP socket

A Unix socket is an inter-process mechanism that allows bidirectional communication between processes running on the same machine. They are faster and lighter than IP sockets.

A TCP/IP socket is a mechanism to allow communication between processes over a network.

#### Permission differences

Unix sockets use system level permissions, TCP/IP can be controlled at the packet filter level.

#### See unix sockets on Linux

```sh
netstat -a -p --unix
```

