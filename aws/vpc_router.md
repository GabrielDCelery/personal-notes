---
title: VPC router
author: Gabor Zeller
date: 2024-09-01T16:46:13Z
tags: ['aws']
draft: true
---

# VPC router

A VPC router is a virtual router. This is a `highly available` service across `all AZs` withing a region. At its core is job is to route traffic `between subnets within a VPC` and `IN or OUT of the VPC`.

This VPC router has a network interface in `every subnet` at the `subnet + 1` address. The routing of that traffic is handled through `route tables` which are `associated with subnets`.

This VPC router is responsible for ensuring that services within the VPC can reach the `AWS public zone` or the `public internet` via `Internet or NAT Gateway`. It also ensures that your traffic can reach `on-premises` networks using a `Virtual Private Gateway` or a `Transit Gateway`.

#### VPC route tables

Every VPC is created with a `default route table`. This is what gets automatically assigned to subnets within the VPC. You can create `custom route tables`, but the moment you associate them with a subnet the default route table gets disassociated. This is important a subnet `can only have a single route table associated with it`. As a best practice the `main route table should not be modified`. That is because if in case the subnet falls back on the default route table you don't want to have any unintended routes to be associated with the subnet.

A route table contains routes where the `most specific route` wins. There is also the `0.0.0.0/0` address which matches anything and is generally used to route traffic to the default gateway. The specificity depends on the network mask. A `/32` mask has a higher specificity than a `/16` so it gets prioritized. If `IPv6` has been enabled in the VPC then the route table can have IPv6 routes as well.




