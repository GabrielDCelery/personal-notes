---
title: Managing project secrets using a vault
author: GaborZeller
date: 2025-03-04T23-42-25Z
tags:
draft: true
---

# Managing project secrets using a vault

1. **AWS Secrets Manager**

   - Cloud-native solution for AWS environments
   - Automatic rotation of secrets
   - Fine-grained IAM permissions
   - Native integration with AWS services

2. **Azure Key Vault**

   - Microsoft's cloud-based key management solution
   - Integrated with Azure AD
   - Hardware Security Module (HSM) backing
   - Certificate management capabilities

3. **Google Cloud Secret Manager**

   - Native to Google Cloud Platform
   - Version control for secrets
   - IAM integration
   - Audit logging

4. **Mozilla SOPS (Secrets OPerationS)**

   - Open-source
   - Works with multiple cloud providers' KMS
   - Can encrypt/decrypt YAML, JSON, ENV files
   - Git-friendly
   - Supports age encryption tool

5. **Sealed Secrets**

   - Kubernetes-native solution
   - Encrypts secrets that are safe to store in git
   - Controller-based approach
   - Open-source

6. **External Secrets Operator**

   - Kubernetes operator pattern
   - Can integrate with multiple backend providers
   - Automatically syncs secrets from external providers
   - Open-source

7. **Doppler**

   - SecretOps platform
   - Universal secrets platform
   - Cross-platform support
   - Developer-friendly UI

8. **1Password Secrets Automation**

   - Built on 1Password's security infrastructure
   - Connect server for automated secrets delivery
   - Familiar interface for teams already using 1Password
   - Audit trails

9. **Akeyless Vault**

   - Cloud-agnostic
   - Zero-trust architecture
   - Distributed fragments architecture
   - Built-in key rotation

10. **Keeper Secrets Manager**
    - Zero-knowledge architecture
    - DevOps and CI/CD integration
    - Role-based access control
    - Audit and compliance features

Key Considerations when choosing:

1. **Hosting Model**: Self-hosted vs Cloud-based
2. **Integration Capabilities**: Native support for your tech stack
3. **Compliance Requirements**: Certifications and standards support
4. **Scaling Needs**: Performance at your required scale
5. **Cost Structure**: Pay-per-secret vs subscription vs self-hosted
6. **Security Features**: Encryption, access control, audit logging
7. **Automation Support**: API accessibility and DevOps integration

For Kubernetes-specific environments, the followings are particularly recommended:

- Sealed Secrets + External Secrets Operator (for a fully Kubernetes-native approach)
- SOPS (for git-based workflows)
- Cloud Provider Solutions (if you're already in a specific cloud ecosystem)
