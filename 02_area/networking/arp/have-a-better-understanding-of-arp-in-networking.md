---
title: "Have a better understanding of ARP in networking"
date: 2025-09-18
tags: ["arp", "networking"]
---

# What is ARP

ARP is the address resolution protocol and it's purpose to translate known IP addresses to unknown MAC addresses.

# How does it work?

The process looks like this:

1. Computer A knows it's own IP address and the IP address that it wants to send the data
2. Computer A knows if the target IP is on the same network then it needs the MAC address of a computer on the same network, if the target is on a different network then it needs the MAC address of the local network's gateway
3. Computer A sends out a broadcast to `every` machine on the network with it's own MAC address
4. Every machine on that network receives the request and after inspection they determine whether it was intended for them or not
5. Computer B recognises that it is the intended target so it sends back it's MAC address to computer A

> [!NOTE]
> When the requestor sends out the broadcast the message includes it's own MAC address, this is how the target knows where to send back the response

# Request ethernet header breakdown

| Field              | Value             | Description                                                                                                      |
| ------------------ | ----------------- | ---------------------------------------------------------------------------------------------------------------- |
| Source MAC address | 00:15:5d:c7:22:09 | This is the sender's address                                                                                     |
| Target MAC address | ff:ff:ff:ff:ff:ff | This is the target address, since we are talking about a broadcast it is sent to `ff:ff:ff:ff:ff:ff`             |
| EtherType          | 0x0806            | This is a reserved value/EtherType for address resolution (the hex value translates to `2054`)                   |
| Padding            | 0s                | Since the minimum size of at ethernet frame is 64 bytes, we need 18 bytes of padding to make a valid ARP request |

# Request payload breakdown

Since the ARP request is sent via an ethernet frame it got a few important items that are worth knowing about:

| Field              | Value             | Description                                                                                                                |
| ------------------ | ----------------- | -------------------------------------------------------------------------------------------------------------------------- |
| Sender MAC address | 00:15:5d:c7:22:09 | This is the sender's address                                                                                               |
| Sender IP address  | 10.12.6.8         | The source IP address                                                                                                      |
| Target MAC address | 00:00:00:00:00:00 | This is the target address, since we do not know the MAC address of the target yet it is all 0s                            |
| Target IP address  | 10.12.6.201       | The target IP address                                                                                                      |
| Hardware Type      | Ethernet (1)      | Type of address we are mapping from (we are mapping from a MAC address to an IP address)                                   |
| Protocol Type      | IPV4 (0x0800)     | Type of address we are mapping to (we are mapping from a MAC address to an IP address, the hex value translates to `2048`) |
| Hardware Size      | 6                 | Size of the Hardware Type, since it is a MAC address it is `6 bytes` (`48 bits`)                                           |
| Protocol Size      | 4                 | Size of the Protocol Type, since it is an IP address it is `4 bytes` (`32 bits`)                                           |
| Opcode request     | request (1)       | Type of the request, either `1` to indicate request and `2` for response                                                   |

# ARP request handling

When the request resolves the corresponding MAC address it is stored in the kernel's memory in the ARP Table/Cache. Laptops, phones and devices that move around a lot are usually cached for 60 seconds, network equipments are stored for 2-4 hours.

> [!NOTE]
> Devices can not tell just by the request if the other device is an ordinary host or a network infrastructure device, so in order to keep up a router's cache is updated when it receives an ARP request

# Footnotes

[^1] [Practical networking - traditional ARP](https://www.practicalnetworking.net/series/arp/traditional-arp/)
[^2] [IETF protocol and documentation for IEEE 802 parameters](https://datatracker.ietf.org/doc/html/rfc7042)
