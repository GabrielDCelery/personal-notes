---
title: Multi-architecture Docker images with BuildX for different processors
author: GaborZeller
date: 2024-10-14T22-07-30Z
tags:
draft: true
---

# Multi-architecture Manifest (Legacy)

[!WARNING] This is the legacy method of building containers for multiple architectures. Use BuildX instead

Create multiple builds

```sh
docker build -f Dockerfile.x86 -t my-image:linux-x86
docker build -f Dockerfile.arm -t my-image:arm
```

Then use the manifest command to merge the docker images together

```sh
docker manifest create my-image:tag my-image:linux-x86 my-image:arm
````


# Multi-architecture Docker images with BuildX for different processors

## Login to Docker
```sh
docker login
```

## Check available configurations

```sh
docker buildx ls
```

After running the above command should see the default builder settings

```sh
NAME/NODE     DRIVER/ENDPOINT STATUS  BUILDKIT PLATFORMS
default *     docker
  default     default         running v0.12.5  linux/amd64, linux/amd64/v2, linux/amd64/v3, linux/arm64, linux/riscv64, linux/ppc64le, linux/s390x, linux/386, linux/mips64le, linux/mips64, linux/arm/v7, linux/arm/v6
```

## Create new builder configuration (best practice)

The below command will create a new builder configuration

```sh
docker buildx create --name mybuilder 
```

Then we have to tell buildx to use it

```sh
docker buildx use mybuilder
```

It is recommended to inspect the configuration before using it with the `--bootstrap` flag. Normally when building our images buildx would do the bootstrapping before attempting to build but its reassuring to do that here instead.

```sh
docker buildx inspect --bootstrap
```
```sh
Name:          mybuilder
Driver:        docker-container
Last Activity: 2024-10-14 22:11:20 +0000 UTC

Nodes:
Name:      mybuilder0
Endpoint:  unix:///var/run/docker.sock
Status:    running
Buildkit:  v0.16.0
Platforms: linux/amd64, linux/amd64/v2, linux/amd64/v3, linux/arm64, linux/riscv64, linux/ppc64le, linux/s390x, linux/386, linux/mips64le, linux/mips64, linux/arm/v7, linux/arm/v6
...
```

## Build image using buildx

```sh
docker buildx build --platform linux/amd64,linux/arm64 -t myimagename .
```

