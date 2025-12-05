---
title: UV and Python in docker build
author: GaborZeller
date: 2025-12-05
tags:
  - docker
  - python
  - uv
draft: true
---

# The problem

Since UV is a package manager that can also manage virtual environments was curios about best practices for building a docker container

# Solutions

## The simple one

The easiest option is to install everything via `uv` then run the container through uv itself or the virtual env

```Dockerfile
# Add uv to the image
COPY --from=ghcr.io/astral-sh/uv:latest /uv /uvx /bin/

# Install dependencies and the virtual environment
COPY pyproject.toml uv.lock ./
RUN uv sync --frozen --no-dev

# Run the container either via the virtual environment or UV directly
CMD [".venv/bin/python", "main.py"]
#OR
CMD ["uv", "run", "python", "main.py"]
```

# Footnotes

[^1] [Example setup by UV on GitHub](https://github.com/astral-sh/uv-docker-example/tree/main)
[^2] [Dedicated section in UV documentation](https://docs.astral.sh/uv/guides/integration/docker/)
