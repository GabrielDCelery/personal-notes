---
title: Manage python virtual environment
author: GaborZeller
date: 2024-09-29T17-37-54Z
tags:
draft: true
---

# Manage python virtual environment


## Setting up a virtual environment

### Create virtual environment

```sh
cd <dirwhereyourpythonprojectlives>
python3 -m venv .venv
```

or 

```sh
python3.11 -m venv .venv # If you want to use a specific version of python
```

This will create a `.env` (convention to name it like that) folder in your project. After creating the virtual environment it has to be activated.

```sh
source .env/bin/activate
```

You can check if the virtual environment was activated correctly by using `which`.

```sh
which python # Should pring something like : /home/gzeller/practice/nltk-practice/.env/bin/python
```

### Installing dependencies into virtual environment

```sh
pip install <packageyouwanttoinstall>
```

### Deactivate virtual environment

Once you are done using your virtual environment you can deactivate it from anywhere running:

```sh
deactivate
```

## Using virtual environment with poetry

```sh
poetry install # This will install your dependencies
```

```sh
poetry shell # creates and runs virtual environment also automatically selects the python version set in the toml file
poetry env use python3.11 # If you want to use a specific version of python, but this does not create a virtual environment
```

- Poetry uses virtual environments to manage project dependencies
- `shell` command allows to both create and run virtual environment 
- `env use` command allows to specify which version to use for current project

