---
title: Copy files to docker volume
author: GaborZeller
tags:
  - docker
  - volume
---

1. create the volume

```sh
docker volume create ollama-models
```

2. copy files to the volume

```sh
docker run --rm -v ollama-models:/dest -v /path/to/local/models:/src alpine cp -r /src/. /dest/
```

3. fix permissions

```sh
docker run --rm -v ollama-models:/data alpine chown -R root:root /data
```
