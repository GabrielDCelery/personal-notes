---
title: "AWS SSO permission sets env specific problems"
date: 2025-11-18
tags: ["aws", "sso"]
---

# The problem

What if resources got specific naming conventions and that is limiting my options to create permission sets

# Possible solutions

### 1. Name your buckets consistently across environments, then use patterns in the permission set:

```terraform
  # Bucket naming convention:
  # - dev:     "myapp-data-dev"
  # - staging: "myapp-data-staging"
  # - prod:    "myapp-data-prod"

  resource "aws_ssoadmin_permission_set" "data_bucket_access" {
    name        = "DataBucketAccess"
    description = "Access to data buckets following myapp-data-* naming pattern"

    inline_policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Effect = "Allow"
          Action = [
            "s3:GetObject",
            "s3:PutObject",
            "s3:DeleteObject"
          ]
          Resource = "arn:aws:s3:::myapp-data-*/*"  # Matches all environments
        },
        {
          Effect = "Allow"
          Action = [
            "s3:ListBucket"
          ]
          Resource = "arn:aws:s3:::myapp-data-*"  # Bucket-level access
        }
      ]
    })
  }

  # Same permission set works in all environments via assignments
  resource "aws_ssoadmin_account_assignment" "dev" {
    permission_set_arn = aws_ssoadmin_permission_set.data_bucket_access.arn
    principal_id       = data.aws_identitystore_group.data_engineers.group_id
    target_id          = var.account_ids.development  # Has bucket: myapp-data-dev
  }

  resource "aws_ssoadmin_account_assignment" "prod" {
    permission_set_arn = aws_ssoadmin_permission_set.data_bucket_access.arn
    principal_id       = data.aws_identitystore_group.data_engineers.group_id
    target_id          = var.account_ids.production  # Has bucket: myapp-data-prod
  }
```

Pros:

- Truly reusable permission set
- Easy to maintain
- Scales to new environments automatically

Cons:

- Requires enforcing naming conventions
- May grant access to more buckets than intended if naming overlaps
