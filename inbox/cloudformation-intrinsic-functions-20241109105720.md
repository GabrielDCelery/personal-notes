---
title: Cloudformation intrinsic functions
author: GaborZeller
date: 2024-11-09T10-57-20Z
tags: 
	- aws
	- cloudformation
draft: true
---

# Cloudformation intrinsic functions

## What are Cloudformation intrinsic functions used for

Intrinsic functions provide a way to use dynamic values in our cloudformation templates as opposed to just static values. Some of the use cases are:

- `Dynamic values` - using values that aren't known at creation time but are known at runtime (deployment)
- `Transformations` - if we want to dynamically create values (for example name of lambda) at runtime
- `Conditional logic` - if we want to do different things based on certain other conditions
- `Map lookups` - if we want to look up values
- `List operations` - Finding elements or selecting elements from a list

## Most common operations

### Ref and Fn::GetAtt

In Cloudformation every `logical resource` and `parameter` has a `main or primary value` that can be referenced using `Ref`. Example of referencing an other resource in a template.

Besides that resources can have other `secondary values` called `attributes` that can be retrieved using the `Fn::GetAtt` function.

An example would be an EC2 instance where the `physical ID` like `i-1234567890abcdef0` could be retrieved using `Ref` and things like `PublicIp` or `PublicDNSName` could be retrieved using `Fn::GetAtt`.

Example for using `Ref`:

```yaml
Resources:
	Instance:
		Type: 'AWS:EC2:Instance'
		Properties:
			ImageId: !Ref LatestAmiId
			InstanceType: 't3.micro'	
			Keyname: 'A4L'
```

Example for using `Fn:GetAtt`

```yaml
Resources:
  MyInstance:
    Type: "AWS::EC2::Instance"
Outputs:
  ServerDNS:
    Value: !GetAtt
      - MyInstance
      - PublicDnsName
  WebsiteURL:
    Description: Website URL for EC2 Instance
    Value: !Sub 'http://${MyInstance.PublicDnsName}'
```

### Fn::GetAZs

These can be used to get a list of availabiliy zones. It is important to note that under the hood you only get the AZs that have a `subnet in the default VPC`.

### Fn::Select, Fn::Join, Fn::Split

These can be used to work with lists.

```
AvailabilityZone: !Select [0, !GetAZs, '']

!Split ['|', 'adam|bob|david']

!Join [',', ['adam', 'bob', 'david']]
!Join ['', ['http', !GetAtt Instance.DNSName]]
```

### Fn::Base64 and Fn::Sub

When configuring an `EC2 instance` with `user data` that is expected to be provided as base64.

The `Fn::Sub` function allows to replace values at runtime in a string that are formatted as for example `${Instance.InstanceID}`. The format has to be either `${Parameter}`, `${LogicalResource}` or `${LogicalResource.AttributeName}`.

[!Warning] Substitute can not be used for self referencing.

### Fn::Cidr

Returns an array of `CIDR address blocks`. Useful when you want to allocate subnets in a more automated way. It takes in the `IP block`, the `number of subnets you want to generate` and the `range of the subnet`. 

```yaml
VPC:
	Type: AWS::EC2::VPC
	Properties:
		CidrBlock: '10.16.0.0/16'
Subnet1:
	Type: AWS::EC2::Subnet
	Properties:
		VpcId: !Ref VPC
		CidrBlock: !Select ['0', !Cidr [ !GetAtt VPC.CidrBlock, 16, 12 ] ]
Subnet2:
	Type: AWS::EC2::Subnet
	Properties:
		VpcId: !Ref VPC
		CidrBlock: !Select ['1', !Cidr [ !GetAtt VPC.CidrBlock, 16, 12 ] ]

```



