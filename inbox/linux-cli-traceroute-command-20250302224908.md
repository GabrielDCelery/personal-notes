---
title: Linux CLI traceroute command
author: GaborZeller
date: 2025-03-02T22-49-08Z
tags:
draft: true
---

# Linux CLI traceroute command

## Common variations and parameters

1. Basic Usage:

```sh
traceroute google.com
```

This shows the path packets take to reach google.com, displaying each hop along the way.

2. Change Protocol:

```sh
# Use TCP instead of UDP (useful when UDP is blocked)
traceroute -T google.com

# Use ICMP
traceroute -I google.com
```

3. Specify Number of Probes:

```sh
# Send 4 probes per hop instead of default 3
traceroute -q 4 google.com
```

4. Set Maximum TTL (hops):

```sh
# Limit trace to 5 hops
traceroute -m 5 google.com
```

5. Change Timeout:

```sh
# Wait 3 seconds for each response
traceroute -w 3 google.com
```

6. Show IP Addresses Only (no DNS resolution):

```sh
traceroute -n google.com
```

7. Specify Source Interface:

```sh
# Useful for multi-homed systems
traceroute -i eth0 google.com
```

8. Set Port Number:

```sh
# Trace to specific destination port
traceroute -p 80 google.com
```

## Common Troubleshooting Scenarios

1. Network Latency Investigation:

```sh
# Show timing for each hop
traceroute -T -n google.com
```

2. Firewall Testing:

```sh
# Try different protocols when one is blocked
traceroute -T -p 80 google.com  # TCP to port 80
traceroute -I google.com        # ICMP
```

3. Route Comparison:

```sh
# Compare paths to different destinations
traceroute site1.com > path1.txt
traceroute site2.com > path2.txt
```

4. MTU Discovery:

```sh
# Find MTU issues along the path
traceroute -F -I google.com
```

Key Parameters Explained:

- `-n`: No DNS resolution (faster, shows IPs only)
- `-T`: Use TCP
- `-I`: Use ICMP
- `-w`: Wait time in seconds
- `-q`: Number of probes per hop
- `-m`: Maximum number of hops
- `-p`: Destination port
- `-F`: Don't fragment packets
- `-i`: Specify source interface
- `-4`: Force IPv4
- `-6`: Force IPv6

These examples cover most common troubleshooting scenarios you might encounter in network diagnostics. The tool is particularly useful for:

- Identifying network bottlenecks
- Verifying network paths
- Debugging connectivity issues
- Testing firewall rules
- Investigating routing problems
- Comparing network paths between different destinations

**Generation complete!** Please review the code suggestions above.

1. Network Troubleshooting:

   - When you're experiencing slow connections to a website/server
   - When you can't connect to a destination at all
   - To identify where exactly in the network path packets are being lost or delayed

2. Network Path Analysis:

   - To understand the route your data takes to reach a destination
   - To verify if traffic is following expected paths (especially useful after network changes)
   - To identify potential bottlenecks in the network

3. ISP Performance Monitoring:

   - To check if your ISP is routing traffic efficiently
   - To verify if you're getting the service quality promised by your ISP
   - To identify if problems are within your ISP's network or beyond

4. Security Analysis:
   - To detect potential security issues like traffic hijacking
   - To verify if traffic is going through expected geographic regions
   - To identify unexpected routing changes that could indicate security problems

Practical Examples:

1. Website Performance Issues:

```sh
traceroute www.example.com
```

Use this when a website is loading slowly to see where the delay is occurring in the network path.

2. Gaming Server Latency:

```sh
traceroute game-server.net
```

Gamers use this to diagnose high ping or connection issues to gaming servers.

3. Corporate Network Troubleshooting:

```sh
traceroute internal-server.company.com
```

System administrators use this to verify corporate network routing and identify internal network issues.

4. Cloud Service Connectivity:

```sh
traceroute aws-region.amazonaws.com
```

DevOps engineers use this to troubleshoot connectivity issues with cloud services.

Remember that traceroute shows you:

- Each network hop (router) between you and the destination
- The time it takes to reach each hop
- Where packets might be getting lost or delayed
- The geographic path your data is taking

This information is invaluable when you need to:

- Prove where a network problem exists
- Communicate issues to your ISP
- Optimize network performance
- Verify network changes
- Document network paths for compliance or documentation purposes
