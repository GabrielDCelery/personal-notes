---
title: Copy SSH key to linux machine and disable password
tags:
  - linux
  - ssh
---

# The problem

After installing Ubuntu server I wanted to add an ssh key and disable password login.

# The solution

1. Generate a key

```sh
ssh-keygen -t ed25519 -b 4096 -f ~/.ssh/homelab_admin
```

2. Copy over to the other machine

```sh
ssh-copy-id -i ~/.ssh/homelab_admin.pub username@ipaddress

# Alternatively if the above does not work
# Check contents of public key
cat ~/.ssh/id_ed25519.pub
# SSH into server
ssh username@ipaddress
# On the server, create the .ssh directory and add key to authorized_keys
mkdir -p ~/.ssh
chmod 700 ~/.ssh
echo "your_public_key_content_here" >> ~/.ssh/authorized_keys
chmod 600 ~/.ssh/authorized_keys
exit
```

2b.

3. Test the connection

```sh
ssh -i ~/.ssh/homelab_admin username@password
```

4. Edit SSH configuration file

```sh
sudo vi /etc/ssh/sshd_config
```

```txt
PasswordAuthentication no
PubkeyAuthentication yes
ChallengeResponseAuthentication no
UsePAM no  # Optional, but more secure
PermitRootLogin no  # Highly recommended
```

5. Restart SSH service

```sh
sudo systemctl restart sshd
```
