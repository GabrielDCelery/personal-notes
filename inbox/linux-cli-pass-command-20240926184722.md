---
title: Linux CLI pass command
author: GaborZeller
date: 2024-09-26T18-47-22Z
tags: linux
draft: true
---

# Linux CLI pass command

Under the hood pass uses gpg to generate passwords so we need a key pair, so `gpg` is a pre-requisite.

## Initialise a password store

```sh
pass init <gpgkeyid>
pass init -p <storename> <gpgkeyid> # will initialise a named password store that will appear as a folder 
```

## Show password store

```sh
pass show # shows the layout the password store
pass show <storenam> # works like pass but shows layout of specific store 
```

## Initialise password store as git repository

```sh
pass git init
```

## Generate new password

```sh
pass insert <passwordname> # will prompt for a password
pass generate <passwordname> # will auto generate a password
# passwords can be generated in a namespaced way,
# e.g. pass generate github/personal which will create a github folder and the personal password inside it 
# you can also do multiple levels of nesting e.g. pass generate aws/personal/account
pass generate <passwordnamewithnamespace>
```

## Adding metadata to passwords

```sh
pass edit <passwordname>
# opens up the password and you can edit the file and add an extra line to it e.g.
# IUHjdjasd$43
# email: someemail@gmail.com
```

## Finding passwords

```sh
pass find <word>
pass grep <pattern> # allows to find a password even by metadata
```

## Get password or copy to clipboard

```sh
pass show <passwordname>
pass show -c <passwordname> # will copy the first line (the password) to the clipboard
```

## Delete passwords

```sh
pass rm <passwordname>
```

## Using git with the password store

```sh
pass git log # show git history
pass git revert HEAD # revert password store to HEAD (basically restore deleted password)
pass git remote add origin <originlocation> # add a remote repo to our password store
pass git push origin main # push to remote
git clone <repository> .password-store # clones our store
```

## Transferring password store to an other machine

For these we will need to clone the password store to the machine where we want to use the store.

```sh
git clone <repository> .password-store # clones our store
```

Then we need to convert our gpg keys (both public and private) to a file.

```sh
mkdir exported-gpg-keys && cd exported-gpg-keys
gpg --output public.pgp --armor --export someemail@someprovider.com
gpg --output private.pgp --armor --export-secret-key someemail@someprovider.com
```

- `--output` - specifies the output file
- `--armor` - file gets saved in an armored format (base64)
- `--export` - allows us to export the key via an identifier (email) rather than its ID

After this transfer the keys to the machine where you cloned the password store.

```sh
scp -r <username>@<hostname>:<foldertocopyfrom> .
```

And import the gpg key.

```sh
gpg --import private.pgp
gpg --import public.pgp
```

[!WARNING] This is only sufficient for decrypting with the gpg key, if you want to use it for encryption you have to do the following:

```sh
gpg --edit-key <keyname>
# in the interactive mode then type `trust` to elevate the trust level of the key (level 5)
```

## Tips and tricks on how to use pass

### Setting environment variables

The below trick is useful to prevent accidentially having passwords in your shell history

```sh
export GITHUB_TOKEN=$(pass show github/api/token)
```

### Using an alias to set access credenitals as part of the command

You can set an alias for a command that to tend to use to set environment variables as part of that command automatically.

```sh
alias aws="AWS_ACCESS_KEY_ID=$(pass show aws/access_id) AWS_SECRET_ACCESS_KEY=$(pass show aws/secret_key) aws"
# then use the command
aws lambda list-functions --region=eu-west-2 # your aws command will execute the aliased command with the passwords
```

