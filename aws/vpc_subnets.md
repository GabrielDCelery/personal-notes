---
title: VPC Subnets
author: Gabor Zeller
date: 2024-09-01T16:46:13Z
tags: 
    - aws
    - vpc
draft: true
---

# VPC Subnets

VPC subnets are sub-sections of a VPC. A subnet always lives within an AZ and is resilient withing that data center. One important thing to remember is that a subnet can never be in multiple AZs. Subnets withing a VPC can communicate with each other.

A subnet always gets allocated with an IPv4 CIDR range that is a subset of the VPCs CIDR range and can not overlap with an other subnet's CIDR range. Subnets can also get optionally allocated with an `IPv6 address` but only if IPv6 was enabled on the VPC. The CIDR range uses a `/64` mask which means we can have 256 subnets in a `/56` VPC CIDR range.

## Reserved IP addresses

Every subnet has `5 reserved` IP addresses that can not be used by any of the instances withing that subnet. In an hypothetic subnet `10.16.16.0/20` the following addresses are unusable:

- `10.16.16.0` - network address - this is not an AWS specific thing, the network address is reserved on every network
- `10.16.16.1` - vpc router - every subnet has a `network interface` and this logical device is resopoinsible routing traffic in and out of the subnet
- `10.16.16.2` - dns resolver - this is a bit tricky because technically AWS uses the VPC+2 address for DNS resolution, but they reserve the address in every subnet regardless
- `10.16.16.3` - reserved - not used for anything, reserved for the future in case it is needed
- `10.16.31.255` - broadcast - even though broadcasting is not supported within a VPC, AWS does reserve this regardless

## DHCP Option set

Every VPC has a configuration set attached to them, called the `DHCP option set`. This configuration is what gets carried over to subnets. It controls things like `DNS servers`, `NTP servers`, `NetBIOS servers` and some other things. You can create option sets to replace existing ones, but `you can not edit them` once they have been created.

## Other subnet settings

There are two other important settings for subnets. One is `auto assigning IPv4 addresses`. If that is turned on than any instance in that subnet alongside their privately allocated IP address will get a public IP address allocated. The other importan option is to enable `auto assigning IPv6 addresses`. For the latter IPv6 has to be enabled both at the VPC and the subnet level.


