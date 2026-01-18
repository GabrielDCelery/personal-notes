---
title: Remove password prompt for sudoer
tags:
  - linux
  - sudo
---

# The problem

Don't want to be prompted for user being in the sudo group

# Solution

Go to `/etc/sudoers.d` and add a file e.g. `username` using sudo. This is recommended over using `visudo` and editing the `sudoers` file.

Add the following line to the file:

```txt
username ALL=(ALL) NOPASSWD: ALL
```
