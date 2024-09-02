---
title: How to convert a Raspberry Pi into a WiFi router
author: Gabor Zeller
date: 2024-09-01T16:46:13Z
tags: ['raspberrypi', 'router']
draft: true
---

# How to convert a Raspberry Pi into a WiFi router

## Why did I want to create my own router for my homelab

As of this writing we were living at a place where we had to use the landlord's internet (which was in an other building) which meant the physical router was in his posession. In order to be able to run my homelab I decided to set up my own router. Main considerations:

- Security
- Proxmox does not have WiFi support out of the box, so having a [own router] - [ethernet cable] - [homelab] setup seemed logical

## Equipment

- Raspberry Pi 3 Model B V1.2 
- Wireless WLAN USB adapter RT5370: (150Mbit/s dongle 802.11n/b/g)
- micro SD card
- micro SD card reader
- BENFEI USB-Ethernet hub (because my laptop did not have an ethernet port)

## Setup

Before even starting had to download a driver for the USB hub from [https://www.think-benfei.com/p_driver.html](https://www.think-benfei.com/p_driver.html). The laptop has Windows on it so checked the network card's status via `Control panel` -> `View network status and tasks`. 

Verified the version of the Raspberry Pi by looking at the board and then verifying that the model I had was the `Raspberry Pi 3 B` and not the `B+` version.

## Install the OS on the Raspberry Pi

Afterwards navigated to the OpenWrt website to find the appropriate version of the operating system. [https://openwrt.org/toh/raspberry_pi_foundation/raspberry_pi](https://openwrt.org/toh/raspberry_pi_foundation/raspberry_pi). As of this writing the `Supported Current Rel` was `23.05.0`.

After downloading the necessary image used the Raspberry Pi imager ([https://www.raspberrypi.com/software/](https://www.raspberrypi.com/software/)] to write the image onto the micro SD card.

In terms of settings was using the `custom image` option and used the `default settings`.

## Connect to the Pi via my laptop

After setting up the Pi I connected to it using the hub and an ethernet cable. The Pi was automatically handing out an IP address via DHCP which I verified via `ipconfig` using the `Command Prompt`, but had it not worked I would have edited the network card's settings via the `Network and sharing center`, select the network card, `Properties`, find `Internet Protocol Version 4` and change the settings from `Obtain IP address automatically` to `Use the following IP address`.

```s
IP Address: 192.168.1.10 (or some valid static IP)
Subnet Mask: 255.255.255.0
Default Gateway: 192.168.1.1
```

After setting up the network card and ethernet connection used `ssh` to get into the Pi (also had to remove the IP address from the known hosts).

```sh
ssh root@192.168.1.1
```

First set the password for root using the `passwd` command.

## Connecting the Raspberry Pi to the landlord's internet

Afterwards navigated to `192.168.1.1` in my browser to connect the Pi to the Wifi using its built in radio interface.

Once logged into the visual interface via the browser went to `Network` then `Wireless` settings and used the `Scan` button to search for the landlord's network (in reality at that point had an other router in the house that worked as a `client ap` so in actuality connected to that, but the principle is the same).

Clicked on `Join network` then checked the tickbox for `Replace wireless configuration`, typed in the `WFA passphrase` (password for the network), then pressed `Save`, `Save` again then `Save & Apply` on the main dashboard.

Below are the changes the interface applied to the Pi (can be previewed before applying).

```sh
# /etc/config/firewall
uci del firewall.cfg02dc81.network
uci add_list firewall.cfg02dc81.network='lan'
uci del firewall.cfg03dc81.network
uci add_list firewall.cfg03dc81.network='wan'
uci add_list firewall.cfg03dc81.network='wan6'
uci add_list firewall.cfg03dc81.network='wwan'
# /etc/config/network
uci set network.wwan=interface
uci set network.wwan.proto='dhcp'
# /etc/config/wireless
uci del wireless.radio0.disabled
uci set wireless.wifinet1=wifi-iface
uci set wireless.wifinet1.device='radio0'
uci set wireless.wifinet1.mode='sta'
uci set wireless.wifinet1.network='wwan'
uci set wireless.wifinet1.ssid='Andromeda83'
uci set wireless.wifinet1.encryption='psk2'
uci set wireless.wifinet1.key='someSuperSecretPassword'
uci set wireless.default_radio0.disabled='1'
uci set wireless.radio0.cell_density='0'
```
Once the changes have been applied simply verified by pinging google from within the Pi.

## Setting up the antenna so to be able to connect to the Pi via WiFi

First sshd into the Pi and did an update.

```sh
opkg update
```

Then downloaded the driver so the pluggen in antenna could be detected.

```sh
# Before installing usb utils
lsusb

Bus 001 Device 003: ID 0424:ec00
Bus 001 Device 002: ID 0424:9514
Bus 001 Device 001: ID 1d6b:0002 Linux 5.15.134 dwc_otg_hcd DWC OTG Controller

# After installing usb utils and plugging in the antenna
opkg install usbutils

Bus 001 Device 004: ID 148f:5370 Ralink 802.11 n WLAN
Bus 001 Device 003: ID 0424:ec00
Bus 001 Device 002: ID 0424:9514
Bus 001 Device 001: ID 1d6b:0002 Linux 5.15.134 dwc_otg_hcd DWC OTG Controller
```

Also installed the necessary drivers for enabling the antenna as an access point.

```sh
opkg install kmod-rt2800-lib kmod-rt2800-usb kmod-rt2x00-lib kmod-rt2x00-usb
```

After installing the drivers verified that the antenna will work by running the `ip a` command

```sh
5: wlan0: <BROADCAST,MULTICAST> mtu 1500 qdisc noop state DOWN qlen 1000
    link/ether 00:0c:e7:41:61:06 brd ff:ff:ff:ff:ff:ff
```

Then turned on the antenna by running 

```sh
ifconfig wlan0 up
```

As the next step I edited the `/etc/config/wireless` file to enable the antenna so I could connect to it using WiFi.

```sh
config wifi-device 'radio1'
        option type 'mac80211'
        option path 'platform/soc/3f980000.usb/usb1/1-1/1-1.2/1-1.2:1.0'
        option channel '1'
        option band '2g'
        option htmode 'HT20'
        option disabled '1'

config wifi-iface 'default_radio1'
        option device 'radio1'
        option network 'lan'
        option mode 'ap'
        option ssid 'OpenWrt'
        option encryption 'none'
```

In the above configuration:

- set `disabled` to `0`
- renamed the `ssid` to a more user friendly name `GaZeRaspRouter`
- added a new row under `default_radio1` that said `option encryption 'psk2'`
- added an other row `option key <password for wireless>`

Ran again `uci commit wireless` then the command `wifi`, `rebooted` the router then verified if I could connect to it using my phone.
