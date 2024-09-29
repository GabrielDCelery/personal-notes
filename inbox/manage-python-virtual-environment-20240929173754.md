---
title: Manage python virtual environment
author: GaborZeller
date: 2024-09-29T17-37-54Z
tags:
draft: true
---

# Manage python virtual environment

## Create virtual environment

```sh
cd <dirwhereyourpythonprojectlives>
python3 -m venv .env
```

This will create a `.env` (convention to name it like that) folder in your project. After creating the virtual environment it has to be activated.

```sh
source .env/bin/activate
```

You can check if the virtual environment was activated correctly by using `which`.

```sh
which python # Should pring something like : /home/gzeller/practice/nltk-practice/.env/bin/python
```

## Installing dependencies into virtual environment

```sh
pip install <packageyouwanttoinstall>
```

## Deactivate virtual environment

Once you are done using your virtual environment you can deactivate it from anywhere running:

```sh
deactivate
```


