---
title: Linux CLI dig command
author: GaborZeller
date: 2025-03-02T22-34-35Z
tags:
draft: true
---

# Linux CLI dig command

## Common patterns and parameters

1. Basic DNS Lookup

```sh
dig example.com             # Basic lookup for A records
dig example.com +short      # Short answer only
```

2. Specific Record Types

```sh
dig example.com A          # IPv4 address records
dig example.com AAAA       # IPv6 address records
dig example.com MX         # Mail exchange records
dig example.com NS         # Nameserver records
dig example.com TXT        # TXT records
dig example.com SOA        # Start of Authority
dig example.com ANY        # All record types (may be disabled on some servers)
```

3. Reverse DNS Lookup

```sh
dig -x 8.8.8.8            # Reverse lookup for IP address
```

4. Query Specific DNS Server

```sh
dig @8.8.8.8 example.com  # Query Google's DNS server
dig @1.1.1.1 example.com  # Query Cloudflare's DNS server
```

5. Trace DNS Path

```sh
dig example.com +trace    # Follow delegation path from root
```

6. Common Output Control Parameters

```sh
dig example.com +noall +answer   # Show only the answer section
dig example.com +nocomments      # Remove comment lines
dig example.com +nostats         # Remove statistics
dig example.com +noquestion      # Remove question section
dig example.com +nocmd           # Remove initial comment line
```

7. DNS Lookup with Timing Information

```sh
dig example.com +stats     # Show query statistics
dig example.com +time=1    # Set query timeout to 1 second
```

8. Multiple Queries

```sh
dig example.com google.com    # Query multiple domains
dig example.com NS MX         # Query multiple record types
```

9. DNSSEC Validation

```sh
dig example.com +dnssec    # Request DNSSEC records
dig example.com +cdflag    # Check DNSSEC validation
```

10. TCP Instead of UDP

```sh
dig example.com +tcp       # Force TCP instead of UDP
```

11. Custom Port Query

```sh
dig example.com -p 53      # Specify custom port (default is 53)
```

12. Zone Transfer (if allowed)

```sh
dig @ns1.example.com example.com AXFR    # Attempt zone transfer
```

## Common Troubleshooting Use Cases

1. Check DNS Propagation

```sh
# Check multiple DNS servers to verify propagation
dig @8.8.8.8 example.com
dig @1.1.1.1 example.com
```

2. Debug Email Setup

```sh
# Check all email-related records
dig example.com MX
dig example.com TXT    # For SPF records
dig _dmarc.example.com TXT    # For DMARC records
```

3. Verify CDN Setup

```sh
# Check CNAME records for CDN configuration
dig example.com CNAME
```

4. DNS Response Time Testing

```sh
# Compare response times from different DNS servers
dig @8.8.8.8 example.com +stats
dig @1.1.1.1 example.com +stats
```

5. DNS Troubleshooting

- When websites aren't loading properly
- To verify if DNS records are correctly configured
- To check if DNS changes have propagated
  Example:

```sh
dig example.com +short    # Quick check of IP addresses
```

6. **Email Server Configuration**

- When setting up or troubleshooting email services
- Verifying MX records are correct
- Checking SPF and DMARC records

```sh
dig example.com MX        # Check mail server records
dig example.com TXT       # Check SPF/DMARC records
```

7. **Website Migration**

- When moving a website to a new host
- Verifying DNS propagation across different DNS servers

```sh
dig @8.8.8.8 example.com  # Check Google DNS
dig @1.1.1.1 example.com  # Check Cloudflare DNS
```

8. **CDN Implementation**

- Verifying CDN setup
- Checking if CDN records are properly configured

```sh
dig example.com CNAME     # Check CDN configuration
```

9. **Security Auditing**

- Investigating DNS-based security issues
- Checking for proper DNSSEC implementation

```sh
dig example.com +dnssec   # Check DNSSEC records
```

10. **Performance Testing**

- Measuring DNS response times
- Comparing different DNS providers

```sh
dig example.com +stats    # Check query time and server response
```

11. **Domain Registration Verification**

- Checking domain ownership details
- Verifying nameserver configuration

```sh
dig example.com NS        # Check nameserver records
dig example.com SOA       # Check Start of Authority
```

12. **Reverse DNS Lookup**

- Verifying PTR records
- Troubleshooting email delivery issues (many email servers check reverse DNS)

```sh
dig -x 8.8.8.8           # Look up hostname for an IP
```

13. **Load Balancer Configuration**

- Verifying multiple A records for load balancing
- Checking round-robin DNS setup

```sh
dig example.com A +noall +answer  # Check all A records
```

14. **DNS Infrastructure Debugging**

- Tracing DNS resolution path
- Debugging DNS delegation issues

```sh
dig example.com +trace   # Follow complete DNS resolution path
```

15. **IPv6 Implementation**

- Verifying IPv6 configuration
- Checking dual-stack setup

```sh
dig example.com AAAA    # Check IPv6 records
```

16. **SSL/TLS Certificate Setup**

- Verifying CAA records for certificate authorities
- Checking domain validation records

```sh
dig example.com CAA     # Check Certificate Authority Authorization
```
