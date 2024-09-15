---
title: Create new .zshrc file after accidentally deleting old
author: Gabor Zeller
date: 2024-09-15T11:46:13Z
tags: ['linux']
draft: true
---

# Create new .zshrc file after accidentally deleting old

If you accidentally deleted your .zshrc file and want to create a new one:

1. Option A

If you have `oh-my-zsh` installed then it has a default template that you can use to create a new `.zshrc`.

```sh
cp ~/.oh-my-zsh/templates/zshrc.zsh-template ~/.zshrc
```

2. Option B

Run the following to set up a bare minimum skeleton

```sh
echo 'export ZSH=$HOME/.oh-my-zsh
ZSH_THEME="robbyrussell"' > ~/.zshrc
```
