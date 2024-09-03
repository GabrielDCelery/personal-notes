---
title: DHCP in VPC
author: Gabor Zeller
date: 2024-09-01T16:46:13Z
tags: ['aws']
draft: true
---

# DHCP in VPC

DHCP allows for the automatic configuration of network resources. Each device that appears on a network only starts with a `MAC address` and starts an `L2 broadcast` to find a `DHCP server`. Once they find each other they start communicating using `layer 2` to acquire `layer 3` capabilities. Once they do the new device acquires an `IP address`, a `subnet mask` and a `default gateway`.

DHCP is also responsible for determining which `DNS servers` to use. Either you can choose the `AmazonProvidedDNS` or select a `Custom DNS Domain`.

With DHCP option sets there is also the nice feature that you can assign your own custom domain names to your newly spun up EC2 instances (as opposed to the default ones) if you add the `custom domain` name and `custom DNS servers` to the DHCP option set.

DHCP option sets are `immutable`, once created they can not be changed only `swapped out`. The option set can be attached to zero or more VPCs. On the other hand a VPC can only have maximum one option set associated with it.

One important thing to remember is that `the association of an option set is immediate`, but `the change to take effect takes time`, because they require a DHCP renew.

There is also an other important consideration with option sets within a VPC. Under normal circumstances a DHCP is responsible for allocating `IP addresses`, the `subnet mask` and the `default gateway` to newly detected devices. In the AWS VPC the DHCP also does that, but you `can not configure these values`. The IP address is allocated automatically, the subnet mask is the same as the subnet in which the new instance is spun up, and the default gateway is always the `VPC + 1 address`.

What you can configure in an option set is the `DNS resolver` which by default is the `VPC + 2 address`. You can also configure your own `NTP server` if time synchronisation is important. By default AWS also manages assigning `DNS hostnames` (both private and public), to use your own custom domains you need your own `custom DNS server`.



