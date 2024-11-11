---
title: How to use the linux ls command
author: GaborZeller
date: 2024-11-11T15-04-39Z
tags:
draft: true
---

# How to use the linux ls command

## Search for files that were modified withing the past X

```sh
find . -type f -mtime -X # files modified withing the past X days
find . -type f -mtime -Xs # files modified within the past X seconds

s       second
m       minute (60 seconds)
h       hour (60 minutes)
d       day (24 hours)
w       week (7 days)

find . -mtime +X # files modified beyond X days
```
