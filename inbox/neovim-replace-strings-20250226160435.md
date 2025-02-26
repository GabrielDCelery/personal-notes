---
title: Neovim replace strings
author: GaborZeller
date: 2025-02-26T16-04-35Z
tags:
draft: true
---

# Neovim replace strings

## Using sed

Press `:` to start typing command in neovim then type the following.

```sh
%s/wordToReplace/replaceWith/gIc
```

- g - global
- I - case sensitive
- c - prompt confirmation
