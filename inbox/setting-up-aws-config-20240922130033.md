---
title: Setting up aws config
author: GaborZeller
date: 2024-09-22T13-00-33Z
tags:
draft: true
---

# Setting up aws config

Example of how to set up aws config when we have an IAM user with an access key and security key

```sh
# profile iamadmin-management - the profile for the aws-cli
# region - region where we want the profile to operate
# mfa_serial - the aws resoruce name of the MFA device that has been set up for the IAM user (you still need an IAM Policy that enforces MFA use if you want the benefit)
# 100848142372 - the account ID
# imadmin - the name of the user that was set up on IAM
# training-awsdevops - the alias of the account ID that makes it easier to identify the account

[profile imadmin-management]
region=eu-west-2
mfa_serial=arn:aws:iam::100848142372:mfa/iamadmin@training-awsdevops-100848142372
```

Example of how to set up aws config where user is using SSO and the IAM Identity Center serves as the identity provider

```sh
# profile sso-management -management - the profile for the aws-cli
# sso_start_url - the SSO login url
# sso_region = the region where the IAM Identity center has been set up
# sso_account_id = the account id whete the IAM Identity center was set up
# sso_role_name = the role that was set up for the user that was added to the IAM Identity center
# region - region where we want the profile to operate
# 100848142372 - the account ID

[profile sso-management]
sso_start_url = https://somenameyouchose.awsapps.com/start
sso_region = eu-west-2
sso_account_id = 100848142372
sso_role_name = AdministratorAccess
region = eu-west-2
```