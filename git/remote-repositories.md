---
title: Remote repositories
author: Gabor Zeller
date: 2024-09-01T16:46:13Z
tags: ['git']
draft: true
---

# Remote repositories

## Remote Git and fetch

In git a remote repository is `just a repository somewhere else`. This does not need to be GitHub it can even be a folder on a remote server or on your machine.

You can see what remote has been configure for your repo by:

```sh
git remote -v

origin	git@github.com:GabrielDCelery/personal-notes.git (fetch)
origin	git@github.com:GabrielDCelery/personal-notes.git (push)
```

If you have a repo you can add a remote by:

```sh
git remote add <name of remote> <path to remote>
```

Where the name of the remote by convention is `origin`. If you have a `fork` of the project you are working on then the fork is the one people call origin and the one true repo that is the true source of truth is what is called `upstream`.

Once you added a remote you can bring in the changes from the remote branch to yours using:

```sh
git fetch
```

But this is not enough. The problem with the above command is that it only pulls in the changes from the remote repository but our branch is still divergent from the origin. We have not merged it in yet.

```sh
git log --oneline

fatal: your current branch 'trunk' does not have any commits yet
```

This is happenning because we fetched the remote and we got a bunch of branches pulled into our repo, but our local repo has no commits yet, so we have no branches effectively. In order to bring them inline we need to run 

```sh
git checkout <mainline branch>

git pull origin <mainline branch>

or

git merge origin/<mainline branch>
```

## Pull

Continuing the previous example after adding a change to the remote and trying to pull we run into an issue.

```sh
git pull

There is no tracking information for the current branch.
Please specify which branch you want to merge with.
See git-pull(1) for details.

    git pull <remote> <branch>

If you wish to set tracking information for this branch you can do so with:

    git branch --set-upstream-to=origin/<branch> trunk
```

Just because we have a branch named trunk on our local repo and one named trunk on the remote as well git won't just automatically assume that they are the same. So we have to set the tracking information by running. After then pulling should work.

```sh
git branch --set-upstream-to=origin/trunk trunk
```

So as you can see `GitHub` is relly just someone else's computer using the same core principals to share and sync repositories.

One of the other configurations you can have for pull is to do it with a `rebase` strategy. Either you 

1. Make your pulls with a flag `git pull --rebase`
2. Set a config to do it all the time with `git config --add  --global pull.rebase true`
