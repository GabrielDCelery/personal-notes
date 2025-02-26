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

## Using \*

1. Navigate over a word that you want to replace
2. Press `*` (select word under visual cursos)
3. Press `gn` (search forward for the last used search pattern)
4. Press `c` to replace selection
5. Type out replacement
6. Press `n` to seach forward
7. Press `.` to repeat last change
