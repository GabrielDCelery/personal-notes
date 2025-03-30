---
title: Install WSL2 on Windows
author: GaborZeller
date: 2025-03-29T12-21-59Z
tags:
draft: true
---

# Install WSL2 on Windows

- Enable Virtualization in BIOS (usually under advanced - CPU)
- Go to "Turn Windows features on or off" and enable "Virtual Machine Platform" and "Windows Subsystem for Linux"
- Install WSL using `wsl --install`
- Check if WSL has been isntalled using `wsl --status`
- Go to the Microsoft Store and search for Ubuntu

- List availbale distributions using `wsl --list --online`
- List installed distributions using `wsl --list --verbose`

- Once installed ubuntu check the version running `lsb_release -a`
- Run an update `sudo apu update`

- Switch default WSL to ubuntu `wsl --set-default ubuntu`

- Add user to docker group `sudo usermod -a -G docker <username>`

- `wsl.exe -d Ubuntu`
