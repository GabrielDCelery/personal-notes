---
title: Custom VPCs
author: Gabor Zeller
date: 2024-09-01T16:46:13Z
tags: ['aws']
draft: true
---

# Custom VPCs

VPC is a regional service that works accross multiple AZs within that region. They create an isolated network within an account an nothing can go in or out of that network boundary without explicit permissions. It provides an isolated blast radius in case something goes wrong within the VPC.

When creating a VPC you have the option to create a `default` or `dedicated tenancy`. The default option means you are going to be sharing hardware with other users, and the network boundaries are logical boundaries. Dedicated tenancy means you get your own hardware that serves as your network.

#### IP addresses within a VPC

Each VPCs get created with one mandatoy IPv4 CIDR block that has the restrictions of being size `/28` minimum and `/16` maximum. After creating a VPC you can assign optional secondary IPv4 blocks.

A VPC can also be configured to use `IPv6` by assigning an IPv6 `/56` CIDR block. One important difference between creating an IPv4 layout as opposed to an IPv6 one is that if the feature is enabled either AWS `automatically assigns` the IP address range to the VPC, or you can select to use an IPv6 range `that you own`. Also there is no difference between `public` and `private` addresses when using IPv6, all the addresses are `public`.

#### DNS resolution within the VPC

Each VPC has DNS resolution provided by `Route53`, at the `VPC base IP + 2 address`. There are also two settings that are critical when setting up a VPC. One is `enableDnsHostNames` which will automatically assign public DNS hostnames to instances that got a public IP when enabled. The other one is `enableDnsSupport` which determines whether there is DNS support within the VPC or not. If enabled the instances withing the VPC can use the base IP+2 address for DNS resolution.


