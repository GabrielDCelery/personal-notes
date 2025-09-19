---
title: "How to remove empty lines in a linux file"
date: 2025-09-19
tags: ["cli", "linux"]
---

The way to solve the problem is to use `grep`, with the `-v` flag which does an inverse search and we want to do it on an empty line pattern which is covered by `^$`.

```sh
grep -v '^$' file.txt
```
