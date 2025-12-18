---
title: "AWS SSO permission sets bad vs good approaches"
date: 2025-11-18
tags: ["aws", "sso"]
---

# The problem

Some anti patterns to avoid

## Summary of anti-patterns

| Anti-Pattern              | Problem             | Better Approach                                  |
| ------------------------- | ------------------- | ------------------------------------------------ |
| Environment-specific sets | Duplication         | Use assignments to control environment access    |
| Team/person-specific sets | Not reusable        | Permission sets = job functions, groups = people |
| One admin for all         | No least privilege  | Granular, workload-based sets                    |
| Hardcoded account ARNs    | Can't reuse         | Use patterns or account variables                |
| Too many micro-sets       | Management overhead | Group by workload (5-15 sets typical)            |
| Reinventing AWS policies  | Maintenance burden  | Use AWS managed policies when possible           |
| Long admin sessions       | Security risk       | 1-2h for high privilege, 8h for read-only        |
| No boundaries             | Unrestricted access | Use permission boundaries + SCPs                 |
| Inconsistent naming       | Confusing           | Pick convention, stick to it                     |
| No documentation          | Knowledge loss      | Descriptions + tags on everything                |

## Bad vs Good examples

### 1. Environment-Specific Permission Sets

❌ Bad:

```terraform
resource "aws_ssoadmin_permission_set" "dev_admin" {
  name = "DevAdmin"
}

resource "aws_ssoadmin_permission_set" "staging_admin" {
  name = "StagingAdmin"
}

resource "aws_ssoadmin_permission_set" "prod_admin" {
  name = "ProdAdmin"
}
```

Why it's bad:

- Creates unnecessary duplication
- Hard to maintain (3x the permission sets to update)
- Doesn't scale (what about QA, UAT, sandbox accounts?)
- Environment should be controlled via assignment, not permission set name

✅ Better:

```terraform
resource "aws_ssoadmin_permission_set" "admin" {
  name = "AdministratorAccess"
}

# Control who gets it WHERE through assignments
resource "aws_ssoadmin_account_assignment" "dev" {
  permission_set_arn = aws_ssoadmin_permission_set.admin.arn
  target_id          = var.account_ids.development  # Environment controlled here
}
```

### 2. Team/Person-Specific Permission Sets

❌ Bad:

```terraform
resource "aws_ssoadmin_permission_set" "engineering_team" {
  name = "EngineeringTeamAccess"
}

resource "aws_ssoadmin_permission_set" "platform_team" {
  name = "PlatformTeamAccess"
}

resource "aws_ssoadmin_permission_set" "alice_permissions" {
  name = "AliceCustomAccess"
}
```

Why it's bad:

- Violates separation of concerns (identity vs permissions)
- Can't reuse across teams
- Becomes a nightmare when people change teams
- Encourages snowflake/custom permissions per person
- What happens when Alice leaves?

✅ Better:

```terraform
# Permission sets are job functions
resource "aws_ssoadmin_permission_set" "data_pipeline_admin" {
  name = "DataPipelineAdmin"
}

# Groups represent teams
data "aws_identitystore_group" "data_engineering" {
  # ...
}

# Assignment connects them
resource "aws_ssoadmin_account_assignment" "example" {
  permission_set_arn = aws_ssoadmin_permission_set.data_pipeline_admin.arn
  principal_id       = data.aws_identitystore_group.engineering.group_id
  target_id          = var.account_ids.production
}

```

### 3. One Giant "Admin" Permission Set for Everything

❌ Bad:

```terraform
  resource "aws_ssoadmin_permission_set" "admin" {
  name = "Admin"
}

# Everyone gets full admin
resource "aws_ssoadmin_account_assignment" "devs" {
  permission_set_arn = aws_ssoadmin_permission_set.admin.arn
  principal_id       = data.aws_identitystore_group.developers.group_id
}

resource "aws_ssoadmin_account_assignment" "data_eng" {
  permission_set_arn = aws_ssoadmin_permission_set.admin.arn
  principal_id       = data.aws_identitystore_group.data_engineers.group_id
}

```

Why it's bad:

- Violates least privilege
- No separation of duties
- Can't track who needs what
- Security/compliance nightmare
- Everyone can delete production databases

✅ Better:

```terraform
  # Granular permission sets
resource "aws_ssoadmin_permission_set" "application_deployer" { }
resource "aws_ssoadmin_permission_set" "data_pipeline_admin" { }
resource "aws_ssoadmin_permission_set" "full_admin" { }  # Only for true admins

# Assign based on actual needs

```

### 4. Account scoped inline policies

❌ Bad:

```terraform
  resource "aws_ssoadmin_permission_set" "s3_access" {
  name = "S3Access"

  inline_policy = jsonencode({
    Statement = [{
      Effect = "Allow"
      Action = "s3:*"
      Resource = "arn:aws:s3:::prod-bucket-12345/*"  # Hardcoded account-specific resource
    }]
  })
}

```

Why it's bad:

- Can't reuse across accounts (hardcoded ARN)
- Defeats the purpose of reusable permission sets
- Need separate permission sets for each environment

✅ Better:

```terraform
  resource "aws_ssoadmin_permission_set" "s3_data_access" {
  name = "S3DataAccess"

  inline_policy = jsonencode({
    Statement = [{
      Effect = "Allow"
      Action = "s3:*"
      Resource = [
        "arn:aws:s3:::data-*/*",  # Pattern-based, works across accounts
        "arn:aws:s3:::data-*"
      ]
    }]
  })
}

# Or use customer managed policies with ${aws:PrincipalAccount} variables.
```

### 5. Too Many Granular Permission Sets

❌ Bad:

```terraform
  resource "aws_ssoadmin_permission_set" "s3_read" { }
 resource "aws_ssoadmin_permission_set" "s3_write" { }
 resource "aws_ssoadmin_permission_set" "glue_read" { }
 resource "aws_ssoadmin_permission_set" "glue_write" { }
 resource "aws_ssoadmin_permission_set" "athena_read" { }
 resource "aws_ssoadmin_permission_set" "athena_write" { }
 resource "aws_ssoadmin_permission_set" "lambda_read" { }
 # ... 50 more permission sets

```

Why it's bad:

- Users end up with 10+ permission sets each
- Creates combinatorial explosion of assignments
- Hard to understand what access someone actually has
- Management overhead is huge

✅ Better:

```terraform
  # Group related services by workload
resource "aws_ssoadmin_permission_set" "data_pipeline_admin" {
  # Includes S3, Glue, Athena together
}

resource "aws_ssoadmin_permission_set" "data_analyst" {
  # Read-only to S3, Glue, Athena
}
```

### 6. Not using AWS Managed policies

❌ Bad:

```terraform
  resource "aws_ssoadmin_permission_set" "readonly" {
  name = "CustomReadOnly"

  inline_policy = jsonencode({
    Statement = [{
      Effect = "Allow"
      Action = [
        "ec2:Describe*",
        "s3:Get*",
        "s3:List*",
        "rds:Describe*",
        # ... manually listing 500 read-only actions
      ]
      Resource = "*"
    }]
  })
}

```

Why it's bad:

- AWS maintains managed policies with new services
- You'll miss new services or permissions
- More code to maintain

✅ Better:

```terraform
  resource "aws_ssoadmin_permission_set" "readonly" {
  name = "ReadOnlyAccess"
}

resource "aws_ssoadmin_managed_policy_attachment" "readonly" {
  instance_arn       = local.sso_instance_arn
  permission_set_arn = aws_ssoadmin_permission_set.readonly.arn
  managed_policy_arn = "arn:aws:iam::aws:policy/ReadOnlyAccess"  # AWS-maintained
}

```

### 7. Long Session Durations for High Privileges

❌ Bad:

```terraform
  resource "aws_ssoadmin_permission_set" "full_admin" {
  name             = "AdministratorAccess"
  session_duration = "PT12H"  # 12 hour admin sessions!
}

```

Why it's bad:

- Compromised credentials valid for 12 hours
- Reduces accountability (long sessions = less re-authentication)
- Violates security best practices

✅ Better:

```terraform
resource "aws_ssoadmin_permission_set" "full_admin" {
  name             = "AdministratorAccess"
  session_duration = "PT1H"  # 1-2 hours max for admin
}

resource "aws_ssoadmin_permission_set" "readonly" {
  name             = "ReadOnlyAccess"
  session_duration = "PT8H"  # Longer OK for low-privilege
}
```

### 8. No Permission Boundaries or SCPs

❌ Bad:

```terraform
  # Give everyone admin, hope they don't break things
 resource "aws_ssoadmin_permission_set" "developer" {
   inline_policy = jsonencode({
     Statement = [{
       Effect   = "Allow"
       Action   = "*"
       Resource = "*"
     }]
   })
 }

```

Why it's bad:

- Can delete CloudTrail logs
- Can modify IAM Identity Center itself
- Can disable GuardDuty
- Can spin up expensive resources

✅ Better:

```terraform
  resource "aws_ssoadmin_permission_set" "developer" {
  inline_policy = jsonencode({
    Statement = [{
      Effect   = "Allow"
      Action   = "*"
      Resource = "*"
    }]
  })

  # Use permission boundaries
  permissions_boundary {
    managed_policy_arn = "arn:aws:iam::aws:policy/PowerUserAccess"
  }
}

# Plus use SCPs at org level to prevent:
# - Leaving organization
# - Disabling CloudTrail
# - Modifying IAM Identity Center
```

### 9. Inconsistent Naming Conventions

❌ Bad:

```terraform
  resource "aws_ssoadmin_permission_set" "access1" {
  name = "readonly"  # lowercase
}

resource "aws_ssoadmin_permission_set" "access2" {
  name = "Deploy-Access"  # kebab-case
}

resource "aws_ssoadmin_permission_set" "access3" {
  name = "Data_Pipeline_ADMIN"  # mixed case with underscores
}

```

Why it's bad:

- Hard to search/filter
- Looks unprofessional
- Creates confusion

✅ Better:

```terraform
  # Pick a convention and stick to it
resource "aws_ssoadmin_permission_set" "read_only" {
  name = "ReadOnlyAccess"  # PascalCase
}

resource "aws_ssoadmin_permission_set" "deploy" {
  name = "ApplicationDeployer"  # PascalCase
}

resource "aws_ssoadmin_permission_set" "data_admin" {
  name = "DataPipelineAdmin"  # PascalCase
}
```

### 10. No Descriptions or Documentation

❌ Bad:

```terraform
  resource "aws_ssoadmin_permission_set" "ps1" {
  name = "PS1"
}

resource "aws_ssoadmin_permission_set" "custom_access" {
  name = "CustomAccess"
}


```

Why it's bad:

- No one knows what it does 6 months later
- Can't determine if still needed
- Onboarding new team members is painful

✅ Better:

```terraform
  resource "aws_ssoadmin_permission_set" "data_pipeline_admin" {
  name        = "DataPipelineAdmin"
  description = "Full admin access to AWS Glue, Athena, Step Functions, and data-* S3 buckets. Used by data engineering team for pipeline development and troubleshooting."

  # tags for better organization
  tags = {
    ManagedBy   = "Terraform"
    Team        = "Platform"
    Sensitivity = "High"
  }
}

```
