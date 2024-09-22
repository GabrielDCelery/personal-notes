---
title: What is the docket socket
author: GaborZeller
date: 2024-09-22T22-51-58Z
tags:
draft: true
---

# What is the docket socket

The docker socket lives at `var/run/docker.sock` and us a Unix socket that the docker daemon is listening to. Its purpose is serve as the main entrypoint to the Docker API.

It can be configured as a TCP socket but by default Docker defaults to a Unix socket.

The Docker cli uses this socket to execute docker commands.

#### Docker socket permissions

The docker socket gets a docker group so processes that are part of that group can execute docker commands without being root, but technically the docker daemon got root permissions otherwise it wouldn't be able to access namespace and cgroups.
