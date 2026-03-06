---
title: Encrypt Git repo with git crypt
author: GaborZeller
date: 2026-03-06
tags: git
---

# The problem

How to ecnrypt a git repo for a public repo

# The solution

1. Install git-crypt `https://github.com/AGWA/git-crypt`

```sh
brew install git-crypt
brew install age # for encrypting the key
```

2. Initialise encryption

```
git-crypt init
```

3. Add a `.gitattributes ` file

```.gitattributes
01_project/**/* filter=git-crypt diff=git-crypt
```

4. Export and encrypt the crypt key

```sh
git-crypt export-key ./git-crypt-notes.key
age -p -o git-crypt-notes.key.age git-crypt-notes.key
rm git-crypt-notes.key
```

5. Decrypt

```sh
git clone
age -d -o git-crypt-notes.key  git-crypt-notes.key.age
git-crypt unlock ./git-crypt-notes.key
```

# Remove encryption

```sh
git-crypt unlock ./git-crypt-notes.key

# Remove files from index and re-add without encryption
git rm -r --cached .
git add .
git commit -m "chore: remove git-crypt encryption"

# Remove git-crypt key and config
rm -rf .git/git-crypt

# If encryption history was already pushed force push to clean it up
git push --force
```
