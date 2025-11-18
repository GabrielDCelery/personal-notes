---
title: Linux CLI SSH agent
description: Tips and tricks on how to use the ssh-agent
author: GaborZeller
date: 2025-03-26T20-34-26Z
tags:
  - linux-cli
  - ssh
draft: false
---

# Switching on SSH agent

The SSH agent is a secure in memory container for private keys so the user does not need to repeatedly type passwords when accessing remote servers. To turn it on:

```sh
eval "$(ssh-agent -s)"
```

The above will run `ssh-agent -s` which will start a new SSH agent and outputs a series of variables that other SSH tools use to communicate with the agent. The `eval` command then executes these outputs so the environment variables get set.

[!WARNING] This will only set the agent for the current shell session

# Adding and removing private SSH key to ssh agent

Once the SSH agent is running you can add any number of private keys with the following command:

```sh
ssh-add ~/.ssh/id_rsa
ssh-add ~/.ssh/personal_key
ssh-add ~/.ssh/work_key

ssh-add -t 1h ~/.ssh/id_rsa # This allows adding a time-limited key
```

If you want to remove a single or multiple keys from the agent:

```sh
ssh-add -d ~/.ssh/specific_key # removes the specified key
ssh-add -D # removes ALL the keys!
```

# View SSH keys stored by the agent

If you want to view the keys that are stored within the agent:

```sh
ssh-add -l
```

# Agent-forwarding to use private keys on remote servers

If you need access to your private key on a remote server without copying the key you can use the agent forwarding feature of the ssh-agent. When enabled the remote server can temporarily use your private keys stored in the ssh-agent.

This could be useful in scenarios where for example you want to clone a git repository on a remote server but don't want to copy the SSH key to access the repo.

```sh
ssh -A user@remote-server
```

[!WARNING] Only do this with trusted servers because this gives the remote server access to your local ssh-agent

[!TIP] Because of its dangerous nature consider using `ssh -J jumpuser@jumphost targetuser@targethost` instead of agent-forwarding when possible
