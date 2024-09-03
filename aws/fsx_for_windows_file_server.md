---
title: FSX for Windows File Server
author: Gabor Zeller
date: 2024-09-01T16:46:13Z
tags: ['aws']
draft: true
---

# FSX for Windows File Server

#### What does the product offer?

FSx for Windows is a native (not emulated!) file server/share that is implemented similar to RDS where the inner workings are hidden.

It has been designed to integrate with Windows environments. Works both with AWS managed or on-premises directory service.

It is designed to be highly available from the ground-up, can either be deployed in Single-AZ or Multi-AZ. Does do data replication even if it has been deployed in a Single-AZ in case there is a hardware failiure.

Can have scheduled or on-demand backups.

Can be shared using VPC peering connection, VPN or even Direct Connect.

FSx is a `native file system` which mans it supports everything you would expect from an on-premises Windows file system. For example `DFS (Distributed File System)`, encryption at rest and in transit.

#### How would it be implemented

Usually we will have the AWS network and an on-premises network that are either connected via Diect Connect or VPN. We will have a Directory Service which is either the company's own Directory Service or a managed service within AWS.

Once the FSx has been set up it can be shared via a `network share`, for eample `\\fs-xxx.animalsforlife.org\catpics`.

Systems that use a Windows based system can access this storage, for exmample AWS Workspaces which is a virtual desktop service similar to Citrix which is also an AWS offering.

#### Some numbers worth remembering

Data transfer speed can vary between `8MB/s` or `2GB/s`, `100k's IOPS` and less than `1ms latency`.

#### Exam tips

Both EFS and FSx are shared network file systems so the question will probably be around being able to differentiate between them.

Look for words like `Directory Service` or `Windows environment`.

You can also have file-level restores that can be initiated by right clicking on the file rather than going into the AWS system or talking to a system administrator. So if you see `VSS` restore that is an indication of FSx over EFS.

The native file system is accessible over the `SMB` protocol. This is important because Linux uses the `NFS` protocol.

Uses the `Windows permission model` for file system access.

`DFS` is a way to replicate or natively scale out file systems in a Windows environment, so if this is mentioned then the answer is probably FSx.

It is also a `fully managed` service so if we want to reduce admin overhead and not spin up EC2 instances with Windows Servers then this is also an indication.
