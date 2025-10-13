---
title: "Terraform store profile setting"
date: 2025-10-13
tags: ["aws", "terraform"]
---

# What does the terraform profile configuration do for the backend

```terraform
{
  required_providers {
    databricks = {
      source = "databricks/databricks"
    }
  }
  backend "s3" {
    profile              = "terraform"

```

In this Terraform configuration, the profile parameter in the S3 backend specifies which AWS credentials profile Terraform should use to access the S3 bucket for storing state.

How it works

When you run terraform init or any Terraform command that needs to access the backend:

1. Terraform reads the AWS credentials from ~/.aws/credentials or ~/.aws/config
2. Looks for a profile named "terraform"
3. Uses those credentials to authenticate to AWS
4. Accesses the S3 bucket to read/write the Terraform state file

Expected AWS configuration

You should have a profile named terraform in your AWS files:

```sh
# ~/.aws/credentials:
[terraform]
aws_access_key_id = AKIA...
aws_secret_access_key = ...
```

```sh
# ~/.aws/config (if using role assumption):
[profile terraform]
role_arn = arn:aws:iam::123456789012:role/TerraformRole
source_profile = default
region = us-east-1
```

What happens without the profile parameter

If you omit profile, Terraform will use the default credential chain:

1. Environment variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
2. The default profile from credentials file
3. EC2 instance metadata (if running on EC2)
4. ECS task role (if running in ECS)

Important notes

- The profile parameter in the backend block only affects backend operations (state storage/retrieval)
- It does not affect the AWS provider used for creating resources
- If you want to use the same profile for the AWS provider, you'd specify it separately:

```terraform
terraform {
  backend "s3" {
    profile = "terraform"
    bucket = "my-terraform-state"
    key = "prod/terraform.tfstate"
    region = "us-east-1"
  }
}
```

```terraform
provider "aws" {
  profile = "terraform" # Separate configuration for resources
  region = "us-east-1"
}
```

This separation allows you to store state in one AWS account (using the backend profile) while managing resources in a different account (using the provider profile).
