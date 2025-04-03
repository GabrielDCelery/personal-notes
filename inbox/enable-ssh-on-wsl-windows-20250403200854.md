---
title: Enable SSH on WSL Windows
author: GaborZeller
date: 2025-04-03T20-08-54Z
tags:
draft: true
---

# Enable SSH on Windows

## Make sure your computer is discoverable

[!WARNING]: Make sure your network is set as a "Private" and not "Public" network

- Open Settings -> Network & Internet
- Click on your active network connection
- Change network profile to Private if it's set to Public

## Install and enable the OpenSSH service

Go to `Settings -> Optional features` and then search and add `OpenSSH Server` and also add `OpenSSH Client` (useful).

Once installed using Powershell in `admin` mode enable and then verify if the service is running. You also have to enable the firewall rules to allow SSH connections coming in.

```sh
# start the service
Start-Service sshd
# verify if the service is running
Get-Service sshd
# Make sure the service starts on reboot
Set-Service -Name sshd -StartupType 'Automatic'
# Verify if the connection through the firewall has been enabled
Get-NetFirewallRule -Name *ssh*
# Look for
## Direction		Inbound
## Action			Allow
```

## Verify if the ssh server is working

If you installed the `OpenSSH Client` use the command `ssh username@localhost` on the same machine where the server is running (Using Powershell os CMD).

You should also be able to SSH into the machine from an other machine on the same network using the `ssh username@ipv4addr` or `ssh username@computername`.

# Enable SSH on WSL Windows

If you enabled SSHing via `Windows OpenSSH` once logged in you can run `wsl.exe -d <wsldistroinstalled>` to get into WSL but you might want to have the ability to directly SSH into WSL from your machine.

## Install Open SSH server and start it

```sh
sudo apt update
sudo apt install openssh-server
sudo systemctl enable ssh
```

## Configure the ssh server

Since the Windows machine can have the OpenSSH server installed and enabled its better practice to enable WSL SSH on a different port.

For that we need to add a new file at `/etc/ssh/sssh_config.d/port.config` and add the following line (this will be read by `/etc/ssh/ssh_config`):

```sh
# /etc/ssh/sssh_config.d/port.config
Port 2200
```

Once done create a file at `/etc/systemd/system/ssh.socket.d/listen.conf` and add the following lines.

```sh
sudo mkdir -p /etc/systemd/system/ssh.socket.d/
sudo touch /etc/systemd/system/ssh.socket.d/override.conf
sudo vim /etc/systemd/system/ssh.socket.d/override.conf
```

```sh
[Socket]
ListenStream= # This is important for unsetting it first
ListenStream=2200
FreeBind=yes
Accept=no
```

Go to the `/etc/wsl.conf` file and make sure these lines are present. If not add them.

```sh
[boot]
systemd=true
```

```sh
sudo systemctl restart ssh.socket ssh.service ssh
sudo systemctl daemon-reload
sudo systemct status ssh # To verify if we are indeed listening on 2200
```

## Open up the port on the Windows machine

Run the following in PowerShell:

```sh
New-NetFirewallRule -Name "WSL SSH" -DisplayName "WSL SSH" -Direction Inbound -Protocol TCP -LocalPort 2200 -Action Allow
```
