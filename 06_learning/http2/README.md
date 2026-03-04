# HTTP/2 and Networking Q&A

This directory contains a comprehensive Q&A session about HTTP/2, TCP, encryption, and related networking concepts.

## Topics Covered

1. [HTTP/2 Reassembly](01-http2-reassembly.md) - How HTTP/2 handles reassembling requests at the application layer
2. [TCP Byte Stream](02-tcp-byte-stream.md) - Understanding TCP's view of data as a continuous byte stream
3. [TCP Connection Establishment](03-tcp-connection-establishment.md) - The three-way handshake and routing
4. [Client SYN/ACK Sequence](04-client-syn-ack-sequence.md) - How the client initiates connections before HTTP flows
5. [Encryption Establishment](05-encryption-establishment.md) - When and how TLS/SSL encryption is established
6. [gRPC Uses HTTP](06-grpc-uses-http.md) - How gRPC leverages HTTP/2 as its transport protocol
7. [Multiple Conversions and Compression](07-multiple-conversions-compression.md) - Understanding layered encoding in gRPC
8. [Plain Text HTTP/2 Encoding](08-plain-text-http2-encoding.md) - What HTTP/2 does with plain text (headers vs body)
9. [CRIME/BREACH Attacks](09-crime-breach-attacks.md) - Why compression + encryption can leak information
10. [gRPC Compression Clarification](10-grpc-compression-clarification.md) - How gRPC's optional compression works
11. [HTTP/2 with gzip Content-Encoding](11-http2-gzip-content-encoding.md) - How application-layer compression works with HTTP/2

## Key Concepts

- **HTTP/2** operates at the application layer and handles frame assembly/reassembly
- **TCP** only sees a continuous byte stream, not application-level boundaries
- **TLS/SSL** encryption happens after TCP but before HTTP data flows
- **gRPC** uses HTTP/2 as transport with Protocol Buffers for serialization
- **HTTP/2 HPACK** only compresses headers, not body content
- **CRIME/BREACH attacks** exploit compression to leak secrets through size observations
- **Content-Encoding** provides application-layer compression separate from HTTP/2 framing

## Network Stack Layers

```
Application:  HTTP/2 / gRPC (framing, multiplexing)
Security:     TLS (encryption)
Transport:    TCP (reliable byte stream)
Network:      IP (routing)
```

Each layer has distinct responsibilities and doesn't need to understand layers above it.
