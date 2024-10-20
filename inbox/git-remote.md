---
title: Git remote repositories
author: Gabor Zeller
date: 2024-09-01T16:46:13Z
tags: ['git']
draft: true
---

# Git remote repositories

## Viewing where is the authoratative source for our repository

In git a remote repository is `just a repository somewhere else`. This does not need to be GitHub it can even be a folder on a remote server or on your machine.

You can see what remote has been configure for your repo by:

```sh
git remote -v

origin	git@github.com:GabrielDCelery/personal-notes.git (fetch)
origin	git@github.com:GabrielDCelery/personal-notes.git (push)
```

## Adding a remote repository

If you have a repo you can add a remote by:

```sh
git remote add <name of remote> <path to remote>
```

Where the name of the remote by convention is `origin`. If you have a `fork` of the project you are working on then the fork is the one people call `origin` and the one true repo that is the true source of truth is what is called `upstream`.

Once you added a remote you can bring in the changes from the remote branch to yours using:

```sh
git fetch
```

But this is not enough. The problem with the above command is that it only pulls in the changes from the remote repository but our branch is still divergent from the origin. So we have to sync our local repo with the remote one.

```sh
git checkout <mainlinebranch>
git merge origin/<mainlinebranch>
git branch --set-upstream-to=origin/mainlinebranch mainlinebranch
```
