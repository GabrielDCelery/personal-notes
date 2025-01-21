---
title: Enter ssh passphrase automatically
author: GaborZeller
date: 2025-01-21T22-32-23Z
tags:
draft: true
---

# Enter ssh passphrase automatically

In order to prevent ssh constantly asking for the passphrase we need to add it to the `ssh agent`.

```sh
eval $(ssh-agent -s) # start the ssh agent
```

```sh
ssh-add ~/.ssh/id_rsa # add the private key to the agent, it will prompt for a password
```

```sh
ssh-add -l # list the added keys to verify our key has been added
```
