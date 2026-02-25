---
title: VCS Dependencies in Python
date: 2026-02-25
tags:
---

Python supports installing packages directly from version control systems using [PEP 508](https://peps.python.org/pep-0508/) URL specifiers. This is how we install internal packages from this monorepo.

## Anatomy of a VCS dependency

```
ds-tools @ git+https://github.com/mycompany/services-myservice-v2.git@pkg/ds_tools/v0.0.8#subdirectory=pkg/ds-tools
│           │   │                                                 │                   │
│           │   └─ repo URL                                       │                   └─ subdir within repo
│           └─ VCS scheme                                         └─ git tag/ref to pin to
└─ package name
```

| Part                         | Value                                            | Meaning                                          |
| ---------------------------- | ------------------------------------------------ | ------------------------------------------------ |
| Package name                 | `ds-tools`                                       | How pip registers the installed package          |
| VCS scheme                   | `git+https://`                                   | Tells pip to clone via git over HTTPS            |
| Repo                         | `github.com/mycompany/services-myservice-v2.git` | The repository to clone                          |
| `@pkg/ds_tools/v0.0.8`       | Git tag                                          | Pins to a specific release — reproducible builds |
| `#subdirectory=pkg/ds-tools` | Fragment                                         | The package root within the repo                 |

## What happens at install time

1. pip/uv clones `services-myservice-v2`
2. Checks out the tag `pkg/ds_tools/v0.0.8`
3. Navigates into `pkg/ds-tools/`
4. Installs the package found there (`pyproject.toml` or `setup.py`)

## VCS schemes

| Scheme           | When to use                                 |
| ---------------- | ------------------------------------------- |
| `git+https://`   | CI and most cases — authenticates via token |
| `git+ssh://git@` | Local dev — authenticates via SSH key       |
| `git+http://`    | Insecure, avoid                             |

`git+ssh` looks like:

```
ds-tools @ git+ssh://git@github.com/mycompany/services-myservice-v2.git@pkg/ds_tools/v0.0.8#subdirectory=pkg/ds-tools
```

## Bumping a version

1. Merge your changes to `main`
2. Create and push a new git tag: `git tag pkg/ds_tools/v0.0.9 && git push origin pkg/ds_tools/v0.0.9`
3. Update the ref in the consuming service's `pyproject.toml`: change `v0.0.8` → `v0.0.9`
4. Re-lock: `uv lock` or `poetry lock`
