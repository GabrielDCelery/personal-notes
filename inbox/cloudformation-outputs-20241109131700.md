---
title: Cloudformation outputs
author: GaborZeller
date: 2024-11-09T13-17-00Z
tags:
	- aws
	- cloudformation
draft: true
---

# Cloudformation outputs

## What to use Cloudformation outputs for

Cloudformation outputs are `values` that can be declared with a `description`. They are useful because

- Outputted to the console when using the `CLI`
- Accessible to `parent stacks`
- Can be references in `other stacks`

## How to use Cloudformation outputs

```yaml
Outputs:
	WordpressURL:
		Description: "Instance Web URL"
		Value: !Join [ '', [ 'https://', !GetAtt Instance.DNSName ] ]
```



