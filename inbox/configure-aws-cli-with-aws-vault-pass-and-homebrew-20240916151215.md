---
title: Configure aws-cli with aws-vault pass and homebrew
author: GaborZeller
date:
tags:
draft: true
---

# Configure aws-cli with aws-vault pass and homebrew

Install the cli tools with homebrew

```sh
brew install aws-cli
brew install aws-vault
brew install pass
```

Create a gpg key

```sh
gpg --full-generate-key
```

Initialize pass with the gpg key

```sh
pass init -p aws-vault YOUR_GPG_KEY_ID
```

Configure aws-vault to use pass as its backend

```sh
export AWS_VAULT_BACKEND=pass
export AWS_VAULT_PASS_PASSWORD_STORE_DIR=$HOME/.password-store/aws-vault
```

Add your aws credentials to pass
```sh
aws-vault add my-aws-account
```
