---
title: Cloudformation template and pseudo parameters
author: GaborZeller
date: 2024-11-09T10-38-34Z
tags:
	- aws
	- cloudformation
draft: true
---

# Cloudformation template and pseudo parameters

## What are template parameters

Template parameters are parameters that the user/application deploying the stack can specify to customize the template at `deployment time`.

Some useful template parameters are:

- `Type` - the type of the template parameter
- `Defaults` - specify a default value if none was provided
- `AllowedValues` - list of allowed values
- `Min` and `Max` - can specify a range to limit what can be deployed
- `AllowedPatterns` - you can use regual expressions to limit what the value can be
- `NoEcho` - this will mask the parameter with `*****` to prevent it being shown on the console, but `DOES NOT MASK IT IN THE METADATA OR OUTPUTS`

```yaml
Parameters:
	InstanceType:
		Type: String
		Default: 't3.micro'
		AllowedValues:
			- 't3.micro'
			- 't3.medium'
```

## What are pseudo parameters

These are parameters that are provided by AWS. If a template references any of these AWS replaces them at `deployment time`.

Examples:

- `AWS:Region` - the region at which the template is being deployed
- `AWS:StackName` and `AWS:StackId`- the name and the unique ID of the stack
- `AWS:AccountId` - populated by the ID of the account where the stack is being created
