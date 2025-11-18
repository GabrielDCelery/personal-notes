---
title: "Figure out why removing NACL ingress allow all broke our system"
date: 2025-09-15
tags: ["aws", "nacl"]
---

# Premise

Besides our secrurity groups we have several rules applied at the NACL level.

The issue is that they were too permissive, as one of the last rules we had an allow `0.0.0.0/0` that was allowing all incoming traffic. As a pre-requisite of a pen testing that rule was removed and suddenly several of our services started timing out.

## What did I try?

Among the NACL rules we had a port `80` and `443` inbound rule, some other (for example 5432 for postgres from specific IP addresses) and a `0.0.0.0/0` inbound rule for all ports. For outbound rules it was `0.0.0.0/0` allow.

First I suspected that maybe port `53` was missing for DNS resolution, but that did not solve the issue, so did some research and started playing around with the ports.

**1-1024**: Well-known ports (also called system ports)

- Reserved for common services and protocols
- Requires root/administrator privileges to bind to these ports
- Examples: HTTP (80), HTTPS (443), SSH (22), FTP (21), SMTP (25)

**1024-49151**: Registered ports (also called user ports)

- Can be registered with IANA for specific services
- Don't require special privileges to use
- Examples: MySQL (3306), PostgreSQL (5432), RDP (3389)

**49152-65535**: Dynamic/Private ports (also called ephemeral ports)

- Used for temporary connections and client-side ports
- Automatically assigned by the OS for outgoing connections
- Not registered with IANA

I was thinking that maybe some of the registered ports between `1024-49151` were used somewhere in the request chain of services that we overlooked and we did not have an extra rule for, but the investigation did not yield results.

Then I was thinking that maybe not one of the outbound rules is missing, but since we are dealing with stateless firewalls for each request port we need to have a corresponding port to accept responses for.

## The answer

When it comes to NACLs unlike security groups they are stateless which means there has to be a rule for accepting both inbound AND outbound traffic for them to work.

Let's assume we got a lambda making an https call. It will create an ephemeral port to make a request (e.g. 64214). Our NACL is configures to allow all OUTBOUND traffic which means the lambda can send the request to the API endpoint via https, but when it receives the response the traffic will try to come back on the ephemeral port the lambda created. Which is what was disabled when we removed the acceptance of all inbound calls so the firewall does not allow the response traffic to get through and the call times out.

Checked the docs and for ephemeral calls `AWS lambda functions use ports 1024-65535`. [Custom network ACLs for your VPC](https://docs.aws.amazon.com/vpc/latest/userguide/custom-network-acl.html)
