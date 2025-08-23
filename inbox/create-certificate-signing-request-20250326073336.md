---
title: Create certificate signing request
author: GaborZeller
date: 2025-03-26T07-33-36Z
tags:
draft: true
---

# Create certificate signing request

```sh
#!/bin/bash

# Set domain name
DOMAIN="gabe.co.uk"
# Set organization details
COUNTRY="GB"
STATE="London"
LOCALITY="London"
ORGANIZATION="Gabe Ltd"
ORGANIZATIONAL_UNIT="IT"
EMAIL="admin@gabe.co.uk"

# Create private key with modern 2048-bit encryption
echo "Generating private key..."
openssl genrsa -out "${DOMAIN}.key" 2048

# Create CSR with appropriate subject information
echo "Creating Certificate Signing Request..."
openssl req -new \
    -key "${DOMAIN}.key" \
    -out "${DOMAIN}.csr" \
    -subj "/C=${COUNTRY}/ST=${STATE}/L=${LOCALITY}/O=${ORGANIZATION}/OU=${ORGANIZATIONAL_UNIT}/CN=${DOMAIN}/emailAddress=${EMAIL}"

# Extract public key
echo "Extracting public key..."
openssl rsa -in "${DOMAIN}.key" -pubout -out "${DOMAIN}.pub"

# Output the CSR content for verification
echo -e "\nHere is your CSR content:"
openssl req -in "${DOMAIN}.csr" -noout -text

echo -e "\nFiles created:"
echo "Private key: ${DOMAIN}.key"
echo "Public key: ${DOMAIN}.pub"
echo "CSR: ${DOMAIN}.csr"
echo -e "\nYou can now submit ${DOMAIN}.csr to your chosen Certificate Authority."
```

# Generate self signed certificate

1. Generate private key

openssl genrsa -out server.key 2048

2. Create certificate signing request

openssl req -new -key server.key -out server.csr

3. Generate self-signed certificate

openssl x509 -req -days 365 -in server.csr -signkey server.key -out server.crt
