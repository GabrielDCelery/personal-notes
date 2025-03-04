---
title: Managing project secrets locally or at service level
author: GaborZeller
date: 2025-03-04T23-00-54Z
tags:
draft: true
---

# Managing project secrets locally or at a service level

1. Directory Structure:

```sh
.cluster-secrets/
└── gaze-homelab/
    ├── talosconfig
    ├── kubeconfig
    └── controlplane.yaml
```

2. Security Recommendations:

   a. File Permissions:

   - Set restrictive permissions: `chmod 600` for files
   - Set restrictive directory permissions: `chmod 700` for directories

   b. Storage Location:

   - Store in your home directory (e.g., `~/.cluster-secrets/`)
   - Never commit to public repositories
   - Consider using encrypted storage

3. Distribution Methods:

   a. Team Access:

   - Use a password manager with secure note sharing
   - Use HashiCorp Vault or similar secret management system
   - Use encrypted git repository with git-crypt

4. Backup Strategy:

   - Keep encrypted backups
   - Use tools like age or gpg for encryption
   - Store backups in secure location

5. Access Control:
   - Maintain an access log
   - Rotate credentials periodically
   - Use separate configurations for different environments

Here's how to set it up:

```sh
# Create directory structure
mkdir -p ~/.cluster-secrets/gaze-homelab

# Move files
mv talosos/talosconfig ~/.cluster-secrets/gaze-homelab/
mv talosos/kubeconfig ~/.cluster-secrets/gaze-homelab/
mv talosos/controlplane.yaml ~/.cluster-secrets/gaze-homelab/

# Set proper permissions
chmod 700 ~/.cluster-secrets
chmod 700 ~/.cluster-secrets/gaze-homelab
chmod 600 ~/.cluster-secrets/gaze-homelab/*

# Optional: Encrypt the directory
# Using age encryption (install age first)
tar czf - ~/.cluster-secrets | age -e -r "age1..." > cluster-secrets.tar.gz.age
```

6. Additional Security Measures:

   - Use environment variables to point to config locations
   - Create symbolic links from standard locations
   - Document the location in your team's secure wiki
   - Consider using tools like SOPS or sealed-secrets for Kubernetes secrets

7. Example environment setup:

```sh
# Add to your ~/.bashrc or ~/.zshrc
export TALOS_CONFIG=~/.cluster-secrets/gaze-homelab/talosconfig
export KUBECONFIG=~/.cluster-secrets/gaze-homelab/kubeconfig
```

Remember:

- Never store these files in public repositories
- Always use encryption for backups and distribution
- Regularly rotate credentials
- Maintain an access control list
- Document the recovery process
- Keep the directory structure simple and clear
