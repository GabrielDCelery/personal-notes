---
title: Python how to use poetry
author: GaborZeller
date: 2025-01-24T10-26-04Z
tags:
draft: true
---

# Python how to use poetry

## Initialize poetry

### If the project already exists and does have a toml file

```sh
cd <projectdir>
poetry install
```

### If the project already exists and does not have a toml file

```sh
cd <projectdir>
poetry init
```

### If the project does not exits

The below will create a `.toml` file.

```sh
poetry new <projectname that will also be the dirname>
```

## Adding packages

```sh
poetry add <packagename>
poetry add <packagename>@2.11.2 # add specific version
poetry add <packagename>^2.11.2 # add specific version up until the major verion
poetry add <packagename>~2.11.2 # add specific version up until the minor version
```

## Removing packages

```sh
poetry remove <packagename>
```

## Viewing installed packages

```sh
poetry show
poetry show <packagename>
```

## Activating virtual environment

```sh
poetry env use $(which python3.12) # use poetry with specific python version (THIS CAN BE OMITTED because the next command will read the toml file for the correct version)
eval $(poetry env activate)
poetry install
```

## Disabling virtual environment

```sh
deactivate
```
