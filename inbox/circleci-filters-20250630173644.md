---
title: CircleCI filters
author: GaborZeller
date: 2025-06-30T17-36-44Z
tags:
draft: true
---

# CircleCI filters

## Using conditions to execute different commands

```yaml
commands:
  dagger-install:
    parameters:
      docker-executor:
        type: boolean
    steps:
      - run:
          name: Install Dagger CLI on Docker Executor
          command: curl -L https://dl.dagger.io/dagger/install.sh | <<# parameters.docker-executor >>BIN_DIR=$HOME/.local/bin <</ parameters.docker-executor >><<^ parameters.docker-executor >>sudo BIN_DIR=/usr/local/bin<</ parameters.docker-executor >> sh
```

- `<<# parameters.docker-executor >>` - Start a section that's included if docker-executor is true
- `<</ parameters.docker-executor >>` - End the section
- `<<^ parameters.docker-executor >>` - Start a section that's included if docker-executor is false

So when the command runs, it resolves to one of these two actual commands:

When `docker-executor: true`:

```bash
curl -L https://dl.dagger.io/dagger/install.sh | BIN_DIR=$HOME/.local/bin sh
```

When `docker-executor: false`:

```bash
curl -L https://dl.dagger.io/dagger/install.sh | sudo BIN_DIR=/usr/local/bin sh
```

The key differences:

1. With `docker-executor: true`
   - No `sudo` (because Docker containers typically don't have sudo)
   - Uses `$HOME/.local/bin` as installation directory (user-specific directory)

2. With `docker-executor: false`
   - Includes `sudo` (because machine executors have sudo access)
   - Uses `/usr/local/bin` as installation directory (system-wide directory)

## Using anchor to create reusable filters

In yaml the `&` symbol is an anchor that can be used to reference reusable config later with `*`.

```yaml
workflows:
  deploy-dev-and-test:
    jobs:
      - test:
          filters: &build-filters
            branches:
              only:
                - master
            tags:
              ignore: /.*/
      - build:
          filters: *build-filters
      - publish:
          filters: *build-filters
          requires:
            - build
            - test
```
