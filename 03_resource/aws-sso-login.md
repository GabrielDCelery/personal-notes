---
title: "AWS SSO login"
date: 2025-10-14
tags: ["aws", "sso"]
---

aws sso login only authenticates you with AWS IAM Identity Center - it doesn't export AWS credentials to environment variables like AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, or AWS_SESSION_TOKEN.

Here's what actually happens:

What aws sso login does:

1. Opens your browser to authenticate with IAM Identity Center
2. Stores the SSO session token in ~/.aws/sso/cache/
3. That's it - no credentials are exported

How the AWS CLI uses SSO:

When you run a command like aws s3 ls --profile gz-management:

1. The CLI reads your profile configuration
2. Sees it's an SSO profile
3. Checks if there's a valid SSO session token cached
4. Uses that token to request temporary credentials for the role
5. Uses those temporary credentials for the command
6. Caches those credentials briefly

The credentials never touch your environment variables - the AWS CLI handles everything internally.

If you need environment variables:

If you want to export credentials (for tools that don't support SSO profiles natively), you need to explicitly export them after logging in:

# First, login

aws sso login --profile gz-management

# Then export credentials to environment variables

eval $(aws configure export-credentials --profile gz-management --format env)

# Or manually:

export AWS_ACCESS_KEY_ID=$(aws configure get aws_access_key_id --profile gz-management)
  export AWS_SECRET_ACCESS_KEY=$(aws configure get aws_secret_access_key --profile gz-management)
export AWS_SESSION_TOKEN=$(aws configure get aws_session_token --profile gz-management)

This is by design - SSO credentials are temporary and managed by the CLI, not static credentials you set once.

Would you like me to add this explanation to your notes?
