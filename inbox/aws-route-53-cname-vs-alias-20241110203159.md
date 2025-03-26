---
title: AWS Route 53 CNAME vs ALIAS
author: GaborZeller
date: 2024-11-10T20-31-59Z
tags:
	- aws
	- route53
draft: true
---

# AWS Route 53 CNAME vs ALIAS

## What is the problem that an ALIAS record tries to solve

Among the normal record types either we have the `A` record that can point to an `IP address` or a `CNAME` that can point to an other name.

The problem that when AWS creates a resource (for example load balancer) `they give a DNS name as opposed to an IP address`.

Which means we can not point the `naked apex` domain to an AWS service with the box standard DNS record types.

## How does an ALIAS record work

It works pretty much the same way as a `CNAME` record, it can be used for normal records, but at the same time can also be used for naked/apex records.

[!TIP] in general AWS recommends using ALIAS record for pointing to AWS services, they also incentivise by not charging for it
