---
title: Cloudformation mappings
author: GaborZeller
date: 2024-11-09T13-12-32Z
tags:
	- aws
	- cloudformation
draft: true
---

# Cloudformation mappings

Cloudformation mappings can be used to look up values in a pre-defined mapping object.

```yaml
Mappings:
	RegionMap:
		us-east-1:
			HVM64: 'ami-something'
			HVMG2: 'ami-something'
		eu-west-2:
			HVM64: 'ami-something'
			HVMG2: 'ami-something'
```

The above can be referenced using `!FindInMap ["RegionMap", !Ref 'AWS:Region', "HVM64"]`
