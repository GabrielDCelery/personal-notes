---
title: "AWS SSO and permission sets"
date: 2025-10-14
tags: ["aws", "sso"]
---

# The problem

Learn more about how AWS permission sets work

## What is a permission set

An IAM Identity Center permission set is essentially a template of AWS permissions that you can assign to users or groups in AWS IAM Identity Center (formerly AWS SSO).

> [!INFO]
> Think of it like this: instead of managing permissions directly in each AWS account, you create a permission set once that says "here's what someone in this role should be able to do" - it might contain policies that grant EC2 access, S3 read permissions, etc. Then you can assign that permission set to users across multiple AWS accounts in your organization.

What makes it useful:

1. Consistency - You define permissions once and apply them uniformly across accounts
2. Centralized management - No need to create and manage IAM roles separately in each account
3. Temporary credentials - When users access AWS through Identity Center, they get short-lived credentials based on the permission set, which is more secure than long-term access keys
4. Multi-account simplicity - If you have 10 AWS accounts and want to give a developer the same permissions in each, you just assign one permission set 10 times instead of creating 10 separate IAM roles

Basically, it's AWS's way of making it easier to manage "who can do what" across an entire organization of AWS accounts, rather than dealing with individual IAM users and roles in each account separately.

## What can permission sets be assigned to

In IAM Identity Center, a permission set can be assigned to:

1. Users - Individual user accounts from your identity source (Identity Center directory, Active Directory, or external IdP like Okta/Azure AD)
2. Groups - Groups of users from your identity source (this is the recommended approach for easier management)

The assignment always happens in the context of:

- A specific AWS account (or multiple accounts)

So the assignment model is essentially: "Give [User/Group] access to [AWS Account] with [Permission Set]"

Key points:

- You can assign the same permission set to the same user/group across multiple AWS accounts
- You can assign multiple different permission sets to the same user/group in the same account (they'll get the combined permissions)
- Groups are recommended over individual user assignments because they're easier to manage at scale
- The identity source (users and groups) can come from:
  - Identity Center's built-in directory
  - Microsoft Active Directory (via AWS Directory Service)
  - External identity providers (Okta, Azure AD, etc.) via SAML 2.0

When a user signs into the AWS access portal, they'll see all the accounts they have access to, and for each account, they'll see the permission sets (displayed as "roles") they can assume.

## Permission set trust policy in the NOT Iam Identity center account

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::138315362805:saml-provider/AWSSSO_af5o7351e0625db1_DO_NOT_DELETE"
      },
      "Action": ["sts:AssumeRoleWithSAML", "sts:TagSession"],
      "Condition": {
        "StringEquals": {
          "SAML:aud": "https://signin.aws.amazon.com/saml"
        }
      }
    }
  ]
}
```

This trust policy defines who can assume an IAM role - specifically, it allows IAM Identity Center to give users temporary credentials for this role through SAML-based federation.

Let me break it down:

Principal (Who can assume the role):
"Federated": "arn:aws:iam::138315362805:saml-provider/AWSSSO_ec2e7351e0625db1_DO_NOT_DELETE"
This is the SAML identity provider that IAM Identity Center automatically created in your AWS account. It's the trusted entity that can vouch for users' identities.

Actions (What they can do):

- sts:AssumeRoleWithSAML - Allows assuming the role using SAML assertions (the authentication token that proves who the user is)
- sts:TagSession - Allows adding session tags during role assumption (useful for passing user attributes like email, department, etc.)

Condition (Under what circumstances):
"SAML:aud": "https://signin.aws.amazon.com/saml"
This ensures the SAML assertion was intended for AWS's sign-in endpoint, not some other service. It's a security check to prevent SAML assertions meant for other systems from being used to access AWS.

In practice:

1. User logs into IAM Identity Center
2. They select an AWS account and permission set
3. Identity Center generates a SAML assertion saying "this user is authenticated"
4. Identity Center uses that SAML assertion to call AssumeRoleWithSAML against this role
5. AWS validates the assertion came from the trusted SAML provider and returns temporary credentials
6. User can now access AWS with the permissions defined in the role's permission policies

This is the trust policy that IAM Identity Center automatically creates when you assign a permission set to an account - it's how the "magic" of SSO works behind the scenes.

## Where is the federated principal on the AWS account

You can find this SAML identity provider in the IAM console:

1.  Go to the IAM console (not IAM Identity Center)
2.  In the left navigation pane, look for Identity providers (under "Access management")
3.  You should see AWSSSO_ec2e7351e0625db1_DO_NOT_DELETE listed there

What you'll see:

- Provider name: AWSSSO_ec2e7351e0625db1_DO_NOT_DELETE
- Provider type: SAML
- The full ARN: arn:aws:iam::138315362805:saml-provider/AWSSSO_ec2e7351e0625db1_DO_NOT_DELETE
- The SAML metadata document that defines how AWS trusts Identity Center

Important notes:

- As the name suggests (DO_NOT_DELETE), this is automatically managed by IAM Identity Center - don't delete or modify it
- This provider exists in each AWS account where you've assigned Identity Center permission sets
- Each account will have its own identity provider with the same naming pattern
- If you delete it accidentally, Identity Center users won't be able to access that account until it's recreated (usually happens automatically when you reassign permission sets)

The direct URL pattern would be:
https://console.aws.amazon.com/iam/home#/identity_providers

## when I create an identity provider via the aws console in iam center is the account where I am creating the identity provider is the service provider

⏺ Yes, exactly! When you create an identity provider in the IAM console, the AWS account where you're creating it acts as the service provider (SP).

Here's the SAML relationship:

- Identity Provider (IdP) - The external authentication system (Okta, Azure AD, IAM Identity Center, etc.) that verifies who the user is
- Service Provider (SP) - Your AWS account that provides the actual services/resources the user wants to access

The flow:

1. User wants to access AWS (the service provider)
2. AWS redirects them to the IdP for authentication
3. User logs into the IdP (Okta, Azure AD, etc.)
4. IdP sends a SAML assertion back to AWS saying "this user is authenticated"
5. AWS (the SP) trusts that assertion because you configured the identity provider
6. AWS grants temporary access based on the role's permissions

When you configure the identity provider in IAM:

You're telling your AWS account: "Trust SAML assertions from this external IdP." You do this by uploading the IdP's metadata document, which contains:

- The IdP's certificate (for validating signatures)
- The IdP's endpoints
- Other trust information

The reverse configuration:

You also need to configure the external IdP to know about AWS - giving it the AWS SAML endpoint (https://signin.aws.amazon.com/saml) so it knows where to send authenticated users.

So yes - your AWS account = Service Provider, and the external system you're configuring = Identity Provider.
