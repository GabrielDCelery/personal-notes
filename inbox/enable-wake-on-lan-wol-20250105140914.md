---
title: Enable Wake on LAN (WOL)
author: GaborZeller
date: 2025-01-05T14-09-14Z
tags:
draft: true
---

# Enable Wake on LAN (WOL)

## Enable WOL in the BIOS

As the first step the feature has to be enabled under the BIOS, the settings should be under somewhere `Advanced Power Management`, either explicitly named as `Wake on LAN` or something like `Power On By PCIE Devices`.

## Enable WOL on the network interface

### Enable WOL on a Linux based system

1. First check the network interfaces and whether they are WOL capable or not.

Use the `ip` command to get a lits of 

```sh
ip -c a
```
Use the `ethtool` command to check the details of the network interface

```sh
ethtool enp7s0
```

Look for the following section:

```sh
Supports Wake-on: pumbg # pumbg means it is WOL capable
Wake-on: d # d means disabled, g means enabled
```

2. Configure the network to have WOL enabled on the network card

Use an editor with root priviliges.

```sh
vi /etc/network/interfaces
```

Add the following line at the end of the interfaces file (before the `source /etc/network/interfaces.d/*` line)

```sh
post-up /usr/sbin/ethtool -s enp7s0 wol g # enp7s0 is the name of the interface in this example
```

## Get a tool to use WOL

### Windows

There is an app called `Wake on LAN` in the `Microsoft store`

### Linux

You can install a tool like `wakeonlan` using `homebrew` and then use the following command to wake up the machine:

```sh
wakeonlan b3:3a:16:29:a3:c5
```

You can also inspect the packets on the Linux machine using the following command:

```sh
 sudo tcpdump -nxXei any ether proto 0x0842 or udp port 9 2>/dev/null
```
