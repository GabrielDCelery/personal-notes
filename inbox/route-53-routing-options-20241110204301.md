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





