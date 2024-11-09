---
title: Cloudformation nested stacks vs cross-stack references
author: GaborZeller
date: 2024-11-09T15-55-08Z
tags:
	- aws
	- cloudformation
draft: true
---

# Cloudformation nested stacks vs cross-stack references

When using Cloudformation the hard limit of a template is `500 resources`. Which means when dealing with large applications it is useful to have a strategy of splitting up those resources.

## Nested stack

### When to use nested stacks

Nested stacks allow for `template reuse`, which means the same template can be deployed in different stacks but you don't have to redeclared.

The other use is deploying stacks together that are `lifecycle dependent`, meaning a resource you deploy in one stack is used by an other stack and they have to be created/updated/deleted together.

### How to use a nested stack

In Cloudformation there is a resource called `Stack` that is treated like any other resource but in actuality is a different stack bundling together multiple resources.

```yaml
ChildStack:
	Type: AWS:Cloudformation:Stack
	Properties:
		TemplateURL: https://somelocation.com/template.yaml
		Parameters:
			Param1: Ref! SomeParam1
			Param2: Ref! SomeParam2
			Param3: Ref! SomeParam3
```

### How to pass around parameters across nested stacks

Nested stacks propagate their output up to the parent that can be references as `ChildStack.Outputs.XXXXX`.

[!WARNING] You can not reference a child stack's resource directly, you have to use an output

[!TIP] The ouptut of a child stack can be passed down via the parent to an other child stack

## Cross-Stack references

### When to use cross-stack references

This method is useful when you have a physical resource deployed via a stack that you want to reference in an other cloudformation stack.

### How to use cross-stack references

In order to use this method you have to `export the resource` that you want to share. That exported name has to be `unique` and it is `accessible withing a region`.

To use it you use the `Fn::ImportValue` function (as opposed to Ref).

```yaml
Outputs:
	SHAREDVPCID:
		Description: Shared VPC ID
		Value: !Ref VPC
		Export:
			Name: SHAREDVPCID
```

The above can be used by an other stack by using `!ImportValue SHAREDVPCID`



