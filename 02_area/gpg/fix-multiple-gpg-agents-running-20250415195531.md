---
title: Fix multiple gpg agents running
author: GaborZeller
date: 2025-04-15T19-55-31Z
tags:
draft: true
---

# Fix multiple gpg agents running

## The issue

I have encountered this issue after installing `gnupg` on my system using `homebrew`. Turned out I had two versions of the agents running and constantly had the below error popping up on my terminal.

```sh
gpg: WARNING: server 'gpg-agent' is older than us (2.4.4 < 2.4.7)
gpg: WARNING: server 'keyboxd' is older than us (2.4.4 < 2.4.7)
gpg: WARNING: server 'keyboxd' is older than us (2.4.4 < 2.4.7)
```

## Diagnostincs

First was running the following commands, but all the below showed the location of the homebrew installations and the version was 2.4.7

```sh
which gpg
gpg --version
which gpg-agent
gpg-agent --version
```

After that was running the following:

```sh
dpkg -l | grep -i gpg

rc  gpg-agent                       2.4.4-2ubuntu17.2                       amd64        GNU privacy guard - cryptographic agent
ii  gpgv                            2.4.4-2ubuntu17.2                       amd64        GNU privacy guard - signature verification tool
ii  libgpg-error-l10n               1.47-3build2.1                          all          library of error values and messages in GnuPG (localization files)
ii  libgpg-error0:amd64             1.47-3build2.1                          amd64        GnuPG development runtime library
```

## Fixing the issue

1. First of all removed the system default gpg installations:

```sh
sudo apt remove gpg gpg-agent gpg-wks-client gpgconf gpgsm
```

2. Cleaned up GPG home dir

```sh
mv ~/.gnupg ~/.gnupg.backup
```

3. Ensured gpg was properly linked

```sh
brew unlink gnupg && brew link gnupg
```

4. Added the following to my `.zsrhc`

```sh
export GPG_TTY=$(tty)
export GNUPGHOME="$HOME/.gnupg"
```

5. Then reloaded my shell configuration

```sh
source ~/.zshrc
```

6. And created a new trustdb

```sh
gpg --list-keys
```
