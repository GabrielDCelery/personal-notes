---
title: "Renew expired gpg key"
date: 2025-09-16
tags: ["gpg"]
---

# Premise

My gpg key has expired on my work computer, needs extending.

## How to fix the issue

First list the keys that I have

```sh
gpg --list-keys
# or gpg --list-secret-keys for secret keys
```

Using the key ID run the following command.

```sh
gpg --edit-key YOUR_KEY_ID
```

Once in the prompt run the following.

```sh
gpg> expire
```

There will be a dynamic prompt asking for how long you want to extend the key.

If there are sub keys first select them and then run the same `expire command`

```sh
gpg> key 1
gpg> expire

```

Once happy run the following to save the changes.

```sh
gpg> save
```
