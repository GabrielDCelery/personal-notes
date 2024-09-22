---
title: Advanced EC2 networking
author: GaborZeller
date: 2024-09-22T18-49-09Z
tags:
draft: true
---

# Advanced EC2 networking

#### Attaching further elastic network interfaces to an EC2 instance

When creating an EC2 instance it gets created with a `primary network interface`. That interface does not get removed nor can be reassigned during the lifecycle of the instance.

Additional `ENIs` can be created and attached in different subnets, but those subnets have to be in the `same availability zone`. These secondary network interfaces can be unattached or re-attached to the running EC2 instance.

#### Securtiy groups and NACLs on ENI interfaces

`Security groups` are attached to the `ENI` interfaces, which means if an `EC2` instacnce has multiple of those than we can control the granularity at which users can connect to an EC2 instance. We can have a an ENI that is more restrictive for public users and one that is more permissive for example for `system admindistrators`.

Since ENIs can exist in different `subnets` then we can further protect our network interfaces by attaching `NACLs` to the subnets where these ENIs reside.

#### IPv4 addressing with ENIs

In terms of `IPv4` addresses the primary `ENI` gets a `private IPv4 address` allocated and through the additional ENIs further private addresses can be allocated to the EC2 instance. The `EC2` instance operating system has `full visibility` of these interfaces.

If configured and in a public subnet the `primary ENI` gets a `public IPv4` address allocated but that is `not visible to the EC2 instance's operating system` since that mapping happens via the `internet gateway`.

`Elastic IP adresses` can be associated both with the an EC2 instance and an ENI. If done through the former then the IP address is associated to the primary ENI.

#### Source and destination checks

By default each ENI is configured to perform a `SRC/DST` check on the traffic going through them. Unless enabled ENIs drop packets where they are not the target for that traffic or where the traffic is not originated from them.

#### Use cases for multiple ENIs

- management or isolated networks (have different firewalls set up on the same instance to allow different people have broader or limited access to an instance)
- software licensing (some legacu licenses are tied to MAC addresses so with ENIs you could set up one and move it between EC2 instances)
- security or network appliances (different traffic through different interfaces mean you could control and monitor traffic differently)
- multi-homed instances (you can have one ENI used for the web application traffic and an other one for your database traffic)
- low budget high availability solution (rather than a load balancer you can have an ENI for managing your traffic and if your EC2 instance fails you can re-attach the ENI to a different EC2 instance)

