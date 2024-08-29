# How to convert a Raspberry Pi into a WiFi router

## Why did I want to create my own router for my homelab

As of this writing we were living at a place where we had to use the landlord's internet (that was in an other building) which meant the physical router was in his posession. In order to be able to run my homelab I decided to set up my own router. Main considerations:

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

First had to check the version of the Raspberry Pi by looking at the board and then verifying that the model I had was the `Raspberry Pi 3 B` and not the `B+` version.

Afterwards navigated to the OpenWrt website to find the appropriate version of the operating system. [https://openwrt.org/toh/raspberry_pi_foundation/raspberry_pi](https://openwrt.org/toh/raspberry_pi_foundation/raspberry_pi). As of this writing the `Supported Current Rel` was `23.05.0`.

After downloading the necessary image used the Raspberry Pi imager ([https://www.raspberrypi.com/software/](https://www.raspberrypi.com/software/)] to write it onto the micro SD card.

In terms of settings was using the `custom image` option and used the `default settings`.

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


```sh
opkg update
```

```sh
opkg install usbutils
lsusb

Bus 001 Device 003: ID 0424:ec00
Bus 001 Device 002: ID 0424:9514
Bus 001 Device 001: ID 1d6b:0002 Linux 5.15.134 dwc_otg_hcd DWC OTG Controller
```

And once plugged in.

```sh
lsusb

Bus 001 Device 004: ID 148f:5370 Ralink 802.11 n WLAN
Bus 001 Device 003: ID 0424:ec00
Bus 001 Device 002: ID 0424:9514
Bus 001 Device 001: ID 1d6b:0002 Linux 5.15.134 dwc_otg_hcd DWC OTG Controller
```

Then went into the `/etc/config` directory to edit the `wireless`, `network` and `firewall` settings. Before editing those files created a copy of each of them using the command `cp <filename> <filename>.bak`.

#### 1. edit /etc/config/network

- under `config interface 'lan'` change the `option ipaddr '192.168.1.1'` line to something like `option ipaddr '10.71.71.1` 
- under `config interface 'lan'` section add a new line `option force_link '1'`
- we need to create a new network interface for the wireless connection (the one we want to connect to)

```sh
config interface 'wwan'
    option proto 'dhcp' # This tells the interface to pull an IP address from whichever wireless network we are connectin to
    option peerdns '0'
    option dns '1.1.1.1 8.8.8.8' # Which addresses to use for DNS resolution
```

#### 2. edit /etc/config/firewall

- under the `config zone` section where `option name` says `wan` change `option input` from `REJECT` to `ACCEPT`

#### 3. reboot

After setting the above we have to reboot using the `reboot` command.

#### 4. edit /etc/config/wireless

We have to make our Raspberry Pi be able to connect to the internet via the built-in wireless network interface, but that needs to be configured.

For that we have to edit the `config wifi-device 'radio0'` section in the `wireless` config.

- set the `option disabled` flag from `1` to `0`

After the change has been done we have to apply the change by running the following:

```sh
uci commit wireless
wifi
```

After this step you can verify if you were successful by checking the available WiFi hotspots around you, a new one called `OpenWrt` should appear around you.

#### 5. Configure the network settings using the OpenWrt GUI

Connect to the new WiFi called `OpenWrt` with your laptop and then visit in the browser the IP address that was used at the `option ipaddr` setting when editing the `/etc/config/network` file.
