---
title: "Create an AWS Managed certificate and integrate it with Cloudflare"
date: 2025-10-10
tags: ["certificate", "aws", "cloudflare"]
---

# The problem

Figure out how to create an AWS managed certificate and integrate it with Cloudflare.

## How I solved the problem

### Create the AWS certificate

Went to the AWS Certificate manager (clickops)

1. Request button on upper right
2. Request a public certificate
3. Set attributes for certiciate

- domain names (e.g. \*.mywebsite.com)
- allow export (likely disable export)
- validation method (DNS validation - will need to add record to Cloudflare)
- key algorithm (RSA 2048)

### Create DNS validation on Cloudflare

1. Add the CNAME that we get from the AWS Console to Cloudflare
