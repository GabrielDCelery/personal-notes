---
title: "Networking round 1, question 2"
date: 2025-11-03
tags: ["networking"]
---

# Problem

There is an `FTP` server running at `18.153.44.182` on port `21225` using `STARTTLS`. Get the TLS cert's `CN` (common name).

# Solution

```sh
openssl s_client -connect 18.153.44.182:21225 -starttls ftp -showcerts 2>/dev/null | openssl x509 -noout -subject
```

## Command breakdown

`openssl s_client` - will act as a generic TLS/SSL client like a browser for exampple
`connect` - specifies the host and the port we want to connect to
`starttls` - start in plain text then upgrade to tls
`showcerts` - show all the certs in the call involved
`2>/dev/null` - the call sends a bunch of info to stderr that we don't need (e.g. handshake info, protocol negotiation, connection details)

`openssl x509` - certificate signer and reader utiltiy
`nouout` - don't output the encoded certificate
`subject` - show the subject field (who the cert is belonging to)

## Step by step:

1. Connect to FTP server on port 21225
2. Start plain FTP connection, then upgrade to TLS (-starttls ftp)
3. Hide connection debug messages (2>/dev/null)
4. Parse the certificate received (openssl x509)
5. Show only the subject line (-noout -subject)
6. Extract just the CN value using grep with regex
