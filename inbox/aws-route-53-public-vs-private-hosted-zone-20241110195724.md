---
title: AWS Route 53 public vs private hosted zone
author: GaborZeller
date: 2024-11-10T19-57-24Z
tags:
	- aws
	- route53
draft: true
---

# AWS Route 53 public vs private hosted zone

## Public hosted zone

It is a `globally resilient` service that is accessible both via the `public internet` and both via from `VPCs`.

It acts as an `authoratative zone` for your domain. It gives you a `zone file` and `4 managed namservers` that can handle both a domain that you registered via AWS or externally.

If `dns resolution` is enabled for a `VPC` then applications from within that VPC can reach the `Route53 resolver` via the `VPC+2 address`.

When someone is using it from the `public internet` then the client walks the DNS tree until it reaches the Route 53 resolver.

## Private hosted zone

The private version of a hosted zone unlike the public one `has to be associated with a region and a VPC`. You can have `multiple VPCs` associated with that hosted zone. All applications withing that VPC have access to the private resolver.

[!WARNING] a private hosted zone can not be queried from outside the public internet

## Split view hosted zone

It is possible to create a `public` and a `private` hosted zone, both with `the same name` and where the public zone shares records with the private one.

This way it is possible to make a subset of records public, while having records that we do not want to share within the private hosted zone.
