---
title: "Azure and AWS comparison"
date: 2025-10-23
tags: ["aws", "azure"]
---

# Secret management

| Azure Service          | AWS Equivalent                    | Use Case                                |
| ---------------------- | --------------------------------- | --------------------------------------- |
| Key Vault Secrets      | Secrets Manager / Parameter Store | Passwords, API keys, connection strings |
| Key Vault Keys         | KMS / CloudHSM                    | Encryption keys, signing keys           |
| Key Vault Certificates | Certificate Manager (ACM)         | SSL/TLS certificates                    |
