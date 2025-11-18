---
title: "Create certificate request using SSH private key"
date: 2025-11-05
tags: ["openssl", "ssh"]
---

# The problem

From a cryptographic perspective SSH keys and OpenSSH keys are generated the same way.

Which means it is possible to create a certificate request using an SSH key.

# The solution

In order to achieve that the SSH key has to be converted to a format that `openssl` can work with.

1. SSH private keys (especially modern ones) use the OpenSSH format, which starts with:

```sh
-----BEGIN OPENSSH PRIVATE KEY-----
```

2. While OpenSSL expects keys in PKCS#1 or PKCS#8 format that start with:

```sh
-----BEGIN RSA PRIVATE KEY-----
or
-----BEGIN PRIVATE KEY-----
```

3. Convert OpenSSH format to PEM format (if needed):

```sh
ssh-keygen -p -m PEM -f gabessh.key # Convert in-place
cp gabessh.key gabessh_pem.key && ssh-keygen -p -m PEM -f gabessh_pem.key -N "" # create a converted copy
```

4. After that key can be used to create a certificate signing request

```sh
openssl req -new -key gabessh_pem.key
```
