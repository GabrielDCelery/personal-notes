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

## Setup

First had to check the version of the Raspberry Pi by looking at the board and then verifying that the model I had was the `Raspberry Pi 3 B` and not the `B+` version.

Afterwards navigated to the OpenWrt website to find the appropriate version of the operating system. [https://openwrt.org/toh/raspberry_pi_foundation/raspberry_pi](https://openwrt.org/toh/raspberry_pi_foundation/raspberry_pi). As of this writing the `Supported Current Rel` was `23.05.0`.

After downloading the necessary image used the Raspberry Pi imager ([https://www.raspberrypi.com/software/](https://www.raspberrypi.com/software/)] to write it onto the micro SD card.

In terms of settings was using the `custom image` option and used the `default settings`.

After setting up the Pi, connected it to an external monitor and powered it on. By default the logged in user is `root` with no passwords.

First set the password for root using the `passwd` command.

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
