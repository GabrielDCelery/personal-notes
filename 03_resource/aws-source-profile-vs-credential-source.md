---
title: "AWS source profile vs credential source"
date: 2025-10-13
tags: ["aws", "cli"]
---

## aws credential_source vs source_profile and when to use which

● In AWS configuration, credential_source and source_profile are both used for role assumption, but they specify different ways to obtain the initial credentials:

source_profile

Used when you want to assume a role using credentials from another named profile in your AWS credentials file.

Example:
[profile base]
aws_access_key_id = AKIA...
aws_secret_access_key = ...

[profile role-profile]
role_arn = arn:aws:iam::123456789012:role/MyRole
source_profile = base

When to use:

- You have static credentials (access keys) stored in a profile
- You're switching between multiple AWS accounts using stored credentials
- Local development with multiple profiles

credential_source

Used when you want to assume a role using credentials from environment variables or instance metadata, not from a profile.

Example:
[profile role-profile]
role_arn = arn:aws:iam::123456789012:role/MyRole
credential_source = Ec2InstanceMetadata

Valid values:

- Ec2InstanceMetadata - Get credentials from EC2 instance metadata (IAM role attached to EC2)
- Environment - Get credentials from environment variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
- EcsContainer - Get credentials from ECS task role

When to use:

- Running on EC2 instances with IAM roles
- Running in ECS/Fargate containers with task roles
- Using environment variables for base credentials
- CI/CD pipelines where credentials come from the environment

Key Differences Summary

| Feature           | source_profile                       | credential_source               |
| ----------------- | ------------------------------------ | ------------------------------- |
| Credential source | Named profile in config file         | Environment/metadata            |
| Use case          | Local development, multiple profiles | EC2, ECS, environment variables |
| Value type        | Profile name (string)                | Special keywords                |

You CANNOT use both

These are mutually exclusive - you must choose one or the other for a given role profile. If you try to use both, AWS will return an error.

## how to use credential_source in aws using env variables

● To use credential_source with environment variables in AWS, follow these steps:

1. Set up your AWS config file

Edit ~/.aws/config:

[profile my-role]
role_arn = arn:aws:iam::123456789012:role/MyRole
credential_source = Environment
region = us-east-1

2. Export your AWS credentials as environment variables

export AWS_ACCESS_KEY_ID=AKIA...
export AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG...

# Optional, if using temporary credentials:

export AWS_SESSION_TOKEN=FwoGZXIvYXdzEBYaD...

3. Use the profile

# Using AWS CLI

aws s3 ls --profile my-role

# Or set as default profile

export AWS_PROFILE=my-role
aws s3 ls

How it works

When you use the profile:

1. AWS SDK/CLI reads the profile configuration
2. Sees credential_source = Environment
3. Retrieves base credentials from environment variables
4. Uses those credentials to assume the role specified in role_arn
5. Returns temporary credentials for the assumed role

Common use cases

CI/CD pipelines:

# GitHub Actions example

- name: Assume role
  env:
  AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
  AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
  run: |
  aws sts get-caller-identity --profile my-role

Docker containers:
docker run -e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID \
             -e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY \
 -v ~/.aws:/root/.aws:ro \
 my-image

The key advantage is you don't need to store long-term credentials in your config file - they come from the environment, and you can still use role assumption for enhanced security.
