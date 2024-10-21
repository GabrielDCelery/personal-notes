---
title: Count number of files in directory in linux
author: GaborZeller
date: 2024-10-21T21-01-37Z
tags:
draft: true
---

# Count number of files in directory in linux

To count the number of files in a directory run the following:

```sh
ls | wc -l
ls -a | wc -l # to include hidden files
```

If you want to count all the files in a directory (including subdirectories use the find command)

```sh
find . -type f | wc -l
```
