---
title: Cloudformation conditions
author: GaborZeller
date: 2024-11-09T15-01-05Z
tags:
	- aws
	- cloudformation
draft: true
---

# Cloudformation conditions

Cloudformation conditions can be used to evaluate if certain conditions are matched (or not matched) at deployment and make the template behave differently.

Example:

1. Have a parameter describing the environment

```yaml
Parameters:
	EnvType:
		Defaults: 'dev'
		Type: String
		AllowedValues:
			- 'dev'
			- 'prod'
```

2. Use the parameter to create a condition

```yaml
Conditions:
	IsProd: !Equals
		- !Ref EnvType
		- 'prod'
```

3. Use the condition to deploy certain resource

```yaml
Resources:
	MyEIP:
		Type: AWS:EC2:EIP
		Condition: IsProd
```
