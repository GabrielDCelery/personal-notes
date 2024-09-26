---
title: Useful gpg commands
author: GaborZeller
date: 2024-09-26T19-15-38Z
tags:
draft: true
---

# Useful gpg commands

## Create gpg key

```sh
gpg --gen-key
```

## List gpg keys

```sh
gpg --list-key
```

## Edit gpg key

```sh
gpg --edit-key <keyid>
```

The above command will enter to an interactive mode where use can use `help` to get a list of available commands and you can use things like `expire` to change the expiry date of your key or `save` to save your changes.


