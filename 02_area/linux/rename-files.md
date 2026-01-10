---
title: "Rename files"
tags: ["linux"]
---

```sh
for file in 03*; do mv "$file" "${file:5}"; done
```
