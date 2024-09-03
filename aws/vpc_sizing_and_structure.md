---
title: VPC sizing and structure
author: Gabor Zeller
date: 2024-09-01T16:46:13Z
tags: ['aws']
draft: true
---

# VPC sizing and structure

Even though AWS provides a virtual network, many of the same considerations have to go into designing an `IP plan` and `VPC sructure` as if it was an on premise network. This is one of the most critical works a solutions architect can do because it is very difficult to get it right, and even more difficult to change later if you don't get it right.

#### Main considerations are

- size of VPC (how many machines do we need to support)
- are there any networks we can not use? VPCs, On-premises, partners & vendors network can overlap with ours which can make communication difficult
- what is going to be the structure? How many subnets, or tiers do we need? How many AZs across should we span our network

#### Example scenario

Corporation with `3 offices` in London, Seattle and New York. The IP range of the London office is `192.168.15.0/24`, New York is `192.168.20.0/24`, Seattle is `192.168.25.0/24`. We also have `field workers` who need to communciate and connect to the business. And the business also has `3 other existing IP ranges`. An on-premise network in Austraila that is `192.168.10.0/24`, an AWS pilot network at `10.0.0.0/16` and an Azure pilot network at `172.31.0.0/16`.

Also there is an extra caveat. A previous architect tried to set up a proof of concept on `Google Cloud`, but can not confirm which networks are being used there, he could only tell us that the default range is `10.128.0.0/9`.

The new VPC design can not use any of these existing networks and can not overlap with any of them.

#### Initial considerations

The Azure pilot network `172.31.0.0/16` uses the same range as the `AWS default VPC`. Which means we have to try to avoid using the default VPC for anything.

Google's `/9` means the range `10.128.0.0/9` covers `10.128.0.0 -> 10.255.255.255`.

A VPC at minimum has to be a `/28` and maximum a `/16` network.

Generally `10.x.y.z` is a good choice on AWS. It is also good practice to avoid using `10.0.y.z` because that is the default and `10.1.y.z-10.10.y.z` because most people tend to pick these to avoid the default.

Since IP addresses use base 2, then `10.16.y.z` is a good starting point.

In terms of regions we want to plan for the maximum number of regions we can be planning for, we can expect to have multiple accounts and 2+ networks per account.

Given all the above we can come up with a plan that we will operate in `5 regions (3 in the US, 1 Europe and 1 in Austraila)`, `4 accounts in each region`, but since we want to have a buffer we create `4 networks per account`, so that will total `80 networks`.

Also knowing that we prefer to start at `10.16.0.0` and we can not use the Google range, that means we can go up to `10.127.0.0` (inclusive).

#### Splitting up the VPC range

The size of the VPC can be determined thinking about how many `subnets`, or better to say how many `AZs` we want to span our VPC across and how many `tiers` we are going to need within the VPC.

It is generally a good rule of thumb to think of starting with `3 AZs` because that works for most regions. It is useful to assume a region can grow to have additional AZs, so we could add an extra to have `4 in total`.

In terms of application tier we can think of the traditional `web`, `app` and `db` tiers, plus here we can also use the logic we used with the AZs, as in adding a spare.

With this we end up having a `4x4 grid`, `16 subnets per VPC`. So if we go with a `/16` network then we get `/20` subnets.

#### How will it look in practice

Given the pre-requisites and the plan we can have our starting ranges at:

- 10.16 (US1)
- 10.32 (US2)
- 10.48 (US3)
- 10.64 (EU)
- 10.80 (Australia)

Which for one region would look like this:

| VPC start | Region  | Account name | VPC name |
| --------- | ------- | ------------ | -------- |
| 10.16     | US1     | General      | VPC1     |
| 10.17     | US1     | General      | VPC2     |
| 10.18     | US1     | General      | VPC3     |
| 10.19     | US1     | General      | VPC4     |
| 10.20     | US1     | Dev          | VPC1     |
| 10.21     | US1     | Dev          | VPC2     |
| 10.22     | US1     | Dev          | VPC3     |
| 10.23     | US1     | Dev          | VPC4     |
| 10.24     | US1     | Prod         | VPC1     |
| 10.25     | US1     | Prod         | VPC2     |
| 10.26     | US1     | Prod         | VPC3     |
| 10.27     | US1     | Prod         | VPC4     |
| 10.28     | US1     | Reserved     | VPC1     |
| 10.29     | US1     | Reserved     | VPC2     |
| 10.30     | US1     | Reserved     | VPC3     |
| 10.31     | US1     | Reserved     | VPC4     |




