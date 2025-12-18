```terraform
  ---
  ✅ Option 1: Use Naming Conventions + Wildcards (Best)

  Name your buckets consistently across environments, then use patterns in the permission set:

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

  Pros:
  - Truly reusable permission set
  - Easy to maintain
  - Scales to new environments automatically

  Cons:
  - Requires enforcing naming conventions
  - May grant access to more buckets than intended if naming overlaps

  ---
  ✅ Option 2: Use Resource Tags (More Flexible)

  Tag your buckets consistently, reference by tags in the policy:

  # Tag your buckets:
  # dev:     { Environment = "dev", DataClassification = "internal" }
  # prod:    { Environment = "prod", DataClassification = "internal" }

  resource "aws_ssoadmin_permission_set" "data_bucket_access" {
    name = "DataBucketAccess"

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
          Resource = "*"
          Condition = {
            StringEquals = {
              "s3:ExistingObjectTag/DataClassification" = "internal"
            }
          }
        },
        {
          Effect = "Allow"
          Action = ["s3:ListBucket"]
          Resource = "*"
          Condition = {
            StringEquals = {
              "aws:ResourceTag/DataClassification" = "internal"
            }
          }
        }
      ]
    })
  }

  Pros:
  - More flexible than naming patterns
  - Can control access by multiple tag dimensions
  - Bucket names can be different

  Cons:
  - Requires consistent tagging discipline
  - More complex IAM conditions
  - Not all S3 actions support tag conditions

  ---
  ✅ Option 3: Use Customer Managed Policies Per Account

  Create account-specific IAM policies, attach them to a permission set:

  # Permission set with no inline policy
  resource "aws_ssoadmin_permission_set" "data_bucket_access" {
    name        = "DataBucketAccess"
    description = "Access to environment-specific data buckets via customer managed policies"
  }

  # In each AWS account, create a customer managed policy
  # (This would be in a separate Terraform workspace/module per account)

  # In DEV account (123456789012):
  resource "aws_iam_policy" "dev_bucket_policy" {
    provider = aws.dev
    name     = "DataBucketAccess"

    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [{
        Effect   = "Allow"
        Action   = ["s3:*"]
        Resource = [
          "arn:aws:s3:::weird-dev-bucket-name-xyz",
          "arn:aws:s3:::weird-dev-bucket-name-xyz/*"
        ]
      }]
    })
  }

  # In PROD account (987654321098):
  resource "aws_iam_policy" "prod_bucket_policy" {
    provider = aws.prod
    name     = "DataBucketAccess"  # Same name, different content

    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [{
        Effect   = "Allow"
        Action   = ["s3:*"]
        Resource = [
          "arn:aws:s3:::completely-different-prod-name",
          "arn:aws:s3:::completely-different-prod-name/*"
        ]
      }]
    })
  }

  # Attach customer managed policies to permission set
  resource "aws_ssoadmin_customer_managed_policy_attachment" "dev" {
    instance_arn       = local.sso_instance_arn
    permission_set_arn = aws_ssoadmin_permission_set.data_bucket_access.arn

    customer_managed_policy_reference {
      name = "DataBucketAccess"
      path = "/"
    }
  }

  How it works:
  - Permission set has no inline policy (or just base permissions)
  - Each AWS account has a customer managed policy with the same name
  - Permission set references the policy by name
  - In each account, the policy has different bucket ARNs

  Pros:
  - Truly flexible - bucket names can be anything
  - Reusable permission set
  - Clear separation of concerns

  Cons:
  - More complex setup (policies in each account)
  - Need to manage policies separately from permission sets
  - Requires Terraform to manage multiple accounts

  ---
  ⚠ Option 4: Session Tags (Advanced)

  Use session tags to dynamically control access:

  resource "aws_ssoadmin_permission_set" "data_bucket_access" {
    name = "DataBucketAccess"

    inline_policy = jsonencode({
      Version = "2012-10-17"
      Statement = [{
        Effect = "Allow"
        Action = ["s3:*"]
        Resource = [
          "arn:aws:s3:::*-$${aws:PrincipalTag/Environment}",
          "arn:aws:s3:::*-$${aws:PrincipalTag/Environment}/*"
        ]
      }]
    })
  }

  # Configure permission set to pass environment as session tag
  # (Session tags would be configured in IAM Identity Center or through SAML attributes)

  Pros:
  - Very dynamic
  - One permission set, truly environment-aware

  Cons:
  - Complex to set up
  - Requires session tag configuration in Identity Center
  - Harder to debug
  - Still requires bucket naming conventions

  ---
  🤷 Option 5: Accept Environment-Specific Permission Sets

  Sometimes it's OK to have different permission sets if resources are truly different:

  resource "aws_ssoadmin_permission_set" "dev_legacy_bucket" {
    name = "DevLegacyBucketAccess"

    inline_policy = jsonencode({
      Statement = [{
        Effect   = "Allow"
        Action   = ["s3:*"]
        Resource = [
          "arn:aws:s3:::weird-old-dev-bucket-12345",
          "arn:aws:s3:::weird-old-dev-bucket-12345/*"
        ]
      }]
    })
  }

  resource "aws_ssoadmin_permission_set" "prod_new_bucket" {
    name = "ProdModernBucketAccess"

    inline_policy = jsonencode({
      Statement = [{
        Effect   = "Allow"
        Action   = ["s3:*"]
        Resource = [
          "arn:aws:s3:::prod-bucket-v2",
          "arn:aws:s3:::prod-bucket-v2/*"
        ]
      }]
    })
  }

  When this is acceptable:
  - Legacy systems with different architectures per environment
  - Temporary during migration
  - Genuinely different resources (not just naming)

  Still avoid if possible - but don't over-engineer if it's truly a one-off situation.

  ---
  🎯 Recommended Approach:

  For new systems: Use Option 1 (Naming Conventions) or Option 2 (Tags)
  - Easiest to implement and maintain
  - Forces good organizational practices

  For existing systems with inconsistent naming: Use Option 3 (Customer Managed Policies)
  - More setup but very flexible
  - Clean separation of concerns

  Avoid: Option 5 unless absolutely necessary

  ---
  📝 Practical Example: Multi-Environment with Naming Pattern

  # Define reusable permission set
  resource "aws_ssoadmin_permission_set" "data_engineer" {
    name        = "DataEngineer"
    description = "Data engineering access to data-* buckets and related services"

    inline_policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Sid    = "DataBucketAccess"
          Effect = "Allow"
          Action = ["s3:*"]
          Resource = [
            "arn:aws:s3:::data-*",      # Any bucket starting with data-
            "arn:aws:s3:::data-*/*"
          ]
        },
        {
          Sid    = "GlueAccess"
          Effect = "Allow"
          Action = ["glue:*"]
          Resource = "*"
        }
      ]
    })
  }

  # Assign to all environments
  locals {
    environments = {
      dev     = var.account_ids.development
      staging = var.account_ids.staging
      prod    = var.account_ids.production
    }
  }

  resource "aws_ssoadmin_account_assignment" "data_engineer" {
    for_each = local.environments

    instance_arn       = local.sso_instance_arn
    permission_set_arn = aws_ssoadmin_permission_set.data_engineer.arn
    principal_id       = data.aws_identitystore_group.data_engineers.group_id
    principal_type     = "GROUP"
    target_id          = each.value
    target_type        = "AWS_ACCOUNT"
  }

  # In each account, buckets follow pattern:
  # dev:     data-raw-dev, data-processed-dev
  # staging: data-raw-staging, data-processed-staging
  # prod:    data-raw-prod, data-processed-prod

  This single permission set works across all environments because the bucket naming is consistent!

```
