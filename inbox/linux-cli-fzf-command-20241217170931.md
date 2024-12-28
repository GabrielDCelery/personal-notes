---
title: Linux CLI fzf command
author: GaborZeller
date: 2024-12-17T17-09-31Z
tags:
draft: true
---

# Linux CLI fzf command

## How to search files with preview mode

```sh
fzf --preview "cat {}" # preview files with cat
fzf --preview "bat --color=always {}" # preview colorized version
```

## How to pipe results of a command into fzf and execute an other command in preview mode

```sh
docker ps | fzf --preview "docker inspect {1}"
```
