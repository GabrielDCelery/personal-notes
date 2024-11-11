---
title: Route 53 routing options
author: GaborZeller
date: 2024-11-10T20-43-01Z
tags:
	- aws
	- route53
draft: true
---

# Route 53 routing options

## Simple routing

Simple routing supports one record per name, each record can have multiple values and `all of the targets` are returned in a `random order` to the client.

```
www		A		1.2.3.4
				0.1.2.3
				3.4.5.6	
```
[!WARNING] does not support health checks

## Failover routing

With failover routing you specify two resources. The first `primary` resource uses health checks and if they pass the first resource is being served. If the health checks fail the secondary resource gets served.

When configuring this option `first you need to create a healath check` then configure a DNS record with that health check and setting its type to `primary` and then create an other record without the health check thay is your `secondary` record.

## Multi-value routing

It is like a mix of `simple routing` and `failover routing`. You create multiple records, `each with its own heatlth checks`. Any records that fail the health check won't be returned to the client. The rest is returned randomly.

[!WARNING] It is not a substitute for a load balancer but offers an active-active approach for serving applications from a DNS perspective

## Weighted routing

Weighted routing offers to have multiple `A records` pointing to `different IP addresses` and each of them `has a different weight` assigned to them as percentages.

It can be combined with health checks where unhelthy records are being skipped, but they do not adjust the chance of selection set in the percentages.

## Latency routing

This type of routing allows you to set up `multiple records with the same names`, but each name gets an `associated region`. AWS keeps track of the clients trying to connect and their locations and picks a record that is the closest to them (technically the one that has the lowest estimated latency to them). This way people can be directed to infratstructures that are closer to them.

## Geolocation routing

Geolocation based routing takes into the client's location, but it is NOT about serving the closest resource but to limit what people can have access to. You can `tag each record` with either `state`, `country`, `continent` flags, or have a `default` flag and based on the client's location AWS tries to find the one that is the `most applicable`.

[!TIP] If no records are found nothing is returned which is how you can use this for restricting access based on location

## Geoproximiy routing

Geoproximity is like latency based routing, but it uses the `actual distance` as opposed to the `estimated latency`. Records can be tagged with a `region` or `lat-long` locations. It also allows allocating a `bias` to our locations, effectively making them `virtually bigger or smaller` to influence the distance measurement from the client.

