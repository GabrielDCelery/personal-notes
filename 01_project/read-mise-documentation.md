---
title: "Read mise documentation"
date: 2025-10-06
tags: ["mise"]
---

# Problem

Spend some time reading the mise documentation

## Things that caught my eyes

### Install dependency with specific backend

Mise uses several backends [package managers](https://mise.jdx.dev/dev-tools/backends/) and while mise chooses package managers by default we can also specify which dependency manager we would like to use:

```sh
mise use --global cargo:starship@1.23.0
mise use --global npm:@anthropic-ai/claude-code
```

Result:

```toml
[tools]
"cargo:starship" = "1.23.0"
node = "latest"
"npm:@anthropic-ai/claude-code" = "latest"
poetry = "latest"
```

### Global tasks

Was thinking of writing tasks in the `global` mise file to run scripts

### Run shell scripts using tasks

Instead of inlinig tasks in the toml file we can have them in separate files.

```sh
# have the following in mise-tasks/build

#!/bin/bash
npm run build
```

### Lockfile

It is experimantal feature as of now but using `mise.lock` for projects and ci/cd pipelines can be really useful.

Example lockfile:

```toml
[[tools.go]]
version = "1.24.5"
backend = "core:go"

[[tools.node]]
version = "20.19.5"
backend = "core:node"
```

### Extending path

Extending path with local projects (or maybe even at a global level)

```toml
[env]
_.path = "./node_modules/.bin"
```

### Dependency management

`mise ls` and `mise prune` are pretty useful commands
