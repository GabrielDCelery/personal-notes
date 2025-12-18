---
title: "AWS SSO permission set best practices"
date: 2025-11-18
tags: ["aws", "sso"]
---

# The problem

How to organize my permission sets

## The simple mental model

Think of it like this = WHO + WHAT + WHERE

- WHO: User or Group from the Identity Store
- WHAT: Permission Set (defines the IAM policies)
- WHERE: AWS Account ID (where the permissions apply)

Example:

```terraform
resource "aws_ssoadmin_account_assignment" "example" {
  # WHAT
  instance_arn       = local.sso_instance_arn
  permission_set_arn = aws_ssoadmin_permission_set.eng_admin.arn

  # WHO
  principal_id   = data.aws_identitystore_group.engineering_admins.group_id
  principal_type = "GROUP"

  # WHERE
  target_id   = var.account_ids.engineering.production
  target_type = "AWS_ACCOUNT"
}
```

This says: "The EngineeringAdmins group gets the eng_admin permission set in the production account"

## What is possible with permission sets

| Scenario                                       | Possible? | Example Use Case                          |
| ---------------------------------------------- | --------- | ----------------------------------------- |
| 1 Permission Set → Many Groups                 | ✅ YES    | All teams get ReadOnly access             |
| 1 Group → Many Permission Sets                 | ✅ YES    | Admins get both ReadOnly + AdminAccess    |
| 1 Group → Many Accounts                        | ✅ YES    | DevOps team across all environments       |
| 1 Group → Different Permissions per Account    | ✅ YES    | Admin in dev, ReadOnly in prod            |
| Many Groups → 1 Permission Set → Many Accounts | ✅ YES    | Multiple teams, same access, all accounts |

### 1. One Permission Set → Multiple Groups

```terraform
# ReadOnlyAccess permission set assigned to multiple groups

resource "aws_ssoadmin_account_assignment" "developers_readonly" {
  permission_set_arn = aws_ssoadmin_permission_set.readonly.arn
  principal_id       = data.aws_identitystore_group.developers.group_id
  target_id          = var.account_ids.production
}

resource "aws_ssoadmin_account_assignment" "analysts_readonly" {
  permission_set_arn = aws_ssoadmin_permission_set.readonly.arn  # Same permission set
  principal_id       = data.aws_identitystore_group.analysts.group_id # Different group
  target_id          = var.account_ids.production # Same account
}
```

### 2. One Group → Multiple Permission Sets

```terraform
# Engineers group gets both ReadOnly AND Deploy permission sets

resource "aws_ssoadmin_account_assignment" "eng_readonly" {
  permission_set_arn = aws_ssoadmin_permission_set.readonly.arn # Read only permission set
  principal_id       = data.aws_identitystore_group.engineers.group_id
  target_id          = var.account_ids.production
}

resource "aws_ssoadmin_account_assignment" "eng_deploy" {
  permission_set_arn = aws_ssoadmin_permission_set.deploy.arn  # Different permission set for deployments
  principal_id       = data.aws_identitystore_group.engineers.group_id  # Same group
  target_id          = var.account_ids.production # Same account
}
```

### 3. Same Group + Same Permission Set → Multiple Accounts

```terraform
# Engineers get AdminAccess in all environments

resource "aws_ssoadmin_account_assignment" "dev" {
  permission_set_arn = aws_ssoadmin_permission_set.admin.arn
  principal_id       = data.aws_identitystore_group.engineeers.group_id
  target_id          = var.account_ids.development
}

resource "aws_ssoadmin_account_assignment" "prod" {
  permission_set_arn = aws_ssoadmin_permission_set.admin.arn  # Same permission set
  principal_id       = data.aws_identitystore_group.engineeers.group_id  # Same group
  target_id          = var.account_ids.production  # Different account
}
```

### 4. Different Permission Sets per Account for Same Group

```terraform
# Engineers get Admin in dev, but only ReadOnly in prod

resource "aws_ssoadmin_account_assignment" "dev_admin" {
  permission_set_arn = aws_ssoadmin_permission_set.admin.arn
  principal_id       = data.aws_identitystore_group.engineers.group_id
  target_id          = var.account_ids.development
}

resource "aws_ssoadmin_account_assignment" "prod_readonly" {
  permission_set_arn = aws_ssoadmin_permission_set.readonly.arn  # Different permission
  principal_id       = data.aws_identitystore_group.engineers.group_id  # Same group
  target_id          = var.account_ids.production  # Different account
}
```
