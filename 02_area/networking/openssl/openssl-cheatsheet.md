---
title: "OpenSSL cheatsheet"
date: 2025-11-04
tags: ["openssl"]
---

# OpenSSL cheatsheet

The general structure of OpenSSL.

```sh
openssl <operation> <input/output flags> <algorithm flags> <other options>
```

# Common Operations

Think of these as "what you want to do":

| Operation            | Description                                        |
| -------------------- | -------------------------------------------------- |
| `genrsa` / `genpkey` | Generate a private key                             |
| `rsa` / `pkey`       | Work with keys (view, convert, extract public key) |
| `req`                | Request a certificate (or create self-signed)      |
| `x509`               | Work with X.509 certificates (view, sign, convert) |
| `enc`                | Encrypt/decrypt files                              |
| `dgst`               | Create digests/hashes                              |
| `s_client`           | Test SSL/TLS connections (like telnet for HTTPS)   |
| `verify`             | Verify certificates                                |

# Common Flags (The Helpers)

These appear everywhere:

| Flag        | Description                                       |
| ----------- | ------------------------------------------------- |
| `-in file`  | Input from file                                   |
| `-out file` | Output to file                                    |
| `-text`     | Show human-readable text output                   |
| `-noout`    | Don't output the encoded version (just show text) |
| `-new`      | Create something new                              |
| `-key file` | Specify a key file                                |
| `-days N`   | Valid for N days                                  |

# Common commands

## Generating and viewing keys

```sh
# Generate an RSA private key
openssl genrsa -out private.key 2048

# Extract public key from private key
openssl rsa -in private.key -pubout -out public.key
```

## Working with certificates

```sh
# Create self signed certificate
openssl req -new -x509 -key private.key -out cert.pem -days 365

# Create certificate signing request
openssl req -new -key private.key -out request.csr

# View CSR content
openssl req -in request.csr -text -noout

# View certificate details
openssl x509 -in cert.pem -text -noout

# Check certificate expiration
openssl x509 -in cert.pem -noout -dates

# Convert PEM to DER
openssl x509 -in cert.pem -outform DER -out cert.der
openssl x509 -in cert.der -inform DER -outform PEM -out cert.pem

```

## Verifying certificates

```sh
# Test HTTPS connection
openssl s_client -connect example.com:443

# Verifiy certificate chain
openssl verify -CAfile ca.pem cert.pem
```
