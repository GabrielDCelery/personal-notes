---
title: How to replace word across codebase using cli
author: GaborZeller
date: 2026-03-02
---

```sh
rg -l 'oldword' | xargs sd 'oldword' 'newword'
```
