---
title: Zip downloaded files in directory
author: GaborZeller
date: 2024-10-21T22-00-45Z
tags:
draft: true
---

# Zip downloaded files in directory

```sh
find ./downloads -type f -name "*.csv" -execdir zip "{}.zip" "{}" \;
```
